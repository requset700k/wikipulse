# container-image — GHCR 빌드/배포 표준 경로

## 트리거
- 신규 서비스(`apps/<service>/`) 추가 시
- 기존 서비스 이미지를 별도 태그로 재발행해야 할 때 (`workflow_dispatch`)
- 새로 추가된 패키지가 처음 GHCR 에 푸시된 후 — visibility 전환 1회

## 사전 조건
- 서비스 디렉토리에 `Dockerfile` 존재 (`apps/<service>/Dockerfile`)
- caller workflow (`.github/workflows/build-<service>.yml`) 에서 reusable
  `build-image.yml` 을 호출
- GHCR 첫 푸시 후 패키지 visibility 가 `public` 으로 전환되어 있어야 클러스터가
  `imagePullSecret` 없이 pull 가능

## 절차

### 1) 신규 서비스 추가
```bash
mkdir -p apps/<service>
$EDITOR apps/<service>/Dockerfile     # multi-stage 권장
echo '*' > apps/<service>/.dockerignore
echo '!Dockerfile' >> apps/<service>/.dockerignore
```

### 2) caller workflow 작성
파일: `.github/workflows/build-<service>.yml`

```yaml
name: build-<service>
on:
  pull_request:
    branches: [main]
    paths:
      - apps/<service>/**
      - .github/workflows/build-image.yml
      - .github/workflows/build-<service>.yml
  push:
    branches: [main]
    paths:
      - apps/<service>/**
      - .github/workflows/build-image.yml
      - .github/workflows/build-<service>.yml
  workflow_dispatch:

permissions:
  contents: read
  packages: write

jobs:
  image:
    uses: ./.github/workflows/build-image.yml
    with:
      context: apps/<service>
      image-name: <service>
      # target: prod         # multi-stage 시
      # platforms: linux/amd64,linux/arm64   # 멀티 아키텍처 필요 시
    secrets: inherit
```

### 3) PR 생성 → 머지
- PR: build 만 (push X, scan X). Dockerfile syntax / 빌드 가능 여부 검증.
- main 머지 후: build + push + Trivy scan (HIGH/CRITICAL fail).

### 4) 최초 1회 셋업 — GHCR 패키지 visibility 전환 (한 번만)
첫 푸시 후 패키지가 GHCR 에 자동 생성됨 (default: private). 클러스터 pull 위해
public 전환 1회 필요:

```bash
# 1) 푸시된 이미지 확인
gh api "/orgs/requset700k/packages/container/<service>" \
  --jq '{visibility,html_url}'

# 2) UI 에서 Change visibility → Public
#   https://github.com/orgs/requset700k/packages/container/<service>/settings
#
# 또는 CLI (PAT 필요, GITHUB_TOKEN 으로는 visibility 변경 불가):
#   gh api -X PATCH /orgs/requset700k/packages/container/<service> \
#     -f visibility=public
```

## 검증
```bash
# 푸시된 태그 확인
gh api "/orgs/requset700k/packages/container/<service>/versions" \
  --jq '.[] | {tags: .metadata.container.tags, created_at}' | head -10

# 클러스터에서 pull 동작 검증 (imagePullSecret 없이)
kubectl run smoke-<service> \
  --image=ghcr.io/requset700k/<service>:latest \
  --restart=Never --rm -i --tty --command -- echo OK

# Workload manifest 에서 immutable sha 태그 사용 권장 (ArgoCD 핀)
# image: ghcr.io/requset700k/<service>:sha-<7글자>
```

## 태깅 정책

| 태그 | 생성 시점 | 용도 |
|---|---|---|
| `sha-<short>` | 모든 main 푸시 | **ArgoCD 핀 — production manifest 는 이 태그만** |
| `<branch>` | 브랜치 푸시 (현재 main 만) | preview / dev |
| `latest` | main 푸시 | 데모 / `kubectl run` 등 ad-hoc |

## 롤백
이미지 자체 롤백은 단순히 manifest 의 sha 태그를 이전 값으로 되돌린 뒤 ArgoCD
재동기화. 이미지 삭제는 권장 X (이전 manifest 가 참조 중일 수 있음).

```bash
# ArgoCD 에서 직접
argocd app rollback <app-name>

# 또는 manifest PR
# image: ghcr.io/requset700k/<service>:sha-<이전 sha>
```

## 참고
- Reusable workflow: `.github/workflows/build-image.yml`
- 첫 사용자/smoke test: `.github/workflows/build-dummy.yml` + `apps/dummy/`
- Trivy 결과는 워크플로 로그의 "Trivy image scan" step 출력 참조
