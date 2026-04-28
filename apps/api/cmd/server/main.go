// 서버 진입점 — 설정 로드, 인프라 초기화, HTTP 서버 시작/종료를 순서대로 처리한다.
// SIGINT/SIGTERM 수신 시 진행 중인 요청을 최대 10초 동안 기다린 후 graceful shutdown.
package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	api "github.com/kt-techuplabs/cledyu/backend/internal/api"
	"github.com/kt-techuplabs/cledyu/backend/internal/config"
	infradb "github.com/kt-techuplabs/cledyu/backend/internal/infra/db"
	"github.com/kt-techuplabs/cledyu/backend/internal/infra/vm"
	"github.com/kt-techuplabs/cledyu/backend/internal/service"
	"go.uber.org/zap"

	infracache "github.com/kt-techuplabs/cledyu/backend/internal/infra/cache"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	var logger *zap.Logger
	if cfg.Server.Mode == "release" {
		logger, _ = zap.NewProduction()
	} else {
		logger, _ = zap.NewDevelopment()
	}
	defer logger.Sync() //nolint:errcheck

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// ── Database ────────────────────────────────────────────────────────────
	pool, err := infradb.Connect(ctx, cfg.Database.URL)
	if err != nil {
		logger.Fatal("db connect", zap.Error(err))
	}
	defer pool.Close()

	if err := infradb.MigrateUp(ctx, pool); err != nil {
		logger.Fatal("db migrate", zap.Error(err))
	}
	logger.Info("database ready")

	// ── Redis (optional) ────────────────────────────────────────────────────
	rdb, err := infracache.Connect(ctx, cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		logger.Warn("redis unavailable — hint rate limiting disabled", zap.Error(err))
		rdb = nil
	} else {
		logger.Info("redis ready")
	}

	// ── VM Orchestrator ─────────────────────────────────────────────────────
	var orch vm.Orchestrator
	switch cfg.VM.Provider {
	case "kubevirt":
		orch = vm.NewKubeVirt(cfg.VM.KubeAPIServer, cfg.VM.KubeVirtNS, cfg.VM.KubeToken)
		logger.Info("vm orchestrator: kubevirt")
	default:
		orch = &vm.StubOrchestrator{}
		logger.Info("vm orchestrator: stub (dev mode)")
	}

	// ── Services ────────────────────────────────────────────────────────────
	sessions := service.NewSessionService(pool, orch, logger)
	hints := service.NewHintService(rdb, cfg.AI.BFFURL, logger)

	// ── HTTP Server ─────────────────────────────────────────────────────────
	router := api.NewRouter(cfg, logger, sessions, hints)

	srv := &http.Server{
		Addr:         cfg.Server.Addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 0, // WebSocket connections must not time out
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("server started", zap.String("addr", cfg.Server.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown error", zap.Error(err))
	}
}
