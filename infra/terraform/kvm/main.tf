provider "libvirt" {
  uri = var.libvirt_uri
}

# 전용 스토리지 풀 (default 풀과 충돌 방지)
resource "libvirt_pool" "wp" {
  name = "${var.cluster_name}-pool"
  type = "dir"

  target {
    path = var.images_pool_path
  }
}

# 베이스 이미지. provider가 URL에서 받아 pool에 업로드 (1회).
resource "libvirt_volume" "base" {
  name   = "${var.cluster_name}-base.qcow2"
  pool   = libvirt_pool.wp.name
  source = var.base_image_source
  format = "qcow2"
}

# 노드별 루트 디스크(베이스에서 CoW)
resource "libvirt_volume" "root" {
  for_each = var.nodes

  name           = "${var.cluster_name}-${each.key}.qcow2"
  pool           = libvirt_pool.wp.name
  base_volume_id = libvirt_volume.base.id
  size           = each.value.disk_gb * 1024 * 1024 * 1024
  format         = "qcow2"
}

resource "libvirt_cloudinit_disk" "ci" {
  for_each = var.nodes

  name = "${var.cluster_name}-${each.key}-ci.iso"
  pool = libvirt_pool.wp.name

  # NoCloud datasource 는 meta-data 에 instance-id 가 없으면 user-data 전체를 무시함.
  # dmacvicar/libvirt 0.9.x 부터 meta_data 를 명시적으로 주지 않으면 비어있는 상태로 생성됨.
  meta_data = "instance-id: ${each.key}\nlocal-hostname: ${each.key}\n"

  user_data = templatefile("${path.module}/cloud-init/user-data.yaml.tftpl", {
    hostname           = each.key
    fqdn               = "${each.key}.${var.cluster_name}.local"
    ssh_authorized_key = var.ssh_authorized_key
  })

  network_config = templatefile("${path.module}/cloud-init/network-config.yaml.tftpl", {
    ip          = each.value.ip
    gateway     = var.gateway
    dns_servers = join(",", var.dns_servers)
  })
}

resource "libvirt_domain" "node" {
  for_each = var.nodes

  name      = "${var.cluster_name}-${each.key}"
  memory    = each.value.memory
  vcpu      = each.value.cpu
  cloudinit = libvirt_cloudinit_disk.ci[each.key].id
  autostart = true

  cpu {
    mode = "host-passthrough"
  }

  network_interface {
    network_name   = var.network_name
    wait_for_lease = false
    hostname       = each.key
  }

  disk {
    volume_id = libvirt_volume.root[each.key].id
  }

  console {
    type        = "pty"
    target_type = "serial"
    target_port = "0"
  }

  graphics {
    type        = "vnc"
    listen_type = "address"
    autoport    = true
  }
}
