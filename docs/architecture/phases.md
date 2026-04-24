# 인프라 구축 단계별 로드맵

김용균(플랫폼 아키텍트) 담당 구축 순서. 각 단계는 이전 단계의 결과물에 의존합니다.
Phase 0~7은 WikiPulse 프로젝트와 **동일 베이스라인**으로 시작되었으며 (계획서 명시),
Phase 7.5 이후는 Cledyu (KT Tech-Up Labs) 고유 확장입니다.

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
| 7.5   | kube-prometheus-stack + cert-manager + metrics-server + Strimzi Kafka 최소본 | `gitops/apps/{monitoring,cert-manager,kafka}/` | 1-2 |
| 8     | KEDA + VPA 오토스케일링 (Lab 세션 트리거)  | `gitops/apps/autoscaling/`            | 2     |
| 9     | Velero 백업 + GKE Autopilot DR             | `gitops/apps/velero/`, `infra/gcp/`   | 2     |
| 10    | Terraform AWS → Crossplane 이관            | `infra/terraform/aws/`, `crossplane/` | 2     |
| 11    | Istio Ambient Mesh                         | `gitops/apps/istio/`                  | 2     |
| 12    | **KubeVirt + CDI** (온프렘 Lab VM 풀)       | `gitops/apps/kubevirt/`               | 1-2   |
| 13    | **EC2 Cloud Orchestrator** (AWS 오버플로우) | `infra/terraform/aws/`, `apps/session-api/` | 2 |

## 단계 간 의존성

```
0 ─▶ 1 ─▶ 2 ─▶ 3 ─▶ 4 ─┬─▶ 5 ─▶ 6 ─▶ 7.5 ─▶ 8
                        ├─▶ 7                 └─▶ 9 ─▶ 10 ─▶ 11
                        └─▶ 12 (KubeVirt)
                             └─▶ 13 (EC2 Orchestrator)
```

KubeVirt(12)는 K8s 플랫폼이 서면 Phase 5~8과 병렬 진행 가능합니다.
EC2 Orchestrator(13)는 Session API 백엔드(한정현) 코드와 함께 개발되며,
Phase 10(Crossplane) 이후 Launch Template 관리를 Composition으로 이관합니다.

## 마일스톤 체크리스트

### Month 1 마일스톤 (Phase 0~6, 12)
- [ ] kubeadm HA 3노드 `Ready`
- [ ] Cilium `CiliumHealthy`, Hubble UI 접근 가능
- [ ] MetalLB LB IP 할당 검증(echo-server)
- [ ] Longhorn PVC 3-replica 검증
- [ ] Tailscale MagicDNS로 API 서버 접근 가능
- [ ] ArgoCD 루트 App-of-Apps 동기화 `Healthy`
- [ ] KubeVirt + CDI 설치, `VirtualMachine` CRD 1개 부팅 성공

### Month 2 마일스톤 (Phase 7~11, 13)
- [ ] PR CI: Trivy/CodeQL/ESLint/Ruff/golangci-lint 모두 그린
- [ ] kube-prometheus-stack + Grafana 접근 가능 (cert-manager TLS)
- [ ] Strimzi Kafka `learning-events` 토픽에 테스트 메시지 왕복
- [ ] KEDA Kafka lag 기반 스케일 이벤트 로그 확인
- [ ] EC2 Cloud Orchestrator: Launch Template + cloud-init으로 t3.medium 세션당 생성/destroy 사이클 E2E 검증
- [ ] Velero 복구 드릴 1회 완료 (RPO 1h / RTO 4h)
- [ ] Crossplane으로 AWS 리소스 1개 이상 관리 전환
- [ ] Istio Ambient mTLS STRICT + AuthorizationPolicy 적용
