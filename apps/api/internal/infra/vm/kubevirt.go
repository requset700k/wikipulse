// KubeVirtOrchestrator: K8s API를 직접 호출해 VirtualMachineInstance CRD를 생성/삭제한다.
// 온프렘 검증용 VM 풀 담당. VM 타입: small(2vCPU/4GB), medium(4vCPU/8GB).
package vm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// KubeVirtOrchestrator provisions VMs via KubeVirt CRD (VirtualMachineInstance).
// Requires: kubeadm cluster + KubeVirt installed + in-cluster or kubeconfig auth.
type KubeVirtOrchestrator struct {
	apiServer  string
	namespace  string
	token      string // ServiceAccount bearer token
	httpClient *http.Client
}

func NewKubeVirt(apiServer, namespace, token string) *KubeVirtOrchestrator {
	return &KubeVirtOrchestrator{
		apiServer:  apiServer,
		namespace:  namespace,
		token:      token,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (k *KubeVirtOrchestrator) Create(ctx context.Context, req CreateRequest) (*VMInfo, error) {
	cpu, mem := vmResources(req.VMType)
	vmiName := "lab-" + req.SessionID[:8]

	manifest := fmt.Sprintf(`{
		"apiVersion": "kubevirt.io/v1",
		"kind": "VirtualMachineInstance",
		"metadata": {
			"name": %q,
			"namespace": %q,
			"labels": {"session-id": %q, "lab-id": %q}
		},
		"spec": {
			"domain": {
				"cpu": {"cores": %d},
				"resources": {"requests": {"memory": %q}},
				"devices": {
					"disks": [{"name": "rootdisk", "disk": {"bus": "virtio"}}],
					"interfaces": [{"name": "default", "masquerade": {}}]
				}
			},
			"networks": [{"name": "default", "pod": {}}],
			"volumes": [{
				"name": "rootdisk",
				"containerDisk": {"image": "quay.io/containerdisks/ubuntu:22.04"}
			}]
		}
	}`, vmiName, k.namespace, req.SessionID, req.LabID, cpu, mem)

	url := fmt.Sprintf("%s/apis/kubevirt.io/v1/namespaces/%s/virtualmachineinstances", k.apiServer, k.namespace)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(manifest))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+k.token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := k.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("create VMI: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("create VMI: HTTP %d", resp.StatusCode)
	}

	return &VMInfo{
		ID:       vmiName,
		Provider: ProviderKubeVirt,
		Port:     22,
	}, nil
}

func (k *KubeVirtOrchestrator) Status(ctx context.Context, vmID string, _ Provider) (*VMStatus, error) {
	url := fmt.Sprintf("%s/apis/kubevirt.io/v1/namespaces/%s/virtualmachineinstances/%s",
		k.apiServer, k.namespace, vmID)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	httpReq.Header.Set("Authorization", "Bearer "+k.token)

	resp, err := k.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck

	var vmi struct {
		Status struct {
			Phase      string `json:"phase"`
			Interfaces []struct {
				IP string `json:"ipAddress"`
			} `json:"interfaces"`
		} `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&vmi); err != nil {
		return nil, err
	}

	ready := vmi.Status.Phase == "Running"
	return &VMStatus{Ready: ready, Phase: vmi.Status.Phase}, nil
}

func (k *KubeVirtOrchestrator) Delete(ctx context.Context, vmID string, _ Provider) error {
	url := fmt.Sprintf("%s/apis/kubevirt.io/v1/namespaces/%s/virtualmachineinstances/%s",
		k.apiServer, k.namespace, vmID)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	httpReq.Header.Set("Authorization", "Bearer "+k.token)

	resp, err := k.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck
	return nil
}

func vmResources(vmType string) (cpu int, mem string) {
	if vmType == "medium" {
		return 4, "8Gi"
	}
	return 2, "4Gi" // small (default)
}
