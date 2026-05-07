# Vault 초기 부트스트랩 런북

## 작업 범위

이 문서는 Cledyu 클러스터에 Vault를 처음 배포한 뒤 초기화, 루트 토큰 발급, 수동 언실까지 진행하는 절차를 정리함.

이번 단계는 Shamir 방식의 수동 언실을 사용함. GCP Cloud KMS 기반 auto-unseal, Kubernetes Auth, External Secrets Operator, 기존 시크릿 이관은 후속 작업으로 진행함.

## 선행 조건

- `platform-vault` ArgoCD Application 또는 Helm 배포가 적용되어 있어야 함.
- `vault` namespace가 존재해야 함.
- `vault-tls` Certificate가 `Ready=True` 상태여야 함.
- Vault Pod 3개가 Running 상태여야 함.

확인 명령어:

```bash
kubectl -n vault get pods
kubectl -n vault get pvc
kubectl -n vault get certificate vault-tls
kubectl -n vault get ingress vault
```

## 초기 상태 확인

초기화 전 Vault 상태 확인:

```bash
kubectl -n vault exec vault-0 -- env VAULT_SKIP_VERIFY=true vault status
```

초기화 전 기대 상태:

```text
Initialized: false
Sealed: true
```

## Vault 초기화

초기화는 `vault-0`에서 한 번만 실행함.

```bash
kubectl -n vault exec vault-0 -- \
  env VAULT_SKIP_VERIFY=true \
  vault operator init \
    -key-shares=5 \
    -key-threshold=3 \
    -format=json
```

명령 실행 결과에는 unseal key 5개와 root token이 포함됨. 이 값은 절대 GitHub, Discord, 공개 Notion, PR 코멘트, 쉘 히스토리에 남기지 않음.

권장 1Password 저장 위치:

```text
Title: Cledyu/vault-bootstrap
Vault URL: https://vault.cledyu.local
Unseal Key 1: <redacted>
Unseal Key 2: <redacted>
Unseal Key 3: <redacted>
Unseal Key 4: <redacted>
Unseal Key 5: <redacted>
Root Token: <redacted>
Key Shares: 5
Key Threshold: 3
Created by: 윤승호
Access: 김용균, 윤승호
```

## 수동 언실

각 Vault Pod에 unseal key 5개 중 3개를 입력함. 같은 key 3개 조합을 모든 Pod에 사용할 수 있음.

```bash
kubectl -n vault exec -it vault-0 -- env VAULT_SKIP_VERIFY=true vault operator unseal
kubectl -n vault exec -it vault-0 -- env VAULT_SKIP_VERIFY=true vault operator unseal
kubectl -n vault exec -it vault-0 -- env VAULT_SKIP_VERIFY=true vault operator unseal

kubectl -n vault exec -it vault-1 -- env VAULT_SKIP_VERIFY=true vault operator unseal
kubectl -n vault exec -it vault-1 -- env VAULT_SKIP_VERIFY=true vault operator unseal
kubectl -n vault exec -it vault-1 -- env VAULT_SKIP_VERIFY=true vault operator unseal

kubectl -n vault exec -it vault-2 -- env VAULT_SKIP_VERIFY=true vault operator unseal
kubectl -n vault exec -it vault-2 -- env VAULT_SKIP_VERIFY=true vault operator unseal
kubectl -n vault exec -it vault-2 -- env VAULT_SKIP_VERIFY=true vault operator unseal
```

## 상태 검증

각 Pod 상태 확인:

```bash
kubectl -n vault exec vault-0 -- env VAULT_SKIP_VERIFY=true vault status
kubectl -n vault exec vault-1 -- env VAULT_SKIP_VERIFY=true vault status
kubectl -n vault exec vault-2 -- env VAULT_SKIP_VERIFY=true vault status
```

기대 상태:

```text
Initialized: true
Sealed: false
Storage Type: raft
HA Enabled: true
HA Mode: active 또는 standby
```

외부 경로 확인:

```bash
curl --cacert infra/kubernetes/kubeconfig/cledyu-root-ca.pem \
  https://vault.cledyu.local/v1/sys/health
```

Windows에서 인증서 폐기 확인 문제로 실패하면 아래처럼 확인할 수 있음.

```powershell
curl.exe --ssl-no-revoke `
  --resolve vault.cledyu.local:443:10.10.0.101 `
  -i https://vault.cledyu.local/v1/sys/health
```

초기화와 언실이 완료된 active node 기대 응답:

```text
HTTP/1.1 200 OK
initialized: true
sealed: false
standby: false
```

## 후속 작업

### 완료됨

- Kubernetes Auth backend 활성화.
- `cledyu/` KV v2 secrets engine 생성.
- Vault policy 5종 생성.
- Kubernetes ServiceAccount role mapping 4종 생성.
- 초기 시크릿 이관.
- Keycloak admin credential 이관.
- Keycloak Postgres credential 이관.
- ArgoCD OIDC client secret 이관.
- `web`, `api`, `tutor` OIDC client metadata 이관.
- `grafana` OIDC client는 secret 미생성 상태라 pending metadata로 기록.
- File audit device 활성화.

### 남음

- root token / unseal key 1Password 팀 vault 최종 등록.
- GCP Cloud KMS auto-unseal 구성.
- 수동 언실 의존성 제거.
- Grafana OIDC client secret 생성 후 Vault 값 갱신.
- Google AI API key 이관.
- Strimzi 준비 후 audit log를 `security-logs` 파이프라인으로 연동.

## Kubernetes Auth / Secret Migration

아래 스크립트는 root token을 로컬 bootstrap JSON에서 읽고, 값을 화면에 출력하지 않은 채 Vault에 초기 설정을 적용함.

```powershell
PowerShell -ExecutionPolicy Bypass -File scripts/vault-bootstrap-configure.ps1
```

적용 항목:

```text
Secrets engine:
- cledyu/                 kv-v2

Auth backend:
- kubernetes/

Audit device:
- file/                    /vault/audit/audit.log

Policies:
- cledyu-argocd
- cledyu-grafana
- cledyu-keycloak-admin
- cledyu-keycloak-db
- cledyu-service-oidc

Kubernetes auth roles:
- cledyu-argocd           argocd/argocd-server
- cledyu-grafana          monitoring/grafana
- cledyu-keycloak         keycloak/cledyu-keycloak
- cledyu-services         web/api/tutor service accounts

Migrated paths:
- cledyu/keycloak/admin
- cledyu/keycloak/postgres
- cledyu/oidc/argocd
- cledyu/oidc/web
- cledyu/oidc/api
- cledyu/oidc/tutor
- cledyu/oidc/grafana
```

검증 명령어:

```bash
vault secrets list
vault auth list
vault audit list
vault policy list
vault list auth/kubernetes/role
vault kv metadata get cledyu/keycloak/admin
vault kv metadata get cledyu/keycloak/postgres
vault kv metadata get cledyu/oidc/argocd
```

## Audit Log 위치와 보존 정책

Vault file audit device는 다음 위치에 기록함.

```text
Audit device: file/
File path: /vault/audit/audit.log
Storage: Vault auditStorage PVC
Size: 5Gi
```

운영 확인 명령:

```bash
vault audit list
vault token lookup
tail -n 5 /vault/audit/audit.log
```

보존 정책:

- 감사 로그는 Vault `auditStorage` PVC에 보존함.
- `vault token lookup`, secret read/write, auth backend 호출 등 Vault API 요청이 HMAC 처리된 형태로 기록됨.
- 원본 audit log에는 민감한 접근 패턴이 포함될 수 있으므로 GitHub, Discord, 공개 Notion에 원문 공유 금지.
- 단일 파일 무한 append로 PVC가 가득 차면 Vault 요청 처리가 중단될 수 있으므로 운영 전 logrotate 또는 sidecar 기반 로테이션을 적용함.
- 장기 보존은 Strimzi Kafka `security-logs` 토픽과 S3 Glacier 파이프라인이 준비된 뒤 연동함.
- 장기 파이프라인 전까지는 `vault-0`의 `/vault/audit/audit.log`를 1차 포렌식 근거로 사용함.

로테이션 예시:

```text
/vault/audit/audit.log {
    rotate 7
    daily
    compress
    missingok
    postrotate
        kill -HUP $(pidof vault)
    endpostrotate
}
```

## GCP KMS Auto-Unseal 전환

현재 PR에서는 `values-gcpckms.example.yaml`만 추가함. 실제 전환은 GCP KMS 키 정보와 권한이 준비된 뒤 진행함.

전환 전 필요한 값:

```text
GCP project id
KMS region
KMS key ring
KMS crypto key
Vault Pod가 사용할 GCP credential 또는 Workload Identity
```

전환 순서:

```text
1. GCP KMS key 생성
2. Vault Pod 권한 부여
3. vault-gcp-kms Secret 또는 Workload Identity 구성
4. values.yaml에 seal "gcpckms" 블록 반영
5. Helm/ArgoCD sync
6. Vault 재시작 후 sealed=false 확인
7. 수동 unseal 의존성 제거
```
