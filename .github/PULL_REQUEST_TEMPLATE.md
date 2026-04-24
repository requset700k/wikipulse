<!--
PR 제목 규칙 (Conventional Commits) — 이 PR은 Squash-merge 되므로 **PR 제목만** main 히스토리에 남습니다.
  <type>(<scope>): <subject>
  type:  feat | fix | refactor | perf | docs | test | chore | ci | build | revert | security
  scope: infra | k8s | terraform | ansible | gitops | kubevirt | ec2-orch | kafka | airflow
         | dbt | lab-dsl | validation | ai | rag | api | web | obs | sec | dr | data
  예) feat(kubevirt): CDI DataVolume importer 파이프라인 설정
      feat(ec2-orch): Launch Template cloud-init user-data 표준화
      fix(validation): virtctl ssh 타임아웃 재시도 로직 수정
      security(vault): Gemini API 키 로테이션 정책 적용

main 동기화는 **rebase 권장** (`git fetch origin main && git rebase origin/main && git push --force-with-lease`).
머지 커밋을 브랜치에 만들지 마세요 — Squash 이후에도 중간 히스토리가 지저분해집니다.
-->

## 1. 변경 요약
<!-- 무엇을, 왜 바꿨는지 3줄 이내. "어떻게"는 diff가 설명하므로 생략. -->

## 2. 관련 이슈 / 스프린트
- Closes #
- Refs: <!-- Month/Week, 마일스톤, 기술부채 링크 -->
- Discord 스레드: <!-- 선택 -->

## 3. 변경 유형 (해당 항목 체크)
- [ ] `feat`     기능 추가
- [ ] `fix`      버그 수정
- [ ] `perf`     성능 개선 (SLO 영향 섹션 필수)
- [ ] `refactor` 동작 변경 없는 개선
- [ ] `security` 보안/시크릿/정책
- [ ] `ci`       CI/CD, 릴리스, 품질 게이트
- [ ] `infra`    Terraform / Ansible / K8s / Helm / ArgoCD / KubeVirt
- [ ] `lab`      Lab DSL / Validation Engine / Lab 콘텐츠
- [ ] `ai`       Gemini 프롬프트 / RAG / ChromaDB / Guardrails
- [ ] `data`     Kafka 토픽·스키마 / Airflow / dbt / GE / DataHub
- [ ] `service`  Go/Gin API / Next.js / xterm.js / Kong / Lambda
- [ ] `obs`      메트릭 / 로그 / 트레이스 / 대시보드 / SLI·SLO
- [ ] `docs`     문서만
- [ ] ⚠️ **BREAKING CHANGE** (하위 호환성 파괴 — 아래 "호환성" 섹션 필수)

## 4. 영향 범위 (Layer × 담당자)
> 담당자는 PR 작성 시 Reviewers 필드에 수동 지정합니다.

| 레이어 | 체크 | 컴포넌트 | 담당 |
|---|---|---|---|
| 플랫폼/CI/DR | [ ] | kubeadm · Cilium · MetalLB · Longhorn · **KubeVirt · CDI · EC2 Orchestrator** · ArgoCD · Velero · Crossplane · Istio | 김용균 |
| 보안 | [ ] | Keycloak · Vault · Falco · Kyverno · WAF · GuardDuty · SIEM · **VM 격리(seccomp/AppArmor/SG)** | 윤승호 |
| 관측성 | [ ] | Prometheus · Grafana · Loki · Tempo · OTel · Hubble · SLO | 조승연 |
| Lab + 데이터 | [ ] | **Lab DSL · Validation Engine · virtctl/SSM 추상화** · Strimzi Kafka · Airflow · dbt · BigQuery · DataHub | 김찬영 |
| AI 튜터 | [ ] | **Gemini 3 Pro/Flash · RAG · ChromaDB · sentence-transformers · Guardrails · Context Caching** | 양성호 |
| 서비스 | [ ] | **Go/Gin Session API · VM Orchestrator · Next.js · xterm.js · 강사 모드** · Kong · Redis · Lambda/SES · CloudFront | 한정현 |
| 문서 | [ ] | `docs/` · README · 런북 · ADR · Lab 운영 가이드 | — |

## 5. 데이터 계약 · 스키마 변경
<!-- Kafka(learning-events, validation-requests/results, security-logs) / Lab YAML DSL / dbt 모델 / BigQuery 마트 / API 스키마 / RAG 문서 인덱스 -->
- [ ] 해당 없음
- [ ] Kafka 토픽/스키마 변경 → 컨슈머 호환성 검토 완료
- [ ] Lab YAML DSL 스키마 변경 → validator 통과 + 기존 Lab 4종 회귀 테스트 통과
- [ ] dbt 모델 `ref()` 그래프에 영향 → `dbt build --select state:modified+` 통과
- [ ] OpenAPI/GraphQL 스키마 변경 → 프론트 타입 재생성
- [ ] DataHub 메타데이터 업데이트
- [ ] RAG 문서 인덱스 / ChromaDB collection 스키마 변경

## 6. 보안 체크리스트
- [ ] 시크릿/토큰/API 키 하드코딩 없음 (Vault · GitHub Secrets 사용)
- [ ] 신규 IAM/RBAC/NetworkPolicy 최소권한 원칙 준수
- [ ] 멀티테넌트 VM 격리 4중 방어 유지 (namespace + ResourceQuota + Cilium NetPol + KubeVirt seccomp/AppArmor + Kyverno)
- [ ] Kyverno 정책 위반 없음 (root 차단, hostPath 금지, 리소스 리밋)
- [ ] EC2 오버플로우 VM: SG egress 허용 도메인만, IAM 최소권한, SSM 접근 감사
- [ ] Trivy · ESLint · Ruff · golangci-lint · OWASP ZAP 게이트 통과
- [ ] 외부 의존성 추가 시 라이선스·CVE 검토
- [ ] PII/보안 로그는 `security-logs` 토픽 · S3 Glacier 경로로만

## 7. 관측성 / SLO 영향
- [ ] 신규 메트릭 · 대시보드 · 알림 추가 또는 갱신
- [ ] 로그 구조화(JSON) + `trace_id` 전파
- [ ] 아래 SLO에 영향 없음 또는 영향 분석 첨부
  - Lab 시작 지연 **< 60s (온프렘) / < 90s (EC2)**
  - VM 부팅 성공률 **> 99.5%**
  - WebSocket 끊김률 **< 1%**
  - Validation 응답 **< 10s**
  - AI 힌트 지연 **< 5s**
- [ ] 카오스 · 부하 테스트 필요 여부 판단

## 8. 비용 · DR 영향
- 월 예상 비용 변화: **+$0 / -$0 (변화 없음)** 또는 `+$XX (AWS/GCP/온프렘 내역 명시)`
- [ ] 예산(AWS $710 / GCP $300 + Google AI Pro $30 크레딧) 잔여 마진 내
- [ ] EC2 오버플로우 예상 사용량 반영 (월별 Budget 알람 고려)
- [ ] DR 경로(Velero → GKE Autopilot) 영향 검토, RPO 1h / RTO 4h 유지
- [ ] Gemini 모델 티어링 (Pro → Flash → 정적 힌트) 페일오버 경로 명시
- [ ] Google AI Pro 크레딧 80% 소진 시 자동 Flash 다운그레이드 유지

## 9. 테스트 · 검증
```bash
# 로컬/사전 검증 커맨드 (해당 항목만)
pre-commit run -a
terraform -chdir=infra/terraform/envs/onprem plan
ansible-playbook ansible/playbooks/*.yml --check --diff
kubectl diff -k <kustomize-path>
pytest -q                # Python (AI BFF)
go test ./... && golangci-lint run   # Go (Session API, Validation Engine)
pnpm test && pnpm lint   # Next.js
dbt build --select state:modified+
k6 run load/<scenario>.js
```
- [ ] 단위/통합 테스트 추가 · 갱신
- [ ] `--check` / `plan` / `diff` 결과 첨부 또는 링크
- [ ] 재현 시나리오 또는 스크린샷 포함
- [ ] Lab 콘텐츠 변경 시 실제 Lab 실행 · 통과 확인

## 10. 배포 · 롤백
- 배포 채널: [ ] 자동(ArgoCD sync) [ ] 수동 [ ] 해당 없음
- 순서:
  1.
  2.
- 롤백 방법: <!-- ArgoCD rollback / terraform apply 이전 state / Helm rollback / 이미지 pin 되돌리기 / Lab YAML revert -->
- 마이그레이션 순서 의존성: <!-- DB, 토픽, 인덱스, ChromaDB 리인덱싱 -->

## 11. 호환성 (BREAKING CHANGE 시 필수)
- 사라지거나 바뀐 인터페이스:
- 영향받는 소비자/서비스:
- 마이그레이션 가이드 / 기간:

## 12. 최종 체크리스트
- [ ] 제목이 Conventional Commits 형식
- [ ] Reviewers에 영향 레이어 담당자 1명 이상 지정
- [ ] `pre-commit run -a` / `terraform fmt` / `ansible-lint` / `yamllint` 통과
- [ ] 시크릿 · 크레덴셜 커밋 없음 (`gitleaks` 통과)
- [ ] 문서(`README.md` · `docs/` · 런북) 업데이트
- [ ] 파괴적 운영 작업(노드 drain · destroy · 재부팅 · 시크릿 회전 · Lab VM 대량 destroy) 여부 명시

## 13. 스크린샷 / 로그 / 참고 자료
<!-- Grafana, Kafka UI, ArgoCD, k6 리포트, Gemini 힌트 샘플, Lab 실행 레코딩 등 -->
