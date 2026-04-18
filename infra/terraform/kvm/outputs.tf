output "nodes" {
  description = "노드 이름 → 고정 IP"
  value       = { for k, v in var.nodes : k => split("/", v.ip)[0] }
}

output "control_plane_ips" {
  description = "컨트롤플레인 IP 목록 (kubeadm init/join 대상)"
  value       = [for k, v in var.nodes : split("/", v.ip)[0] if v.role == "control-plane"]
}

output "ansible_inventory_snippet" {
  description = "ansible/inventory.yml 에 붙여넣을 k8s_nodes 정의"
  value = yamlencode({
    all = {
      children = {
        k8s_nodes = {
          children = {
            control_plane = {
              hosts = {
                for k, v in var.nodes : k => { ansible_host = split("/", v.ip)[0] } if v.role == "control-plane"
              }
            }
          }
        }
      }
    }
  })
}
