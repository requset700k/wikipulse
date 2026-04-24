# Cledyu 인프라

KT Cloud Tech-Up 교육생을 1차 타겟으로 한 **KodeKloud 스타일 멀티테넌트 VM 기반 Hands-on 교육 플랫폼 + AI 학습 도우미 + 자동 채점 엔진** — 플랫폼/인프라 저장소입니다.
계획서의 김용균(플랫폼 아키텍트) 담당 범위를 코드로 관리합니다.

## 담당 범위
- 온프레미스 쿠버네티스 HA (kubeadm, KVM 3노드, stacked etcd, kube-vip VIP)
- CNI: Cilium (kube-proxy 대체) · LB: MetalLB · 스토리지: Longhorn
- Lab VM 풀: **KubeVirt + CDI (온프렘 검증용) + AWS EC2 오버플로우**
- GitOps: ArgoCD (App-of-Apps) · 서비스 메시: Istio Ambient
- 메시 VPN: Tailscale (MagicDNS, Funnel로 퍼블릭 노출)
- CI/CD: GitHub Actions + Trivy + CodeQL + ESLint + Ruff + golangci-lint
- DR: Velero + GKE Autopilot 대기 사이트 · 오토스케일: KEDA + VPA (Lab 세션 수 기반)
- 클라우드: Terraform(AWS/GCP) → Month 2에 Crossplane Composition으로 이관

## 디렉토리 구조
```
ansible/      호스트 준비(libvirt/bridge) + VM 준비(containerd/kubeadm 사전요건)
infra/
  terraform/  KVM VM 프로비저닝, AWS(EC2 Launch Template 등), GCP DR
  kubernetes/ kubeadm, cilium, metallb, longhorn, kube-vip, kubevirt 매니페스트
gitops/       ArgoCD 설치 + 루트 App-of-Apps
ci/           GitHub Actions 재사용 워크플로우
docs/         아키텍처, 런북, IP 설계, 단계별 로드맵
```

## 빠른 시작 (온프레미스 호스트)
사전 요구사항: Ubuntu 22.04 호스트 (32C/128GB 이상, nested virtualization 권장), SSH 접근, sudo 권한.

```bash
# 1) 호스트에 libvirt 설치 + NAT 네트워크 생성
make host-prep HOST=<호스트-IP>

# 2) Terraform으로 VM 3대 프로비저닝 (libvirt provider)
make kvm-apply

# 3) kubeadm HA 클러스터 부트스트랩
make k8s-bootstrap

# 4) Cilium, MetalLB, Longhorn 설치
make platform-up

# 5) Tailscale subnet router 설치
make tailscale-up AUTHKEY=tskey-auth-...

# 6) ArgoCD 설치 + 루트 App-of-Apps 동기화
make argocd-up
```

자세한 12단계 로드맵은 `docs/architecture/phases.md` 참고.

## 네트워크 설계
`docs/architecture/network.md` 참고.

## 연혁

이 레포는 `wikipulse` 프로젝트에서 pivot되었으며, Phase 12(KubeVirt) 투입 전
운영 식별자(`cluster_name`, libvirt pool, libvirt bridge, Cilium ClusterMesh 이름)를
모두 `cledyu` / `cledyubr0`로 통일했습니다. pivot 배경과 재배포 절차는
`docs/architecture/pivot-to-cledyu.md` 참고.
