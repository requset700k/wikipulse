# ADR: Keycloak RBAC Model

- **Status:** Proposed
- **Date:** 2026-04-26
- **Proposer:** 김용균 / Platform
- **Decision makers:** 김용균 / 윤승호
- **Related:** Keycloak Ingress + TLS, Keycloak Client Trust Setup

## 1. Context

`https://keycloak.cledyu.local` is now externally reachable through Traefik and
Cledyu internal TLS. From this point, running only with the bootstrap admin and
an empty realm creates avoidable security debt.

This ADR defines the application-level Keycloak authorization model so the
realm, clients, roles, groups, and team member access can be implemented
declaratively with Terraform.

## 2. Problem / Goals

Goals:

- Minimize the window where the default bootstrap admin is used for operations.
- Apply least privilege across the 6 team members.
- Define OIDC inputs for `web`, `api`, `tutor`, `argocd`, and `grafana`.
- Define the MVP lifecycle for instructors and students.
- Keep authentication and admin audit events aligned with the `security-logs`
  pipeline.

Non-goals for this ADR:

- External learner realm separation.
- Google Workspace or SAML IdP integration.
- Keycloak audit event forwarding to Kafka. This depends on Strimzi and remains
  a follow-up PR.

## 3. Considered Options

### Option A: Single `cledyu` realm

- Pros: Simple, low operational overhead, appropriate for a 6-person internal
  project.
- Cons: External learner separation requires a future ADR if the project expands.

### Option B: `master` plus `cledyu` with stricter admin isolation

- Pros: Stronger separation between platform operators and application users.
- Cons: More operational complexity than the MVP needs.

### Option C: Split `cledyu-internal` and `cledyu-external`

- Pros: Best long-term isolation for external learners or enterprise customers.
- Cons: Premature for the current 3-month internal project.

## 4. Decision

Adopt Option A.

- `master` is reserved for Keycloak super-admin access only.
- Business users, services, learners, instructors, and team groups live in the
  `cledyu` realm.
- The model is implemented with Terraform using the `mrparkers/keycloak`
  provider.

## 5. Realm

| Realm | Purpose |
|---|---|
| `cledyu` | All Cledyu service users, team users, learners, instructors, roles, and OIDC clients |

Realm policy:

- Self-registration disabled.
- Remember-me disabled.
- Offline tokens disabled for the MVP.
- First login password update is required for managed users.

## 6. Clients

| Client | Service | Type | Flow | Notes |
|---|---|---|---|---|
| `web` | Next.js | public | Authorization Code + PKCE | SPA-facing login |
| `api` | Go/Gin | bearer-only | JWT validation only | No direct login |
| `tutor` | FastAPI | bearer-only | JWT validation only | No direct login |
| `argocd` | ArgoCD | confidential | Authorization Code | Uses group claim mapping |
| `grafana` | Grafana | confidential | Authorization Code | Uses role/group mapping |

## 7. Realm Roles

| Role | Meaning | Primary assignee |
|---|---|---|
| `student` | Lab access and AI tutor usage | Learners |
| `instructor` | Learner creation, lab variation, instructor mode | Instructors |
| `admin` | Operational administration | Team domain owners |
| `observer` | Read-only dashboard/log access | Cross-domain team access |

Client roles are intentionally not introduced in the MVP. If ArgoCD projects,
Grafana folders, or service-specific scopes need stronger separation, client
roles will be added in a follow-up ADR.

## 8. Groups

| Group | Realm role mapping | Members |
|---|---|---|
| `team-platform` | `admin` | 김용균 |
| `team-security` | `admin` | 윤승호 |
| `team-observability` | `admin`, `observer` | 조승연 |
| `team-lab-data` | `admin`, `observer` | 김찬영 |
| `team-ai` | `admin`, `observer` | 양성호 |
| `team-service` | `admin`, `observer` | 한정현 |
| `students-cohort-0` | `student` | Learners invited by instructors |
| `instructors` | `instructor` | Instructors |

Domain-specific authorization is handled by downstream applications through
group claims. Keycloak only emits the group and role claims.

## 9. Team Permission Matrix

| Name | Team | Keycloak | ArgoCD | Grafana | Web/API/Tutor |
|---|---|---|---|---|---|
| 김용균 | platform | super-admin | admin | admin | admin |
| 윤승호 | security | admin | reader | reader | observer |
| 조승연 | observability | reader | reader | admin | observer |
| 김찬영 | lab-data | reader | reader | reader | observer |
| 양성호 | ai | reader | reader | reader | observer |
| 한정현 | service | reader | reader | reader | admin |

Principle:

- Admin on the owner domain.
- Reader or observer outside the owner domain.
- Keycloak super-admin is held by 김용균, with 윤승호 as break-glass backup.

## 10. Learner / Instructor Lifecycle

MVP:

- Instructor creates learner accounts from the admin console.
- Instructor assigns temporary passwords.
- Learners must change passwords on first login.
- Self-registration remains disabled.

Future:

- CSV import for cohort users.
- Google Workspace or enterprise SAML IdP integration after a separate ADR.

## 11. Token / Session Policy

| Item | Value | Reason |
|---|---|---|
| Access token TTL | 15 minutes | Balance UX and security |
| Refresh token TTL | 8 hours | One-day learning session |
| SSO session idle | 30 minutes | Protect unattended learner devices |
| SSO session max | 8 hours | Align with refresh token lifetime |
| Offline token | Disabled | Simpler MVP security model |
| Remember me | Disabled | Shared learner PC risk |

## 12. Audit Event Policy

| Category | Keycloak event types | Target |
|---|---|---|
| Authentication | `LOGIN_SUCCESS`, `LOGIN_ERROR`, `LOGOUT`, `REGISTER`, `UPDATE_PASSWORD` | `security-logs` |
| Admin changes | `CREATE_USER`, `DELETE_USER`, `ASSIGN_ROLE`, `UPDATE_USER`, `IMPERSONATE` | `security-logs` |
| Client login | `CLIENT_LOGIN`, `CLIENT_LOGIN_ERROR` | `security-logs` |

Kafka forwarding is implemented later through a Keycloak Event Listener SPI or
Webhook-to-Kafka publisher after Strimzi is ready.

## 13. Bootstrap / DR

- Initial super-admin: `kylekim`.
- Initial admin credential is stored in 1Password Team Vault:
  `Cledyu/keycloak-admin`.
- 윤승호 has break-glass access to the same 1Password item.
- Realm export backup is planned as a daily Kubernetes CronJob with 30-day S3
  retention.
- Super-admin password rotation target: every 90 days.

## 14. Consequences

Positive:

- Keycloak no longer depends on ad-hoc default admin usage after exposure.
- OIDC integration inputs are explicit for all 5 services.
- Team blast radius is reduced through group-based least privilege.
- Audit policy aligns with the future `security-logs` SIEM pipeline.

Trade-offs:

- Learner self-registration is disabled, so instructors carry the MVP account
  creation workload.
- Detailed ArgoCD project or Grafana folder authorization is deferred to
  downstream applications.
- Kafka audit forwarding waits for Strimzi readiness.

## 15. Verification Criteria

- All 6 team users can log in to `https://keycloak.cledyu.local`.
- Each team user lands in the expected group.
- `web`, `api`, `tutor`, `argocd`, and `grafana` clients exist in the `cledyu`
  realm.
- `student`, `instructor`, `admin`, and `observer` realm roles exist.
- A learner invite flow succeeds:
  instructor creates learner, learner first login forces password update, and
  learner enters the lab.
- `kcadm.sh get events` shows login and login-error events.

## 16. References

- Terraform provider: `mrparkers/keycloak`
- Keycloak Admin Console: `https://keycloak.cledyu.local/admin/`
