resource "keycloak_role" "realm_roles" {
  for_each = local.realm_roles

  realm_id    = keycloak_realm.cledyu.id
  name        = each.key
  description = each.value.description
}
