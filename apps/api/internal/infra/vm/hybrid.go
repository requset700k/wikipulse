// HybridOrchestrator: 온프렘 KubeVirt 우선 사용, 풀 소진 시 AWS EC2 폴백.
// 기획서의 "온프렘 우선 + EC2 오버플로우" 라우팅 정책 구현.
package vm

import (
	"context"
	"fmt"
)

// HybridOrchestrator는 온프렘 KubeVirt를 우선 사용하고 풀이 소진되면 AWS EC2로 폴백한다.
// 기획서의 "온프렘 우선 + EC2 오버플로우" 라우팅 정책을 구현.
type HybridOrchestrator struct {
	kubevirt       Orchestrator
	ec2            Orchestrator
	maxKubeVirtVMs int // pool capacity threshold
}

func NewHybrid(kubevirt, ec2 Orchestrator, maxKubeVirtVMs int) *HybridOrchestrator {
	return &HybridOrchestrator{
		kubevirt:       kubevirt,
		ec2:            ec2,
		maxKubeVirtVMs: maxKubeVirtVMs,
	}
}

func (h *HybridOrchestrator) Create(ctx context.Context, req CreateRequest) (*VMInfo, error) {
	// TODO: query KubeVirt pool usage to decide routing.
	// For now, always prefer KubeVirt.
	info, err := h.kubevirt.Create(ctx, req)
	if err != nil {
		// Fallback to EC2
		info, err = h.ec2.Create(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("both KubeVirt and EC2 failed: %w", err)
		}
	}
	return info, nil
}

func (h *HybridOrchestrator) Delete(ctx context.Context, vmID string, provider Provider) error {
	switch provider {
	case ProviderKubeVirt:
		return h.kubevirt.Delete(ctx, vmID, provider)
	case ProviderEC2:
		return h.ec2.Delete(ctx, vmID, provider)
	default:
		return fmt.Errorf("unknown provider: %s", provider)
	}
}

func (h *HybridOrchestrator) Status(ctx context.Context, vmID string, provider Provider) (*VMStatus, error) {
	switch provider {
	case ProviderKubeVirt:
		return h.kubevirt.Status(ctx, vmID, provider)
	case ProviderEC2:
		return h.ec2.Status(ctx, vmID, provider)
	default:
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}
}
