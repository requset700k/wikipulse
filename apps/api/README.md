# apps/api — Cledyu 백엔드

Go + Gin 기반 REST/WebSocket API 서버.

## 기술 스택

| 항목 | 선택 |
|---|---|
| 언어 | Go 1.24 |
| 프레임워크 | Gin |
| DB | PostgreSQL |
| 캐시 | Redis |
| VM 프로비저닝 | KubeVirt / EC2 (hybrid) |
| 컨테이너 | distroless/static-debian12 |

## 엔드포인트 구조

```
GET  /healthz                              헬스체크
GET  /api/v1/openapi                       OpenAPI 스펙
POST /api/v1/auth/login                    Keycloak OIDC 시작
GET  /api/v1/auth/callback                 OIDC 콜백
POST /api/v1/auth/logout
GET  /api/v1/me                            내 프로필 (JWT 필요)
GET  /api/v1/labs                          Lab 목록
GET  /api/v1/labs/:id
POST /api/v1/sessions                      세션 생성 (VM 프로비저닝)
GET  /api/v1/sessions/:id
DELETE /api/v1/sessions/:id
GET  /api/v1/sessions/:id/steps
POST /api/v1/sessions/:id/validate
POST /api/v1/sessions/:id/hint
GET  /api/v1/sessions/:id/ws               터미널 WebSocket
GET  /api/v1/instructor/sessions           강사 전용
POST /api/v1/instructor/sessions/:id/inject
```

## 로컬 실행

```bash
cd apps/api
cp config.example.yaml config.yaml
# config.yaml 에서 DB/Redis/Keycloak URL 수정
go run ./cmd/server

# 또는 hot-reload (air 설치 필요)
air
```

## 환경변수 / 설정

`config.yaml` 또는 환경변수로 주입 (config.go 참조).  
`config.example.yaml` 에 전체 항목과 기본값 정리되어 있음.

## Docker 빌드

```bash
docker build -t api apps/api/
```

최종 이미지: `distroless/static-debian12` (shell 없음, 바이너리만 포함).  
디버깅: `kubectl debug -it <pod> --image=busybox --target=api`

## 마이그레이션

```bash
psql $DATABASE_URL < migrations/001_init.sql
```
