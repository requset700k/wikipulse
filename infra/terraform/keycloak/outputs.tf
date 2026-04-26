output "realm_id" {
  description = "Managed Cledyu realm ID."
  value       = keycloak_realm.cledyu.id
}

output "realm_roles" {
  description = "Managed realm role names."
  value       = keys(keycloak_role.realm_roles)
}

output "groups" {
  description = "Managed group names."
  value       = keys(keycloak_group.groups)
}

output "oidc_client_ids" {
  description = "Managed OIDC client IDs."
  value       = keys(keycloak_openid_client.clients)
}

output "confidential_client_ids" {
  description = "Confidential OIDC clients that require secure credential delivery."
  value       = local.confidential_client_ids
  sensitive   = true
}
