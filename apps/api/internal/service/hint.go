// AI 힌트 서비스 — 양성호 담당 AI BFF(Python FastAPI)에 힌트를 요청한다.
// BFF가 없거나 장애 시 mockHint로 폴백.
//
// Rate Limit: Redis INCR+Expire 패턴으로 사용자×세션 단위 분당 6회 제한.
//
//	Redis 장애 시에는 제한 없이 허용 (서비스 가용성 우선).
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var ErrRateLimitExceeded = errors.New("rate limit exceeded")

type HintParams struct {
	UserID          string
	SessionID       string
	StepID          int
	HintLevel       int
	TerminalHistory string
}

type HintResult struct {
	HintText       string   `json:"hint_text"`
	RelatedDocs    []string `json:"related_docs"`
	HintsRemaining int      `json:"hints_remaining"`
}

type HintService struct {
	rdb    *redis.Client
	bffURL string
	log    *zap.Logger
}

func NewHintService(rdb *redis.Client, bffURL string, log *zap.Logger) *HintService {
	return &HintService{rdb: rdb, bffURL: bffURL, log: log}
}

const hintLimitPerMinute = 6

func (s *HintService) Request(ctx context.Context, p HintParams) (*HintResult, error) {
	remaining, err := s.checkRateLimit(ctx, p.UserID, p.SessionID)
	if err != nil {
		return nil, err
	}

	var result *HintResult
	if s.bffURL != "" {
		result, err = s.callBFF(ctx, p)
		if err != nil {
			s.log.Warn("ai bff failed, falling back to mock", zap.Error(err))
			result = mockHint(p.HintLevel, remaining)
		}
	} else {
		result = mockHint(p.HintLevel, remaining)
	}

	result.HintsRemaining = remaining
	return result, nil
}

// checkRateLimit은 Redis INCR로 카운터를 증가시키고 남은 횟수를 반환한다.
// 첫 요청(count==1)일 때만 Expire를 설정해 1분 윈도우를 만든다.
// INCR과 Expire가 분리된 이유: INCR+Expire를 원자적으로 처리하는 SET NX EX를 쓰지 않는 것은
// 첫 요청에서만 TTL을 붙이는 "슬라이딩 윈도우 없는 고정 윈도우" 방식이기 때문.
func (s *HintService) checkRateLimit(ctx context.Context, userID, sessionID string) (int, error) {
	if s.rdb == nil {
		return hintLimitPerMinute, nil
	}
	key := fmt.Sprintf("hint:%s:%s", userID, sessionID)
	count, err := s.rdb.Incr(ctx, key).Result()
	if err != nil {
		return hintLimitPerMinute, nil //nolint:nilerr // redis 장애 시 허용
	}
	if count == 1 {
		s.rdb.Expire(ctx, key, time.Minute) //nolint:errcheck
	}
	remaining := hintLimitPerMinute - int(count)
	if remaining < 0 {
		return 0, ErrRateLimitExceeded
	}
	return remaining, nil
}

func (s *HintService) callBFF(ctx context.Context, p HintParams) (*HintResult, error) {
	body, _ := json.Marshal(map[string]any{
		"session_id":       p.SessionID,
		"step_id":          p.StepID,
		"hint_level":       p.HintLevel,
		"terminal_history": p.TerminalHistory,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.bffURL+"/hint", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck

	var result HintResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// mockHint returns a Socratic-style hint based on level (no AI BFF required).
var mockHints = map[int]string{
	1: "어떤 명령어로 현재 상태를 확인할 수 있을지 생각해보세요. man 페이지나 --help 옵션이 도움이 될 수 있어요.",
	2: "지금 작업하는 디렉토리와 목표 파일의 위치를 다시 확인해보세요. 상대 경로와 절대 경로의 차이를 생각해보세요.",
	3: "이전 단계에서 사용한 명령어를 다시 검토해보세요. 옵션 하나가 결과를 크게 바꿀 수 있어요.",
}

func mockHint(level, remaining int) *HintResult {
	hint, ok := mockHints[level]
	if !ok {
		hint = mockHints[1]
	}
	return &HintResult{
		HintText:       hint,
		RelatedDocs:    []string{"https://man7.org/linux/man-pages/"},
		HintsRemaining: remaining,
	}
}
