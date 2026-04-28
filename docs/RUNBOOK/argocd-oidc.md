# ArgoCD OIDC + Traefik Ingress 운영 런북

ArgoCD 를 Keycloak 으로 OIDC 로그인 + Traefik Ingress 로 외부 노출 + cert-manager 인증서로 보호하는 구성의 운영 가이드.

## 1. 구성 요약

```
브라우저 (https://argocd.cledyu.local)
   ↓ DNS: hosts 등록 (10.10.0.101) — local_dns Ansible role
Traefik (10.10.0.101, TLS 종료, cert=argocd-tls)
   ↓ HTTP
ArgoCD Server (ClusterIP, --insecure)
   ↓ OIDC discovery (cledyu-ca 신뢰)
Keycloak (https://keycloak.cledyu.local/realms/cledyu, client=argocd)
   ↓ JWT (groups claim)
ArgoCD RBAC (Keycloak group → role 매핑)
```

## 2. 영구화된 매니페스트

| 파일 | 역할 |
|---|---|
| `gitops/apps/argocd/values.yaml` | Helm values — Ingress / OIDC config / RBAC / rootCA placeholder |
| `infra/kubernetes/argocd/certificate.yaml` | cert-manager Certificate (`argocd-tls`, 90일/30일 갱신) |
| `infra/kubernetes/coredns/configmap.yaml` | CoreDNS hosts plugin — `*.cledyu.local → 10.10.0.101` (클러스터 내부 OIDC discovery 용) |
| `infra/terraform/keycloak/clients.tf` | Keycloak client `argocd` 정의 + group claim mapper |

## 3. 처음 배포 절차 (재구성 시)

```bash
# 1. Cert
kubectl apply -f infra/kubernetes/argocd/certificate.yaml
kubectl -n argocd wait --for=condition=Ready certificate/argocd-tls --timeout=2m

# 2. CoreDNS hosts (클러스터 내부 cledyu.local 해석용)
kubectl apply -f infra/kubernetes/coredns/configmap.yaml
kubectl -n kube-system rollout restart deployment coredns

# 3. Helm values 의 rootCA placeholder 를 실제 PEM 으로 치환
ROOT_CA_PEM=$(kubectl -n cert-manager get secret cledyu-root-ca \
  -o jsonpath='{.data.tls\.crt}' | base64 -d)
ROOT_CA_INDENTED=$(echo "$ROOT_CA_PEM" | sed 's/^/        /')
sed "/# NOTE:.*placeholder/,/-----END CERTIFICATE-----/c\\
$ROOT_CA_INDENTED" gitops/apps/argocd/values.yaml > /tmp/argocd-values-rendered.yaml

# 4. argocd-secret 에 OIDC client secret 주입 (1Password 에서 받아서)
ARGOCD_SECRET=$(op item get "Cledyu/keycloak-bootstrap-secrets" --fields argocd)
SECRET_B64=$(echo -n "$ARGOCD_SECRET" | base64)
kubectl -n argocd patch secret argocd-secret --type=merge \
  -p "{\"data\":{\"oidc.keycloak.clientSecret\":\"$SECRET_B64\"}}"

# 5. Helm install/upgrade
helm repo add argo https://argoproj.github.io/argo-helm
helm upgrade --install argocd argo/argo-cd \
  -n argocd --create-namespace \
  --version 7.7.10 \
  -f /tmp/argocd-values-rendered.yaml

# 6. 검증
kubectl -n argocd rollout status deployment argocd-server
curl -kI https://argocd.cledyu.local/
curl -sk https://argocd.cledyu.local/api/v1/settings | jq '.oidcConfig.name'
# → "Keycloak" 출력되면 성공
```

## 4. RBAC 매트릭스 (ADR-0001 동기화)

| Group (Keycloak) | ArgoCD Role | 의미 |
|---|---|---|
| `team-platform` | `role:admin` | 전체 관리 (김용균) |
| `team-security` | `role:readonly` | 윤승호 |
| `team-observability` | `role:readonly` | 조승연 |
| `team-lab-data` | `role:readonly` | 김찬영 |
| `team-ai` | `role:readonly` | 양성호 |
| `team-service` | `role:readonly` | 한정현 |
| (default) | `role:readonly` | 그 외 |

권한 확장 시 `gitops/apps/argocd/values.yaml` 의 `configs.rbac.policy.csv` 수정.

## 5. Secret 분배

`oidc.keycloak.clientSecret` (32-byte hex) 의 출처:
- 1Password 항목: `Cledyu/keycloak-bootstrap-secrets`
- 필드: `argocd`
- 발급: Keycloak Terraform apply 시점 (`infra/terraform/keycloak/terraform.tfvars`)
- 회전 시: Terraform 의 `oidc_client_secrets.argocd` 갱신 → terraform apply → 1Password 갱신 → 위 §3 의 step 4 재실행

## 6. 트러블슈팅

| 증상 | 원인 | 해결 |
|---|---|---|
| `LOG IN VIA KEYCLOAK` 버튼 안 뜸 | `argocd-cm` 의 `oidc.config` YAML parse 에러 | `kubectl -n argocd logs -l app=argocd-server \| grep oidc` 로 line 번호 확인. 보통 nested block scalar (`rootCA: \|`) 의 들여쓰기. |
| `dial tcp: lookup keycloak.cledyu.local on 10.43.0.10:53: no such host` | CoreDNS hosts plugin 에 미등록 | 위 §3 step 2 (configmap 적용 + restart) |
| `invalid_scope: openid profile email groups` | Keycloak realm 에 `groups` client scope 미등록 | Quick fix: values.yaml 의 `requestedScopes` 에서 `groups` 제거. Proper fix: Terraform 에 `keycloak_openid_client_scope "groups"` 추가 (후속 PR). |
| `http: named cookie not present` | 비번 변경 등으로 OIDC state cookie expire | 사이트 데이터 (cookie) 삭제 또는 incognito 로 재시도 |
| 브라우저 인증서 경고 | macOS keychain 에 `cledyu-root-ca` 미신뢰 | `bash scripts/trust-cledyu-root-ca.sh` |

## 7. 후속 작업

- [ ] **ArgoCD self-managed 화** — `gitops/argocd/apps/platform-argocd.yaml` Application 으로 helm release 자동 sync (현재는 helm install/upgrade 수동)
- [ ] **`groups` client scope 정식 등록** — Terraform 매니페스트에 추가 후 values.yaml 의 `requestedScopes` 에 `groups` 다시 포함
- [ ] **CoreDNS ConfigMap 변경 자동화** — Ansible role `coredns_hosts` 또는 Kustomize patch
- [ ] **OIDC client secret 의 Sealed Secret/Vault 마이그레이션** — 현재 1Password 수동 patch
