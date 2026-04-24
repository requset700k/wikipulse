# 네트워크 설계

## 온프레미스 네트워크 다이어그램

```
┌──────────────────────────────────────────────────────────┐
│ 온프레미스 호스트 (Ubuntu 22.04, 32C/128GB)                │
│                                                          │
│   libvirt NAT 네트워크: cledyubr0 (10.10.0.0/24)          │
│   ├─ cp01  10.10.0.11   (control-plane + worker)         │
│   ├─ cp02  10.10.0.12   (control-plane + worker)         │
│   └─ cp03  10.10.0.13   (control-plane + worker)         │
│                                                          │
│   kube-vip 컨트롤플레인 VIP: 10.10.0.10:6443              │
│   MetalLB L2 풀:             10.10.0.200 ~ 10.10.0.220    │
└──────────────────────────────────────────────────────────┘
         │                                │
         │ Tailscale subnet router (tailnet 100.x.x.x)
         ▼                                ▼
      팀원 노트북                        AWS VPC / GCP DR
```

## IP / CIDR 할당

| 항목                   | 값                        | 비고                                   |
| ---------------------- | ------------------------- | -------------------------------------- |
| 호스트 물리 LAN        | (호스트 기본값 유지)      | 호스트가 NAT 게이트웨이 역할           |
| libvirt bridge         | `cledyubr0`               | NAT, dnsmasq DHCP 비활성(고정 IP 사용) |
| VM 네트워크            | `10.10.0.0/24`            | libvirt NAT 대역                       |
| 게이트웨이             | `10.10.0.1`               | libvirt가 자동 할당(호스트 인터페이스) |
| cp01 / cp02 / cp03     | `10.10.0.11/.12/.13`      | 고정 IP, cloud-init에서 주입           |
| 컨트롤플레인 API VIP   | `10.10.0.10:6443`         | kube-vip (ARP 모드)                    |
| MetalLB LoadBalancer   | `10.10.0.200-10.10.0.220` | L2Advertisement                        |
| Pod CIDR               | `10.42.0.0/16`            | Cilium cluster-pool IPAM (계획서)      |
| Service CIDR           | `10.43.0.0/16`            | kube-apiserver `--service-cidr`        |
| Tailscale tailnet      | `100.64.0.0/10`           | 자동 할당, MagicDNS 사용               |

## 인바운드/아웃바운드 원칙
- **북-남 트래픽**: 외부 유저는 CloudFront + WAF → Tailscale Funnel → 온프렘 Kong Gateway → Cilium 서비스 메시 → KubeVirt Lab VM 또는 EC2 오버플로우 Lab VM.
- **컨트롤플레인 접근**: Tailscale을 통해서만 `10.10.0.10:6443` 노출. ACL로 김용균/윤승호 제한.
- **크로스 클라우드**: 온프렘 ↔ AWS는 Tailscale subnet router로 연결. Lab VM이 EC2로 오버플로우될 때도 동일 tailnet 경로 사용. 데이터 레이어(S3/BigQuery)는 HTTPS 직접 접근.
- **수강생 VM 격리**: KubeVirt NetworkAttachmentDefinition(온프렘) + VPC Subnet/Security Group(EC2)로 VM 간 통신 차단, egress는 허용 도메인만.
- **DNS**: 내부는 MagicDNS(`*.tailnet.ts.net`), 외부용 Route 53은 퍼블릭 도메인 발급 이후 추가.

## 포트 체계

| 포트          | 용도                           | 노출 대상                 |
| ------------- | ------------------------------ | ------------------------- |
| 22            | SSH                            | Tailscale만               |
| 6443          | kube-apiserver                 | Tailscale만 (ACL)         |
| 2379-2380     | etcd (peer/client)             | VM 내부만                 |
| 10250         | kubelet                        | VM 내부만                 |
| 4240          | Cilium health                  | VM 내부만                 |
| 80/443        | Ingress(LB)                    | LAN + Tailscale           |
| 30000-32767   | NodePort (사용 지양)           | 내부                      |
