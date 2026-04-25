# cledyu CI 검증용 dummy 이미지.
# Reusable build/push workflow (.github/workflows/build-image.yml) 의 첫 사용자 +
# 클러스터에서 GHCR public pull 동작을 검증하기 위한 최소 이미지.
# 신규 서비스 추가 시 docs/RUNBOOK/container-image.md 참고.
FROM alpine:3.20
LABEL org.opencontainers.image.source="https://github.com/requset700k/cledyu"
LABEL org.opencontainers.image.description="cledyu container registry foundation smoke test"
LABEL org.opencontainers.image.licenses="MIT"
CMD ["echo", "cledyu container registry foundation OK"]
