# apps/web — Cledyu 프론트엔드

Next.js 14 (App Router) 기반 학습자/강사 UI.

## 기술 스택

| 항목 | 선택 |
|---|---|
| 프레임워크 | Next.js 14 (App Router, RSC) |
| 언어 | TypeScript |
| 스타일 | Tailwind CSS |
| 서버 상태 | TanStack Query v5 |
| 터미널 | @xterm/xterm + WebSocket |
| 패키지 매니저 | pnpm |
| 컨테이너 | distroless/nodejs20-debian12 |

## 페이지 구조

```
/               → /labs 로 리다이렉트
/login          → Keycloak SSO 로그인 (현재 mock 버튼)
/labs           → Lab 카탈로그 (난이도 필터)
/labs/[id]      → 터미널 + 단계 목록 + AI 힌트
/instructor     → 강사 대시보드 (수강생 세션 관제)
```

## 로컬 실행

```bash
cd apps/web
pnpm install
pnpm dev        # http://localhost:3000
```

개발 모드에서는 `middleware.ts` 인증 우회, API 요청에 `Authorization: Bearer dev-token` 자동 주입.

## 환경변수

| 변수 | 종류 | 기본값 | 설명 |
|---|---|---|---|
| `BACKEND_URL` | 런타임 (서버) | `http://localhost:8080` | Next.js rewrite 대상. HTTP `/api/*` 프록시 |
| `NEXT_PUBLIC_WS_URL` | **빌드 타임** | `ws://localhost:8080` | 터미널 WebSocket URL (브라우저 직접 연결) |

> **주의** — `NEXT_PUBLIC_*` 변수는 `pnpm build` 시점에 JS 번들에 포함됨.
> 런타임 env로 주입해도 반영되지 않음.
> 백엔드 외부 URL 확정 후 Dockerfile `--build-arg`로 전달 필요.

## Docker 빌드

```bash
docker build -t web \
  --build-arg NEXT_PUBLIC_WS_URL=wss://api.cledyu.local \
  apps/web/
```

## 알려진 이슈

- `/leaderboard` — Navbar에 링크가 있으나 페이지 미구현 (작업 설명서 외 범위)
- `NEXT_PUBLIC_WS_URL` — 백엔드 Ingress 확정 전까지 터미널 WebSocket 미동작
