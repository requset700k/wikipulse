# onprem 환경 — KVM 3노드 프로비저닝

Phase 1 결과물. Ansible 로 libvirt + cledyubr0 네트워크가 준비된 **이후**에 실행.

## 사전 요구사항
1. `make host-prep HOST=<호스트-IP>` 완료 (libvirt + cledyubr0 + base 이미지 준비됨)
2. 로컬에 Terraform >= 1.6 설치
3. 호스트에 SSH 키 기반 접속 가능 (원격 libvirt 드라이버)

## 실행
```bash
cp terraform.tfvars.example terraform.tfvars
$EDITOR terraform.tfvars        # libvirt_uri, ssh_authorized_key 채우기

terraform -chdir=. init
terraform -chdir=. plan
terraform -chdir=. apply
```

## 결과 확인
```bash
terraform output nodes
terraform output ansible_inventory > ../../../ansible/inventory.yml
```

## 파기
```bash
terraform destroy    # VM/볼륨/cloud-init ISO 모두 삭제
```
