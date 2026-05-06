# Next.js 14 standalone 빌드 → distroless 최소 이미지 3단계 파이프라인.
# standalone 모드(next.config.mjs output:'standalone')가 활성화되어 있어야
# builder 단계에서 .next/standalone 디렉터리가 생성됨.

FROM node:20-alpine AS base

# ── deps: 의존성 설치만 담당 ───────────────────────────────────────────────────
# package.json / pnpm-lock.yaml 만 먼저 COPY 해서 소스 변경 시 이 레이어를
# 재사용(캐시 히트)할 수 있게 분리.
FROM base AS deps
WORKDIR /app
COPY package.json pnpm-lock.yaml ./
# packageManager 필드(pnpm@10.33.2)와 버전을 맞춰 corepack으로 핀.
# --frozen-lockfile: lockfile 불일치 시 빌드 실패로 환경 드리프트 방지.
RUN corepack enable && corepack prepare pnpm@10.33.2 --activate \
    && pnpm install --frozen-lockfile

# ── builder: Next.js 빌드 ──────────────────────────────────────────────────────
FROM base AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
# NEXT_TELEMETRY_DISABLED: 빌드/런타임 단계 모두 텔레메트리 전송 차단.
ENV NEXT_TELEMETRY_DISABLED=1
RUN corepack enable && pnpm build

# ── runner: distroless 최소 이미지 ────────────────────────────────────────────
# distroless/nodejs20: shell·패키지 관리자 없음 → 공격 면적 최소화.
# standalone 결과물(server.js + 필요 node_modules)만 복사하므로
# 전체 node_modules 없이 동작.
FROM gcr.io/distroless/nodejs20-debian12 AS runner
WORKDIR /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
# PORT/HOSTNAME: Next.js standalone server.js 가 참조하는 바인딩 설정.
ENV PORT=3000
ENV HOSTNAME=0.0.0.0

# standalone: 실행에 필요한 서버 코드 + 최소 node_modules
# static: 빌드된 JS/CSS 에셋 (standalone에 포함되지 않아 별도 복사 필요)
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static

EXPOSE 3000
CMD ["server.js"]
