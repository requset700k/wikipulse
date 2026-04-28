# Master realm super-admin 영구화 (ADR-0001 §13).
#
# `master` realm 은 Keycloak Operator 가 자동 생성하는 운영 realm. 본 모듈은
# 기존 master realm 을 data source 로 가져와서 그 안에 김용균 / 윤승호 user 를
# 추가하고 admin role 을 부여한다.
#
# super-admin 이력:
#   - kylekim 김용균 (Phase 4 진행 중 GUI 필요로 부여)
#   - yunseungho 윤승호 (break-glass 백업)
#
# 비번 정책:
#   - Initial password 는 1Password "Cledyu/keycloak-master-admin-{username}"
#     항목에서 secrets-manager 로 주입 (terraform.tfvars 의 master_admin_initial_passwords).
#   - lifecycle.ignore_changes 로 첫 변경 후 reset 되지 않게 보장.

data "keycloak_realm" "master" {
  realm = "master"
}

data "keycloak_role" "master_admin" {
  realm_id = data.keycloak_realm.master.id
  name     = "admin"
}

# Keycloak 이 모든 user 에게 자동 부여하는 default composite role.
# keycloak_user_roles 의 exhaustive=true 가 default 라 명시 안 하면 제거됨.
# 제거 시 admin console 의 account 메뉴 등 기본 access 차단.
data "keycloak_role" "master_default_roles" {
  realm_id = data.keycloak_realm.master.id
  name     = "default-roles-master"
}

resource "keycloak_user" "master_super_admins" {
  for_each = var.master_super_admins

  realm_id = data.keycloak_realm.master.id
  username = each.value.username
  enabled  = true

  email          = each.value.email
  first_name     = each.value.first_name
  last_name      = each.value.last_name
  email_verified = true

  required_actions = ["UPDATE_PASSWORD"]

  initial_password {
    value     = var.master_admin_initial_passwords[each.key]
    temporary = true
  }

  # 첫 로그인 후 사용자가 비번 변경하면 required_actions 가 비워진다.
  # terraform apply 가 매번 reset 하지 않도록 ignore (cledyu realm 의 users.tf 와 동일 패턴).
  lifecycle {
    ignore_changes = [
      required_actions,
      initial_password,
    ]
  }
}

resource "keycloak_user_roles" "master_super_admins" {
  for_each = var.master_super_admins

  realm_id = data.keycloak_realm.master.id
  user_id  = keycloak_user.master_super_admins[each.key].id

  role_ids = [
    data.keycloak_role.master_admin.id,
    data.keycloak_role.master_default_roles.id,
  ]
}
