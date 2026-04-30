# Kafka 운영 런북

Strimzi KRaft 모드 Kafka 클러스터(브로커 3대) 운영 절차.

## 선행 조건

- `kubectl get pods -n kafka` 에서 브로커 3대 Running 확인
- ArgoCD `data-kafka-cluster` Application Synced 확인

```bash
kubectl get pods -n kafka
kubectl get kafka cledyu-kafka -n kafka
kubectl get kafkatopic -n kafka
```

---

## 1. 토픽 추가

### 절차

1. `gitops/apps/kafka-cluster/topics/` 에 yaml 파일 생성:

   ```yaml
   apiVersion: kafka.strimzi.io/v1beta2
   kind: KafkaTopic
   metadata:
     name: <토픽-이름>
     namespace: kafka
     labels:
       strimzi.io/cluster: cledyu-kafka
   spec:
     partitions: 6        # 컨슈머 수에 맞게 조정
     replicas: 3
     config:
       retention.ms: "604800000"    # 7일
       min.insync.replicas: "2"
       cleanup.policy: delete
   ```

2. Git push → ArgoCD 자동 sync:

   ```bash
   git add gitops/apps/kafka-cluster/topics/<토픽-이름>.yaml
   git commit -m "feat(data): add <토픽-이름> topic"
   git push
   ```

### 검증

```bash
kubectl get kafkatopic <토픽-이름> -n kafka
```

기대 출력:

```
NAME           CLUSTER         PARTITIONS   REPLICATION FACTOR   READY
<토픽-이름>   cledyu-kafka    6            3                    True
```

### 롤백

```bash
# yaml 파일 삭제 후 push → ArgoCD prune으로 자동 제거
git rm gitops/apps/kafka-cluster/topics/<토픽-이름>.yaml
git commit -m "revert(data): remove <토픽-이름> topic"
git push
```

---

## 2. 토픽 제거

토픽 yaml 파일 삭제 후 push. ArgoCD `prune: true` 설정으로 자동 삭제됨.

> ⚠️ 토픽 삭제 시 데이터도 함께 삭제됨. 컨슈머 연결 여부 먼저 확인.

```bash
# 컨슈머 그룹 확인 (kcat Pod 필요)
kubectl exec -n kafka kcat -- kcat -b cledyu-kafka-kafka-bootstrap.kafka.svc:9092 -L | grep <토픽-이름>
```

---

## 3. 브로커 Scale

### 스케일 업 (예: 3 → 5대)

`gitops/apps/kafka-cluster/kafka-nodepool.yaml` 의 `replicas` 값 수정:

```yaml
spec:
  replicas: 5    # 기존 3에서 변경
```

Git push → ArgoCD sync → Strimzi가 새 브로커 Pod 자동 생성.

### 검증

```bash
kubectl get pods -n kafka -l app.kubernetes.io/name=kafka
kubectl get kafkanodepool -n kafka
```

### 주의사항

- 스케일 다운 시 해당 브로커의 파티션이 다른 브로커로 재분배된 뒤 Pod 종료됨
- 재분배 완료 전 강제 삭제 금지

---

## 4. CA 인증서 갱신 (CA Rotate)
## 임시, 후속 조치 예정

cert-manager가 cluster CA 또는 root CA를 갱신한 경우, `ca-secret-sync` Job을 재실행해야 Strimzi에 새 체인이 반영됨.

> Job은 ArgoCD Sync hook으로만 실행되므로 인증서 갱신 시 자동 반영되지 않음.

### 절차

```bash
# 1. 기존 Job 삭제
kubectl delete job kafka-ca-secret-sync -n kafka

# 2. ArgoCD force sync
# ArgoCD UI → data-kafka-cluster → Sync → Force 체크 → Synchronize
# 또는 CLI:
argocd app sync data-kafka-cluster --force
```

### 검증

```bash
# Job 완료 확인
kubectl get job kafka-ca-secret-sync -n kafka

# 브로커 재시작 확인 (Strimzi가 새 인증서 감지 후 rolling restart)
kubectl get pods -n kafka -w
```

기대 출력:

```
kafka-ca-secret-sync   1/1   Complete   ...
```

---

## 5. 트러블슈팅

### 브로커 Pod가 Pending 상태

```bash
kubectl describe pod <pod-name> -n kafka
```

원인별 해결:

| 원인 | 해결 |
|---|---|
| PVC 미생성 | `kubectl get pvc -n kafka` 확인, Longhorn 상태 점검 |
| anti-affinity 충돌 | 노드 수 부족. 브로커 수 줄이거나 노드 추가 |
| 리소스 부족 | `kubectl describe node \| grep -A 5 "Allocated resources"` 로 가용 리소스 확인 |

### 브로커 Pod가 CrashLoopBackOff

```bash
kubectl logs <pod-name> -n kafka --previous
```

원인별 해결:

| 원인 | 해결 |
|---|---|
| CA Secret 없음 또는 라벨 누락 | `ca-secret-sync` Job 재실행 (4번 절차) |
| PVC 권한 오류 | `fsGroup: 1000` 설정 확인, PVC 재생성 필요할 수 있음 |
| JVM 메모리 부족 | `jvmOptions` `-Xmx` 값 조정 |

### KafkaTopic이 Ready가 안 됨

```bash
kubectl describe kafkatopic <토픽-이름> -n kafka
```

entity-operator (Topic Operator) 로그 확인:

```bash
kubectl logs -n kafka -l app.kubernetes.io/name=entity-operator -c topic-operator
```

### ArgoCD OutOfSync 노이즈
### N-1: data-kafka-cluster.yaml 에 ignoreDifferences 추가 예정 (Kafka / KafkaTopic CR 의 .status drift 노이즈 제거)

Kafka CR `.status` 필드 drift로 인한 OutOfSync는 정상. 실제 변경사항이 없으면 무시.

---

## 참고

- Strimzi 공식 문서: https://strimzi.io/docs/operators/latest/
- 관련 파일:
  - `gitops/apps/kafka-cluster/` — Kafka 클러스터 매니페스트
  - `gitops/apps/kafka-cluster/topics/` — 토픽 yaml
  - `gitops/apps/trust-manager/` — CA 공개키 분배
  - `gitops/apps/kafka-cluster/ca-secret-sync.yaml` — CA Secret 변환 Job
