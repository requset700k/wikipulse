variable "libvirt_uri" {
  description = "libvirt URI. 로컬 실행시 qemu:///system, 원격시 qemu+ssh://ubuntu@<호스트>/system"
  type        = string
}

variable "ssh_authorized_key" {
  description = "cloud-init 으로 주입할 SSH 공개키 내용"
  type        = string
}
