# ADR: Keycloak 권한 모델 (RBAC)

- **상태:** Proposed
- **날짜:** 2026-04-26
- **제안자:** 김용균 / Platform
- **의사결정자:** 김용균 / 윤승호
- **관련:** Keycloak Ingress + TLS, Keycloak Client Trust Setup

## 1. 컨텍스트

`https://keycloak.cledyu.local` 가 Traefik + Cledyu 내부 TLS 로 외부 노출 완료된 시점부터, default bootstrap admin 만으로 운영하는 것은 그 자체로 보안 부채가 된다.

본 ADR 은 application 레벨 Keycloak 권한 모델 — realm / client / role / group / 팀원 접근 — 을 Terraform 으로 선언적으로 구현 가능하도록 합의한다.

## 2. 문제 / 목표

목표:

- default bootstrap admin 으로 운영되는 윈도우 최소화
- 6명 팀원의 최소권한 원칙 적용
- `web`, `api`, `tutor`, `argocd`, `grafana` 의 OIDC 입력값 정의
- 강사 / 학습자 lifecycle 의 MVP 정의
- 인증 / admin audit 이벤트를 `security-logs` 파이프라인에 정렬

본 ADR 의 비-목표:

- 외부 학습자용 별도 realm 분리
- Google Workspace / SAML IdP 연동
- Keycloak audit → Kafka 연동 (Strimzi 의존, 후속 PR)

## 3. 고려한 대안

### 옵션 A: `cledyu` 단일 realm

- 장점: 단순, 운영 부담 작음, 6인 내부 프로젝트에 적합
- 단점: 외부 학습자 분리 필요 시 별도 ADR 필요

### 옵션 B: `master` + `cledyu` 분리, admin 격리 강화

- 장점: 플랫폼 운영자 / 일반 사용자 격리 강함
- 단점: MVP 단계에 과한 운영 복잡도

### 옵션 C: `cledyu-internal` + `cledyu-external` 분리

- 장점: 외부 학습자 / 기업 고객 장기 격리에 최적
- 단점: 3개월 내부 프로젝트에 시기상조

## 4. 결정

**옵션 A 채택.**

- `master` realm 은 Keycloak super-admin 전용으로 보존
- 비즈니스 사용자 / 서비스 / 학습자 / 강사 / 팀 그룹은 모두 `cledyu` realm 에 둔다
- 구현은 Terraform `mrparkers/keycloak` provider 사용

## 5. Realm

| Realm | 목적 |
|---|---|
| `cledyu` | 모든 Cledyu 서비스 사용자, 팀원, 학습자, 강사, role, OIDC client |

Realm 정책:

- self-registration 비활성
- remember-me 비활성
- offline token MVP 단계는 비활성
- 첫 로그인 시 비밀번호 변경 강제 (managed user)

## 6. Client

| Client | 서비스 | Type | Flow | 비고 |
|---|---|---|---|---|
| `web` | Next.js | public | Authorization Code + PKCE | SPA 로그인 |
| `api` | Go/Gin | bearer-only | JWT 검증만 | 직접 로그인 X |
| `tutor` | FastAPI | bearer-only | JWT 검증만 | 직접 로그인 X |
| `argocd` | ArgoCD | confidential | Authorization Code | group claim 매핑 사용 |
| `grafana` | Grafana | confidential | Authorization Code | role / group 매핑 사용 |

## 7. Realm Role

| Role | 의미 | 주요 부여 대상 |
|---|---|---|
| `student` | Lab 접근 + AI 튜터 사용 | 학습자 |
| `instructor` | 학습자 계정 생성 + Lab 변형 + 강사 모드 | 강사 |
| `admin` | 운영 관리 | 팀 도메인 owner |
| `observer` | 대시보드 / 로그 read-only | 도메인 외 팀원 접근 |

Client role 은 MVP 에서는 도입하지 않는다. ArgoCD project / Grafana folder / 서비스 별 scope 세분화가 필요해지면 후속 ADR 로 client role 추가.

## 8. Group

| Group | Realm role 매핑 | 멤버 |
|---|---|---|
| `team-platform` | `admin` | 김용균 |
| `team-security` | `admin` | 윤승호 |
| `team-observability` | `admin`, `observer` | 조승연 |
| `team-lab-data` | `admin`, `observer` | 김찬영 |
| `team-ai` | `admin`, `observer` | 양성호 |
| `team-service` | `admin`, `observer` | 한정현 |
| `students-cohort-0` | `student` | 강사가 초대한 학습자 |
| `instructors` | `instructor` | 강사 |

도메인 별 권한은 downstream application 이 group claim 으로 처리. Keycloak 자체는 group / role claim 만 발행.

## 9. 팀 권한 매트릭스

| 이름 | 팀 | Keycloak | ArgoCD | Grafana | Web/API/Tutor |
|---|---|---|---|---|---|
| 김용균 | platform | super-admin | admin | admin | admin |
| 윤승호 | security | admin | reader | reader | observer |
| 조승연 | observability | reader | reader | admin | observer |
| 김찬영 | lab-data | reader | reader | reader | observer |
| 양성호 | ai | reader | reader | reader | observer |
| 한정현 | service | reader | reader | reader | admin |

원칙:

- 본인 도메인은 admin
- 도메인 외에는 reader 또는 observer
- Keycloak super-admin 은 김용균, break-glass 백업은 윤승호

## 10. 학습자 / 강사 lifecycle

MVP:

- 강사가 admin console 에서 학습자 계정 생성
- 임시 비밀번호 부여
- 학습자는 첫 로그인 시 비밀번호 변경 강제
- self-registration 비활성 유지

장기:

- cohort 사용자 CSV import
- 별도 ADR 후 Google Workspace 또는 기업 SAML IdP 연동

## 11. Token / Session 정책

| 항목 | 값 | 근거 |
|---|---|---|
| Access token TTL | 15분 | UX 와 보안 균형 |
| Refresh token TTL | 8시간 | 1일 1세션 학습 흐름 |
| SSO session idle | 30분 | 학습자 PC 미사용 보호 |
| SSO session max | 8시간 | refresh token 과 정렬 |
| Offline token | 비활성 | MVP 보안 모델 단순화 |
| Remember me | 비활성 | 학습자 PC 공유 risk |

## 12. Audit 이벤트 정책

| 카테고리 | Keycloak 이벤트 타입 | 대상 |
|---|---|---|
| 인증 | `LOGIN_SUCCESS`, `LOGIN_ERROR`, `LOGOUT`, `REGISTER`, `UPDATE_PASSWORD` | `security-logs` |
| Admin 변경 | `CREATE_USER`, `DELETE_USER`, `ASSIGN_ROLE`, `UPDATE_USER`, `IMPERSONATE` | `security-logs` |
| Client 로그인 | `CLIENT_LOGIN`, `CLIENT_LOGIN_ERROR` | `security-logs` |

Kafka 전송은 Strimzi 준비 후 Keycloak Event Listener SPI 또는 Webhook-to-Kafka publisher 로 별도 PR 에서 구현.

## 13. Bootstrap / DR

- Initial super-admin: 김용균 (`kylekim`)
- Admin credential 보관: 1Password Team Vault `Cledyu/keycloak-admin`
- 윤승호가 동일 1Password 항목 break-glass 접근권 보유
- Realm export 백업: 일 1회 Kubernetes CronJob, S3 30일 보존
- super-admin 비밀번호 회전: 90일

## 14. 결과 (Consequences)

긍정:

- 외부 노출 후 default admin 의존 제거
- 5개 서비스의 OIDC 입력값이 명시 — 후속 통합 PR 의 ambiguity 제거
- Group 기반 최소권한으로 팀 blast radius 격리
- Audit 정책이 `security-logs` SIEM 파이프라인과 정렬

트레이드오프:

- 학습자 self-registration 비활성으로 강사가 MVP 계정 생성 부담
- ArgoCD project / Grafana folder 의 세부 권한은 downstream application 에 위임
- Kafka audit 전송은 Strimzi 준비를 기다림

## 15. 검증 기준

- 6명 팀원 모두 `https://keycloak.cledyu.local` 로그인 가능
- 각 팀원이 의도한 group 에 자동 매핑됨
- `web`, `api`, `tutor`, `argocd`, `grafana` client 가 `cledyu` realm 에 존재
- `student`, `instructor`, `admin`, `observer` realm role 이 존재
- 학습자 초대 시나리오 e2e: 강사가 학습자 생성 → 학습자 첫 로그인 시 비밀번호 변경 강제 → Lab 진입
- `kcadm.sh get events` 로 LOGIN / LOGIN_ERROR 이벤트 확인

## 16. 참고

- Terraform provider: `mrparkers/keycloak`
- Keycloak Admin Console: `https://keycloak.cledyu.local/admin/`
