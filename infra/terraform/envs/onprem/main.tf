terraform {
  required_version = ">= 1.5.0"
  # 로컬 state 로 시작. Phase 10 에서 S3 백엔드로 이관.
}

module "kvm" {
  source = "../../kvm"

  libvirt_uri        = var.libvirt_uri
  ssh_authorized_key = var.ssh_authorized_key

  # 기본값을 덮어쓰고 싶으면 여기에서 추가
  # nodes = { ... }
}

output "nodes" { value = module.kvm.nodes }
output "control_plane_ips" { value = module.kvm.control_plane_ips }
output "ansible_inventory" { value = module.kvm.ansible_inventory_snippet }
