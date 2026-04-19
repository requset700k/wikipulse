variable "libvirt_uri" {
  description = "libvirt 연결 URI. 로컬 실행: qemu:///system, 원격: qemu+ssh://ubuntu@<host>/system"
  type        = string
  default     = "qemu:///system"
}

variable "network_name" {
  description = "사용할 libvirt 네트워크 이름(Ansible이 wpbr0로 생성)"
  type        = string
  default     = "wpbr0"
}

variable "base_image_source" {
  description = "우분투 22.04 클라우드 이미지 소스(URL 또는 로컬 경로). provider가 받아서 pool에 업로드."
  type        = string
  default     = "https://cloud-images.ubuntu.com/releases/22.04/release/ubuntu-22.04-server-cloudimg-amd64.img"
}

variable "images_pool_path" {
  description = "VM 디스크가 저장될 디렉토리 (default libvirt 풀과 충돌 방지를 위해 하위 디렉토리 사용)"
  type        = string
  default     = "/var/lib/libvirt/images/wikipulse"
}

variable "ssh_authorized_key" {
  description = "cloud-init 으로 주입할 SSH 공개키 (예: file(\"~/.ssh/id_ed25519.pub\"))"
  type        = string
  sensitive   = false
}

variable "cluster_name" {
  description = "클러스터 이름 prefix"
  type        = string
  default     = "wikipulse"
}

variable "dns_servers" {
  description = "VM이 사용할 DNS 서버"
  type        = list(string)
  default     = ["1.1.1.1", "8.8.8.8"]
}

variable "gateway" {
  description = "VM 네트워크 기본 게이트웨이"
  type        = string
  default     = "10.10.0.1"
}

variable "nodes" {
  description = "K8s 노드 정의. 계획서 기준 3노드 HA(컨트롤플레인+워커 겸용)"
  type = map(object({
    role    = string # "control-plane" (Phase 2에서 taint 해제하여 워커 겸용)
    ip      = string
    cpu     = number
    memory  = number # MB
    disk_gb = number
  }))

  # 호스트 실측: AMD Ryzen 9 7950X (16C/32T), 128GB, 1.8TB NVMe (1.7TB free). 이 프로젝트 전용.
  # 합계: 48 vCPU (32논리코어 대비 1.5x, Zen4 SMT 효율↑) / 96 GB(1:1, 호스트 여유 32GB) / 1.5 TB(thin)
  default = {
    cp01 = { role = "control-plane", ip = "10.10.0.11/24", cpu = 16, memory = 32768, disk_gb = 500 }
    cp02 = { role = "control-plane", ip = "10.10.0.12/24", cpu = 16, memory = 32768, disk_gb = 500 }
    cp03 = { role = "control-plane", ip = "10.10.0.13/24", cpu = 16, memory = 32768, disk_gb = 500 }
  }
}
