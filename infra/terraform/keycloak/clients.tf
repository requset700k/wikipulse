resource "keycloak_openid_client" "clients" {
  for_each = var.oidc_clients

  realm_id  = keycloak_realm.cledyu.id
  client_id = each.key
  name      = each.value.name
  enabled   = true

  access_type                  = upper(each.value.access_type)
  standard_flow_enabled        = each.value.standard_flow_enabled
  direct_access_grants_enabled = each.value.direct_access_grants_enabled
  implicit_flow_enabled        = each.value.implicit_flow_enabled
  service_accounts_enabled     = each.value.service_accounts_enabled

  pkce_code_challenge_method = try(each.value.pkce_code_challenge_method, null)

  valid_redirect_uris             = try(each.value.valid_redirect_uris, [])
  valid_post_logout_redirect_uris = try(each.value.valid_post_logout_redirect_uris, [])
  web_origins                     = try(each.value.web_origins, [])

  root_url  = try(each.value.root_url, null)
  base_url  = try(each.value.base_url, null)
  admin_url = try(each.value.admin_url, null)

  client_secret = try(var.oidc_client_secrets[each.key], null)
}

resource "keycloak_openid_group_membership_protocol_mapper" "groups" {
  for_each = {
    for client_id, client in var.oidc_clients : client_id => client
    if upper(client.access_type) != "BEARER-ONLY"
  }

  realm_id  = keycloak_realm.cledyu.id
  client_id = keycloak_openid_client.clients[each.key].id
  name      = "groups"

  claim_name = "groups"
  full_path  = false

  add_to_access_token = true
  add_to_id_token     = true
  add_to_userinfo     = true
}
