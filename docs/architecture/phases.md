# 인프라 구축 단계별 로드맵

김용균(플랫폼 아키텍트) 담당 구축 순서. 각 단계는 이전 단계의 결과물에 의존합니다.

| Phase | 범위                                       | 산출물 위치                           | Month |
| ----- | ------------------------------------------ | ------------------------------------- | ----- |
| 0     | Repo/IaC 구조 스캐폴드                     | `/` (루트 파일)                       | 1     |
| 1     | KVM 3노드 프로비저닝                       | `ansible/`, `infra/terraform/kvm/`    | 1     |
| 2     | kubeadm HA 컨트롤플레인 (kube-vip)         | `infra/kubernetes/kubeadm/`           | 1     |
| 3     | Cilium CNI (kube-proxy replacement)        | `infra/kubernetes/cilium/`            | 1     |
| 4     | MetalLB + Longhorn                         | `infra/kubernetes/{metallb,longhorn}` | 1     |
| 5     | Tailscale subnet router + MagicDNS         | `infra/kubernetes/tailscale/`         | 1     |
| 6     | ArgoCD + GitOps App-of-Apps                | `gitops/`                             | 1     |
| 7     | GitHub Actions CI                          | `.github/workflows/`, `ci/`           | 1-2   |
| 8     | KEDA + VPA 오토스케일링                    | `gitops/apps/autoscaling/`            | 2     |
| 9     | Velero 백업 + GKE Autopilot DR             | `gitops/apps/velero/`, `infra/gcp/`   | 2     |
| 10    | Terraform AWS → Crossplane 이관            | `infra/terraform/aws/`, `crossplane/` | 2     |
| 11    | Istio Ambient Mesh                         | `gitops/apps/istio/`                  | 2     |

## 단계 간 의존성

```
0 ─▶ 1 ─▶ 2 ─▶ 3 ─▶ 4 ─┬─▶ 5 ─▶ 6 ─▶ 8
                        └─▶ 7            └─▶ 9 ─▶ 10 ─▶ 11
```

## 마일스톤 체크리스트

### Month 1 마일스톤 (Phase 0~6)
- [ ] kubeadm HA 3노드 `Ready`
- [ ] Cilium `CiliumHealthy`, Hubble UI 접근 가능
- [ ] MetalLB LB IP 할당 검증(echo-server)
- [ ] Longhorn PVC 3-replica 검증
- [ ] Tailscale MagicDNS로 API 서버 접근 가능
- [ ] ArgoCD 루트 App-of-Apps 동기화 `Healthy`

### Month 2 마일스톤 (Phase 7~11)
- [ ] PR CI: Trivy/CodeQL/ESLint/Ruff 모두 그린
- [ ] KEDA Kafka lag 기반 스케일 이벤트 로그 확인
- [ ] Velero 복구 드릴 1회 완료 (RPO 1h / RTO 4h)
- [ ] Crossplane으로 AWS 리소스 1개 이상 관리 전환
- [ ] Istio Ambient mTLS STRICT + AuthorizationPolicy 적용
