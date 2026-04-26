resource "keycloak_group" "groups" {
  for_each = local.groups

  realm_id = keycloak_realm.cledyu.id
  name     = each.key
}

resource "keycloak_group_roles" "group_roles" {
  for_each = local.groups

  realm_id = keycloak_realm.cledyu.id
  group_id = keycloak_group.groups[each.key].id

  role_ids = [
    for role_name in each.value.realm_roles : keycloak_role.realm_roles[role_name].id
  ]
}
