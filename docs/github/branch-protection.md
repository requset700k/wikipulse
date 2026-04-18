# GitHub 브랜치 보호 규칙 (Free tier 호환)

GitHub Free tier **조직** 또는 **퍼블릭 레포**에서 사용 가능한 설정만 기재합니다.
프라이빗 + Free tier는 일부 보호 규칙이 제한되므로 주석 참고.

## 설정 경로

`Repository → Settings → Branches → Branch protection rules → Add rule`

## `main` 브랜치 규칙 권장값

| 항목 | 값 | 비고 |
|---|---|---|
| Branch name pattern | `main` | |
| Require a pull request before merging | ✅ ON | |
| &nbsp;&nbsp;├ Require approvals | ✅ **1** | Free tier OK |
| &nbsp;&nbsp;├ Dismiss stale pull request approvals when new commits are pushed | ✅ ON | |
| &nbsp;&nbsp;└ Require review from Code Owners | ❌ **OFF** | Free tier 프라이빗 레포에서 강제 불가 — 수동 지정 사용 |
| Require status checks to pass before merging | ✅ ON | |
| &nbsp;&nbsp;├ Require branches to be up to date before merging | ✅ ON | |
| &nbsp;&nbsp;└ Required status checks | `pre-commit (all hooks)`, `PR title (Conventional Commits)`, `terraform validate (infra/terraform/envs/onprem)`, `terraform validate (infra/terraform/kvm)`, `gitleaks (secrets scan)` | 각 워크플로 첫 실행 후 목록에 노출됨 |
| Require conversation resolution before merging | ✅ ON | |
| Require signed commits | ⚠️ 선택 | GPG 셋업 합의 후 활성화 |
| Require linear history | ✅ ON | Squash/Rebase merge만 허용과 잘 맞음 |
| Require deployments to succeed before merging | ❌ OFF | Environments 필요 (유료 기능 일부) |
| Lock branch | ❌ OFF | |
| Do not allow bypassing the above settings | ✅ ON | 관리자도 적용 |
| Restrict who can push to matching branches | ❌ OFF (Free) | 조직 Team 단위 제한은 유료 |
| Allow force pushes | ❌ OFF | |
| Allow deletions | ❌ OFF | |

## 병합 전략

`Settings → General → Pull Requests`

- Allow **squash** merging : ✅ (기본값)
- Allow **rebase** merging : ✅
- Allow **merge commits** : ❌
- Default merge commit message: **"Pull request title and description"**
- Automatically delete head branches: ✅

## 릴리스 태그 보호 (선택)

`Settings → Tags → Add rule` → `v*` 패턴 보호, 관리자만 생성 가능.

## 설정 순서 (최초 1회)

1. `.github/workflows/lint.yml` 과 `.github/workflows/commitlint.yml` 을 `main`에 머지
2. 각 워크플로를 최소 1회 실행 (PR 또는 `workflow_dispatch`)
3. 브랜치 보호 규칙에서 위 status checks 지정
4. `.github/labels.yml` 머지 → `labels-sync` 워크플로가 라벨 자동 생성

## 일반 리포 설정 권장 (Settings → General)

- Default branch: `main`
- Features: Issues ✅, Discussions ✅ (설계 토론), Projects ✅, Wiki ❌
- Pull Requests: 위 "병합 전략" 참고
- Pushes: Limit how many branches can be updated ❌ (Free)
- Archives: Include Git LFS objects ❌

## Actions 권한 (Settings → Actions → General)

- Workflow permissions: **Read and write permissions** (labels-sync에 필요)
- Allow GitHub Actions to create and approve pull requests: ✅ (Dependabot 자동 병합 시)
- Fork pull request workflows from outside collaborators: **Require approval for first-time contributors**

## Secrets / Variables (Settings → Secrets and variables → Actions)

레포 시크릿 예:
- `TF_VAR_*` — Terraform 변수 (민감값만)
- `ANSIBLE_VAULT_PASSWORD` — Ansible Vault
- `KUBECONFIG_DEV` — DR 드릴용 (Base64 인코딩)

**절대 커밋 금지**: Vault 토큰, AWS/GCP long-lived keys → OIDC federation 우선 검토.
