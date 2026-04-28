# service-web 배포

## 트리거

- `feat/service-init` 브랜치를 main에 머지할 때 (최초 배포)
- 새 기능 PR이 main에 머지되어 이미지 태그를 갱신할 때
- pod CrashLoopBackOff / ImagePullBackOff 발생 시

## 사전 조건

- `kubectl` + OIDC kubeconfig 설정 완료 (`infra/kubernetes/kubeconfig/cledyu-oidc.yaml`)
- ArgoCD CLI 설치 (`brew install argocd`) 또는 ArgoCD UI 접근 가능
- GHCR `web` 패키지가 public 으로 전환되어 있어야 클러스터가 pull 가능 (최초 1회)

```bash
# 사전 체크
kubectl get nodes
gh api "/orgs/requset700k/packages/container/web" --jq '{visibility}'
```

## 절차

### 최초 배포

#### 1) PR 머지 → 이미지 빌드

`feat/service-init` → main 머지 시 `build-web.yml` 자동 실행.
PR 단계: build only / main 머지: build + push + Trivy scan.

#### 2) GHCR `web` 패키지 public 전환 (org admin 1회)

```bash
# UI: https://github.com/orgs/requset700k/packages/container/web/settings
# → Change visibility → Public
```

#### 3) 실제 SHA로 image.tag 업데이트

```bash
# GitHub Actions 로그 또는 아래 명령으로 SHA 확인
git log main --oneline -1

# gitops/apps/web/values.yaml 수정
# image.tag: sha-0000000  →  sha-<실제 7자>
# 커밋 후 push
```

#### 4) ArgoCD Application 적용

```bash
kubectl apply -f gitops/argocd/apps/service-web.yaml
```

### 일반 배포 (이미지 태그 갱신)

```bash
# 1. gitops/apps/web/values.yaml 의 image.tag 를 새 SHA 로 변경
# 2. 커밋 후 main 푸시
# 3. ArgoCD 자동 sync (automated + selfHeal 설정됨)
```

## 검증

```bash
kubectl -n web get pods
# 예상 출력: web-xxxx   1/1   Running

kubectl -n web get certificate web-tls
# 예상 출력: READY=True

# DNS 없이 Ingress 접속 검증
curl -k --resolve app.cledyu.local:443:10.10.0.101 https://app.cledyu.local
# 예상 출력: HTTP 200, Next.js HTML
```

## 롤백

```bash
# values.yaml 의 image.tag 를 이전 SHA 로 되돌린 후 커밋 + push
# ArgoCD 자동 sync, 또는 즉시 반영:
argocd app rollback service-web
```

## 참고

- 빌드 파이프라인: `.github/workflows/build-web.yml`, `build-image.yml`
- 이미지 관리 공통: `docs/RUNBOOK/container-image.md`
- 개발자 가이드: `apps/web/README.md`
- Keycloak RBAC (web client): `docs/ADR/keycloak-rbac.md`

### NEXT_PUBLIC_WS_URL 설정 (백엔드 Ingress 확정 후)

터미널 WebSocket URL은 빌드 타임에 번들에 포함됨. 백엔드 Ingress 생성 후 `build-web.yml`에 `build-args` 추가 필요.

```bash
docker build --build-arg NEXT_PUBLIC_WS_URL=wss://api.cledyu.local apps/web/
```

### 디버깅 (distroless)

distroless 이미지라 `kubectl exec -- sh` 불가. ephemeral container 사용:

```bash
kubectl debug -it <pod-name> -n web --image=busybox --target=web
```
