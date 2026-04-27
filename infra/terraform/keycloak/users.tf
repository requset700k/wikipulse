resource "keycloak_user" "team_members" {
  for_each = var.team_members

  realm_id = keycloak_realm.cledyu.id
  username = each.value.username
  enabled  = each.value.enabled

  email          = each.value.email
  first_name     = each.value.first_name
  last_name      = each.value.last_name
  email_verified = each.value.email_verified

  required_actions = ["UPDATE_PASSWORD"]

  initial_password {
    value     = var.team_member_initial_passwords[each.key]
    temporary = each.value.temporary_password
  }

  # 첫 로그인 시 사용자가 비번 변경하면 required_actions 가 비워짐 + initial_password
  # 도 무효화된다. terraform apply 가 매번 이 둘을 reset 시키지 않도록 ignore.
  # 비번 reset 이 정말 필요하면 admin console / kcadm.sh 로 직접 처리.
  lifecycle {
    ignore_changes = [
      required_actions,
      initial_password,
    ]
  }
}

resource "keycloak_user_groups" "team_members" {
  for_each = var.team_members

  realm_id = keycloak_realm.cledyu.id
  user_id  = keycloak_user.team_members[each.key].id

  group_ids = [
    for group_name in each.value.groups : keycloak_group.groups[group_name].id
  ]
}
