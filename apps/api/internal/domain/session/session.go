// Package session은 Lab 세션 도메인 타입을 정의한다.
// 세션 상태 흐름: provisioning → ready → active → completed / failed
// 프론트의 SessionStatus, StepStatus 타입과 1:1 대응.
package session

import "time"

type Status string

const (
	StatusProvisioning Status = "provisioning"
	StatusReady        Status = "ready"
	StatusActive       Status = "active"
	StatusCompleted    Status = "completed"
	StatusFailed       Status = "failed"
)

type VMProvider string

const (
	VMProviderKubeVirt VMProvider = "kubevirt"
	VMProviderEC2      VMProvider = "ec2"
)

type Session struct {
	ID          string     `json:"id"`
	LabID       string     `json:"lab_id"`
	UserID      string     `json:"user_id"`
	Status      Status     `json:"status"`
	VMProvider  VMProvider `json:"vm_provider,omitempty"`
	TerminalURL string     `json:"terminal_url,omitempty"`
	CurrentStep int        `json:"current_step"`
	StartedAt   time.Time  `json:"started_at"`
	ExpiresAt   time.Time  `json:"expires_at"`
}

type StepStatus string

const (
	StepPending StepStatus = "pending"
	StepActive  StepStatus = "active"
	StepPassed  StepStatus = "passed"
	StepFailed  StepStatus = "failed"
)

type StepProgress struct {
	StepID   int        `json:"step_id"`
	Status   StepStatus `json:"status"`
	Attempts int        `json:"attempts"`
}
