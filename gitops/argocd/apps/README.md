# ArgoCD App-of-Apps

이 디렉터리 아래의 모든 `*.yaml` 은 ArgoCD `root-apps` Application 이 recursive 하게 sync 한다.

## 규칙
- 파일명: `<영역>-<앱이름>.yaml` (예: `platform-kube-prometheus-stack.yaml`, `data-kafka.yaml`)
- 각 파일은 `kind: Application` (또는 `kind: ApplicationSet`) 하나를 정의
- `spec.syncPolicy.automated.{prune,selfHeal}` 은 초기 단계에서 모두 `true` (merge 즉시 반영)
- Namespace 생성이 필요하면 `syncOptions: [CreateNamespace=true]` 포함
- 리소스 요구가 큰 앱(예: Kafka) 은 별도 Project 로 분리 예정 (Phase 8+)

## 추가 절차
1. 이 디렉터리에 `<영역>-<앱>.yaml` 추가 → PR
2. main 에 merge
3. ArgoCD 가 자동으로 새 Application 을 생성·sync (1-3분 내)
4. ArgoCD UI 에서 health=Healthy, sync=Synced 확인

## 현재 등록된 앱
- (비어 있음 — Phase 7+ 에서 플랫폼/데이터/AI 앱 추가 예정)
