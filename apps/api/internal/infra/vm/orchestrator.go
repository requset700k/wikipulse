// Package vm는 VM 오케스트레이션 추상화 레이어를 제공한다.
// Orchestrator 인터페이스로 KubeVirt/EC2/Stub을 동일하게 다룬다.
// StubOrchestrator: 로컬 개발 전용, 5초 후 Port=0 VMInfo 반환 → 터미널 핸들러가 로컬 bash 실행.
package vm

import (
	"context"
	"time"
)

type Provider string

const (
	ProviderKubeVirt Provider = "kubevirt"
	ProviderEC2      Provider = "ec2"
	ProviderStub     Provider = "stub"
)

type CreateRequest struct {
	SessionID string
	UserID    string
	LabID     string
	VMType    string // small | medium
}

type VMInfo struct {
	ID       string
	Provider Provider
	IP       string
	Port     int
}

type VMStatus struct {
	Ready bool
	Phase string
}

// Orchestrator abstracts VM lifecycle across KubeVirt and EC2.
type Orchestrator interface {
	Create(ctx context.Context, req CreateRequest) (*VMInfo, error)
	Delete(ctx context.Context, vmID string, provider Provider) error
	Status(ctx context.Context, vmID string, provider Provider) (*VMStatus, error)
}

// StubOrchestrator simulates a 5-second VM provisioning for local development.
type StubOrchestrator struct{}

func (s *StubOrchestrator) Create(_ context.Context, req CreateRequest) (*VMInfo, error) {
	time.Sleep(5 * time.Second)
	return &VMInfo{
		ID:       "stub-" + req.SessionID[:8],
		Provider: ProviderStub,
		IP:       "127.0.0.1",
		Port:     0, // terminal handler uses local shell when Port == 0
	}, nil
}

func (s *StubOrchestrator) Delete(_ context.Context, _ string, _ Provider) error {
	return nil
}

func (s *StubOrchestrator) Status(_ context.Context, _ string, _ Provider) (*VMStatus, error) {
	return &VMStatus{Ready: true, Phase: "Running"}, nil
}
