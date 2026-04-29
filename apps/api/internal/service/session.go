// Package service는 핵심 비즈니스 로직을 담당한다.
// 핸들러(HTTP 계층)와 인프라(DB/VM/Redis) 사이의 중간 계층.
package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kt-techuplabs/cledyu/backend/internal/domain/session"
	"github.com/kt-techuplabs/cledyu/backend/internal/infra/vm"
	"go.uber.org/zap"
)

type SessionService struct {
	db  *pgxpool.Pool
	vm  vm.Orchestrator
	log *zap.Logger
}

func NewSessionService(db *pgxpool.Pool, orch vm.Orchestrator, log *zap.Logger) *SessionService {
	return &SessionService{db: db, vm: orch, log: log}
}

type CreateParams struct {
	UserID    string
	LabID     string
	VMType    string
	StepCount int
}

// Create는 새 Lab 세션을 생성하고 VM 프로비저닝을 비동기로 시작한다.
// 즉시 status=provisioning 상태의 세션을 반환하고, 프론트는 GET /sessions/:id 폴링으로 ready를 기다린다.
func (s *SessionService) Create(ctx context.Context, p CreateParams) (*session.Session, error) {
	if err := s.ensureUser(ctx, p.UserID); err != nil {
		return nil, fmt.Errorf("ensure user: %w", err)
	}

	// 동일 Lab에 이미 활성 세션이 있으면 재사용 (중복 bash 프로세스 방지)
	// started_at을 현재 시간으로 갱신해 강사 대시보드 필터에 포함되도록 함
	var existingID string
	err := s.db.QueryRow(ctx, `
		SELECT se.id FROM sessions se
		JOIN users u ON u.id = se.user_id
		WHERE u.keycloak_id = $1 AND se.lab_id = $2
		  AND se.status IN ('provisioning','ready','active')
		LIMIT 1
	`, p.UserID, p.LabID).Scan(&existingID)
	if err == nil && existingID != "" {
		if _, execErr := s.db.Exec(ctx, `
			UPDATE sessions SET started_at = NOW(), expires_at = NOW() + INTERVAL '3 hours'
			WHERE id = $1
		`, existingID); execErr != nil {
			s.log.Warn("session renew", zap.Error(execErr))
		}
		return s.Get(ctx, existingID, p.UserID)
	}

	id := newID()
	now := time.Now().UTC()
	expires := now.Add(3 * time.Hour)

	_, execErr := s.db.Exec(ctx, `
		INSERT INTO sessions (id, lab_id, user_id, status, started_at, expires_at)
		VALUES ($1, $2, (SELECT id FROM users WHERE keycloak_id = $3), 'provisioning', $4, $5)
	`, id, p.LabID, p.UserID, now, expires)
	if execErr != nil {
		return nil, fmt.Errorf("insert session: %w", execErr)
	}

	if err := s.initSteps(ctx, id, p.StepCount); err != nil {
		s.log.Warn("init steps failed", zap.Error(err))
	}

	go s.provisionAsync(id, p) //nolint:gosec

	return &session.Session{
		ID:          id,
		LabID:       p.LabID,
		UserID:      p.UserID,
		Status:      session.StatusProvisioning,
		CurrentStep: 0,
		StartedAt:   now,
		ExpiresAt:   expires,
	}, nil
}

func (s *SessionService) Get(ctx context.Context, id, userID string) (*session.Session, error) {
	row := s.db.QueryRow(ctx, `
		SELECT se.id, se.lab_id, u.keycloak_id, se.status,
		       COALESCE(se.vm_provider,''), COALESCE(se.vm_ip,''), se.vm_port,
		       se.current_step, se.started_at, se.expires_at
		FROM sessions se
		JOIN users u ON u.id = se.user_id
		WHERE se.id = $1 AND u.keycloak_id = $2
	`, id, userID)

	var sess session.Session
	var providerStr, vmIP string
	var vmPort int
	err := row.Scan(
		&sess.ID, &sess.LabID, &sess.UserID, &sess.Status,
		&providerStr, &vmIP, &vmPort,
		&sess.CurrentStep, &sess.StartedAt, &sess.ExpiresAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}
	if sess.Status == session.StatusReady {
		sess.TerminalURL = fmt.Sprintf("/api/v1/sessions/%s/ws", sess.ID)
	}
	sess.VMProvider = session.VMProvider(providerStr)
	return &sess, nil
}

func (s *SessionService) Delete(ctx context.Context, id, userID string) error {
	var vmID, providerStr string
	err := s.db.QueryRow(ctx, `
		SELECT COALESCE(se.vm_id,''), COALESCE(se.vm_provider,'')
		FROM sessions se JOIN users u ON u.id = se.user_id
		WHERE se.id = $1 AND u.keycloak_id = $2
	`, id, userID).Scan(&vmID, &providerStr)
	if err != nil {
		return fmt.Errorf("find session: %w", err)
	}
	if vmID != "" {
		go s.vm.Delete(context.Background(), vmID, vm.Provider(providerStr)) //nolint:errcheck,gosec
	}
	_, err = s.db.Exec(ctx,
		"UPDATE sessions SET status='completed', completed_at=NOW() WHERE id=$1", id)
	return err
}

func (s *SessionService) VMInfo(ctx context.Context, id, userID string) (ip string, port int, provider vm.Provider, err error) {
	row := s.db.QueryRow(ctx, `
		SELECT COALESCE(se.vm_ip,''), se.vm_port, COALESCE(se.vm_provider,''), se.status
		FROM sessions se JOIN users u ON u.id = se.user_id
		WHERE se.id = $1 AND u.keycloak_id = $2
	`, id, userID)
	var status string
	if err = row.Scan(&ip, &port, &provider, &status); err != nil {
		return
	}
	if status != string(session.StatusReady) {
		err = fmt.Errorf("session not ready (status=%s)", status)
	}
	return
}

// ── Step progress ──────────────────────────────────────────────────────────

func (s *SessionService) GetSteps(ctx context.Context, sessionID, userID string) ([]session.StepProgress, error) {
	if err := s.assertOwner(ctx, sessionID, userID); err != nil {
		return nil, err
	}
	rows, err := s.db.Query(ctx,
		"SELECT step_id, status, attempts FROM step_progress WHERE session_id=$1 ORDER BY step_id",
		sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	steps := make([]session.StepProgress, 0)
	for rows.Next() {
		var sp session.StepProgress
		var statusStr string
		if err := rows.Scan(&sp.StepID, &statusStr, &sp.Attempts); err != nil {
			return nil, err
		}
		sp.Status = session.StepStatus(statusStr)
		steps = append(steps, sp)
	}
	return steps, nil
}

func (s *SessionService) UpdateStep(ctx context.Context, sessionID, userID string, stepID int, status session.StepStatus) error {
	if err := s.assertOwner(ctx, sessionID, userID); err != nil {
		return err
	}
	_, err := s.db.Exec(ctx, `
		UPDATE step_progress
		SET status=$3, attempts = attempts + 1, updated_at = NOW()
		WHERE session_id=$1 AND step_id=$2
	`, sessionID, stepID, string(status))
	return err
}

// SimulateValidation marks a step as passed after 2s — used when Kafka is not set up.
func (s *SessionService) SimulateValidation(sessionID string, stepID int) {
	go func() {
		time.Sleep(2 * time.Second)
		ctx := context.Background()
		if _, err := s.db.Exec(ctx, `
			UPDATE step_progress
			SET status='passed', attempts = attempts + 1, updated_at = NOW()
			WHERE session_id=$1 AND step_id=$2
		`, sessionID, stepID); err != nil {
			s.log.Warn("simulate validation step", zap.Error(err))
		}

		// Advance current_step
		if _, err := s.db.Exec(ctx,
			"UPDATE sessions SET current_step = current_step + 1 WHERE id=$1",
			sessionID); err != nil {
			s.log.Warn("simulate validation step advance", zap.Error(err))
		}
	}()
}

// ActiveSessions returns sessions visible to instructors.
func (s *SessionService) ActiveSessions(ctx context.Context) ([]InstructorSession, error) {
	rows, err := s.db.Query(ctx, `
		SELECT se.id, u.name, u.email, se.lab_id, se.status, se.current_step, se.started_at,
		       COUNT(sp.id) FILTER (WHERE sp.status = 'passed') AS steps_done
		FROM sessions se
		JOIN users u ON u.id = se.user_id
		LEFT JOIN step_progress sp ON sp.session_id = se.id
		WHERE se.status IN ('provisioning','ready','active')
		  AND se.started_at > NOW() - INTERVAL '3 hours'
		GROUP BY se.id, u.name, u.email
		ORDER BY se.started_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []InstructorSession
	for rows.Next() {
		var is InstructorSession
		if err := rows.Scan(&is.ID, &is.UserName, &is.UserEmail, &is.LabID,
			&is.Status, &is.CurrentStep, &is.StartedAt, &is.StepsDone); err != nil {
			return nil, err
		}
		list = append(list, is)
	}
	return list, nil
}

type InstructorSession struct {
	ID          string    `json:"id"`
	UserName    string    `json:"user_name"`
	UserEmail   string    `json:"user_email"`
	LabID       string    `json:"lab_id"`
	Status      string    `json:"status"`
	CurrentStep int       `json:"current_step"`
	StepsDone   int       `json:"steps_done"`
	StartedAt   time.Time `json:"started_at"`
}

// ── internal helpers ──────────────────────────────────────────────────────

func (s *SessionService) initSteps(ctx context.Context, sessionID string, stepCount int) error {
	batch := &pgx.Batch{}
	for i := 1; i <= stepCount; i++ {
		batch.Queue(
			"INSERT INTO step_progress (session_id, step_id, status) VALUES ($1, $2, 'pending') ON CONFLICT DO NOTHING",
			sessionID, i,
		)
	}
	return s.db.SendBatch(ctx, batch).Close()
}

func (s *SessionService) provisionAsync(sessionID string, p CreateParams) {
	ctx := context.Background()
	vmInfo, err := s.vm.Create(ctx, vm.CreateRequest{
		SessionID: sessionID,
		UserID:    p.UserID,
		LabID:     p.LabID,
		VMType:    p.VMType,
	})
	if err != nil {
		s.log.Error("vm provision failed", zap.String("session", sessionID), zap.Error(err))
		s.setStatus(ctx, sessionID, "failed")
		return
	}
	if _, err := s.db.Exec(ctx, `
		UPDATE sessions SET status='ready', vm_id=$2, vm_ip=$3, vm_port=$4, vm_provider=$5 WHERE id=$1
	`, sessionID, vmInfo.ID, vmInfo.IP, vmInfo.Port, string(vmInfo.Provider)); err != nil {
		s.log.Error("provision update session", zap.String("session", sessionID), zap.Error(err))
	}
}

func (s *SessionService) setStatus(ctx context.Context, id, status string) {
	s.db.Exec(ctx, "UPDATE sessions SET status=$2 WHERE id=$1", id, status) //nolint:errcheck
}

func (s *SessionService) ensureUser(ctx context.Context, keycloakID string) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO users (keycloak_id, email, name, role)
		VALUES ($1, $1 || '@dev.local', 'Dev User', 'student')
		ON CONFLICT (keycloak_id) DO NOTHING
	`, keycloakID)
	return err
}

func (s *SessionService) assertOwner(ctx context.Context, sessionID, userID string) error {
	var exists bool
	err := s.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM sessions se JOIN users u ON u.id = se.user_id
			WHERE se.id=$1 AND u.keycloak_id=$2
		)`, sessionID, userID).Scan(&exists)
	if err != nil || !exists {
		return fmt.Errorf("session not found")
	}
	return nil
}

func newID() string {
	b := make([]byte, 16)
	rand.Read(b) //nolint:errcheck
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
