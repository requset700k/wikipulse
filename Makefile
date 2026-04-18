SHELL := /usr/bin/env bash
.DEFAULT_GOAL := help

# ────── 환경 변수 ──────
HOST        ?=                       # 온프레미스 호스트 IP (예: make host-prep HOST=192.168.0.50)
HOST_USER   ?= ubuntu
ANSIBLE_DIR := ansible
TF_KVM_DIR  := infra/terraform/kvm
TF_KVM_ENV  := infra/terraform/envs/onprem
K8S_DIR     := infra/kubernetes
GITOPS_DIR  := gitops

# ────── 도움말 ──────
help: ## 사용 가능한 명령 목록
	@awk 'BEGIN{FS=":.*##"; printf "\n사용법:\n  make \033[36m<target>\033[0m\n\n"} \
		/^[a-zA-Z_0-9\-]+:.*?##/ {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# ────── Phase 1: 호스트 / VM 프로비저닝 ──────
host-prep: ## 원격 호스트에 libvirt + wpbr0 NAT 네트워크 설치 (HOST=<ip> 필수)
	@test -n "$(HOST)" || (echo "HOST=<ip> 를 지정하세요" && exit 1)
	cd $(ANSIBLE_DIR) && \
		ansible-playbook -i "$(HOST_USER)@$(HOST)," playbooks/00-host-prep.yml

kvm-init: ## Terraform 초기화 (libvirt provider)
	cd $(TF_KVM_ENV) && terraform init

kvm-plan: ## KVM VM 프로비저닝 계획 확인
	cd $(TF_KVM_ENV) && terraform plan

kvm-apply: ## KVM VM 3대 (cp01/cp02/cp03) 프로비저닝
	cd $(TF_KVM_ENV) && terraform apply

kvm-destroy: ## KVM VM 모두 제거 (주의)
	cd $(TF_KVM_ENV) && terraform destroy

# ────── Phase 2: K8s 부트스트랩 ──────
vm-prep: ## VM에 containerd/kubeadm/kubelet 사전 설치
	cd $(ANSIBLE_DIR) && ansible-playbook -i inventory.yml playbooks/10-k8s-nodes.yml

k8s-bootstrap: ## kube-vip + kubeadm HA 컨트롤플레인 부트스트랩
	bash $(K8S_DIR)/kubeadm/bootstrap.sh

# ────── Phase 3~4: 플랫폼 스택 ──────
platform-up: ## Cilium → MetalLB → Longhorn 순차 설치
	bash $(K8S_DIR)/scripts/install-platform.sh

# ────── Phase 6: GitOps ──────
gitops-up: ## ArgoCD 설치 + 루트 App-of-Apps 적용
	bash $(GITOPS_DIR)/argocd/install.sh

.PHONY: help host-prep kvm-init kvm-plan kvm-apply kvm-destroy vm-prep k8s-bootstrap platform-up gitops-up
