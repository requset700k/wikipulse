# 보안 정책

WikiPulse는 Wikipedia SSE · Reddit OAuth2 · Gemini API · Vault 관리 시크릿을 다루므로
취약점은 **공개 이슈로 먼저 보고하지 말아 주세요**.

## 신고 경로 (우선순위 순)

1. **GitHub Security Advisory** — `Security → Advisories → Report a vulnerability`
   비공개 스레드에서 메인테이너와 직접 논의합니다. (권장)
2. **Discord DM** — `#security` 채널의 **윤승호 (DevSecOps)** 또는 **김용균 (Platform Lead)**
3. 위 경로가 불가한 경우에만 `security@` 도메인 메일 (운영 시점에 공지)

## 신고 시 포함해주세요

- 취약점 유형 / 예상 심각도 (CVSS 가능 시)
- 재현 절차 (환경·커맨드·페이로드)
- 영향 컴포넌트 (K8s 클러스터 / 파이프라인 / API / 대시보드)
- PoC 코드 또는 로그 (민감 정보는 마스킹)

## 대응 SLA

| 심각도 | 초기 응답 | 완화 | 공개 패치 |
|---|---|---|---|
| Critical (RCE / 인증 우회 / 데이터 유출) | 24h | 72h | 14d |
| High | 48h | 7d | 30d |
| Medium / Low | 5 영업일 | 다음 릴리스 | — |

## 지원 범위

- `main` 브랜치 HEAD
- 최신 태그 릴리스 1개
- 이전 릴리스는 best-effort

## 자동 보안 게이트 (참고)

- Trivy (컨테이너 CVE) / gitleaks (시크릿) / Ruff `S` (bandit) / OWASP ZAP (API)
- Kyverno (런타임 정책) / Falco + SIEM (OpenSearch Security Analytics)

## 책임 공개

신고 후 패치가 릴리스되면 신고자 동의 하에 Advisory / 릴리스 노트에 기재합니다.
