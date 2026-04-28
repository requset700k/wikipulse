# deps: 의존성만 별도 설치. package.json/pnpm-lock.yaml이 변경되지 않으면 이 레이어를 캐시로 재사용.
FROM node:20-alpine AS deps
WORKDIR /app
RUN corepack enable
COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

# builder: 소스 코드 빌드. output:standalone이 .next/standalone에 최소 실행 파일을 생성.
FROM node:20-alpine AS builder
WORKDIR /app
RUN corepack enable
COPY --from=deps /app/node_modules ./node_modules
COPY . .
ENV NEXT_TELEMETRY_DISABLED=1
RUN pnpm build

# distroless: shell 없음 → 보안 강화. 디버깅 필요 시 kubectl debug 사용.
# nonroot UID=65532. deployment.yaml의 runAsUser와 반드시 일치해야 함.
FROM gcr.io/distroless/nodejs20-debian12 AS runner
LABEL org.opencontainers.image.source="https://github.com/requset700k/cledyu"
LABEL org.opencontainers.image.description="Cledyu frontend (Next.js)"
LABEL org.opencontainers.image.licenses="MIT"
WORKDIR /app
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
COPY --from=builder --chown=nonroot:nonroot /app/.next/standalone ./
COPY --from=builder --chown=nonroot:nonroot /app/.next/static ./.next/static
USER nonroot
EXPOSE 3000
ENV PORT=3000
# distroless nodejs는 ENTRYPOINT=["/nodejs/bin/node"]로 고정되어 있어 CMD에 스크립트만 지정
CMD ["server.js"]
