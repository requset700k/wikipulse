# WikiPulse 인프라

Wikipedia 편집 폭증 기반 실시간 이슈 감지 + AI 여론 분석 플랫폼의 **플랫폼/인프라 저장소**입니다.
계획서의 김용균(플랫폼 아키텍트) 담당 범위를 코드로 관리합니다.

## 담당 범위
- 온프레미스 쿠버네티스 HA (kubeadm, KVM 3노드, stacked etcd)
- CNI: Cilium (kube-proxy 대체) · LB: MetalLB · 스토리지: Longhorn
- GitOps: ArgoCD (App-of-Apps) · 서비스 메시: Istio Ambient
- 메시 VPN: Tailscale (MagicDNS 우선, 퍼블릭 도메인은 추후 TODO)
- CI/CD: GitHub Actions + Trivy + CodeQL + ESLint + Ruff
- DR: Velero + GKE Autopilot 대기 사이트 · 오토스케일: KEDA + VPA
- 클라우드: Terraform(AWS/GCP) → Month 2에 Crossplane Composition으로 이관

## 디렉토리 구조
```
ansible/      호스트 준비(libvirt/bridge) + VM 준비(containerd/kubeadm 사전요건)
infra/
  terraform/  KVM VM 프로비저닝, AWS, GCP DR
  kubernetes/ kubeadm, cilium, metallb, longhorn, kube-vip 매니페스트
gitops/       ArgoCD 설치 + 루트 App-of-Apps
ci/           GitHub Actions 재사용 워크플로우
docs/         아키텍처, 런북, IP 설계, 단계별 로드맵
```

## 빠른 시작 (온프레미스 호스트)
사전 요구사항: Ubuntu 22.04 호스트 (32C/128GB 이상), SSH 접근, sudo 권한.

```bash
# 1) 호스트에 libvirt 설치 + wpbr0 NAT 네트워크 생성
make host-prep HOST=<호스트-IP>

# 2) Terraform으로 VM 3대 프로비저닝 (libvirt provider)
make kvm-apply

# 3) kubeadm HA 클러스터 부트스트랩
make k8s-bootstrap

# 4) Cilium, MetalLB, Longhorn 설치
make platform-up

# 5) ArgoCD 설치 + 루트 App-of-Apps 동기화
make gitops-up
```

자세한 12단계 로드맵은 `docs/architecture/phases.md` 참고.

## 네트워크 설계
`docs/architecture/network.md` 참고.
