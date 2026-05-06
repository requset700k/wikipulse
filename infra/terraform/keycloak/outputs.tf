output "realm_id" {
  description = "Terraform이 관리하는 Cledyu realm ID."
  value       = keycloak_realm.cledyu.id
}

output "realm_roles" {
  description = "Terraform이 관리하는 realm role 이름 목록."
  value       = keys(keycloak_role.realm_roles)
}

output "groups" {
  description = "Terraform이 관리하는 group 이름 목록."
  value       = keys(keycloak_group.groups)
}

output "oidc_client_ids" {
  description = "Terraform이 관리하는 OIDC client ID 목록."
  value       = keys(keycloak_openid_client.clients)
}

output "confidential_client_ids" {
  description = "보안 채널로 credential 전달이 필요한 confidential OIDC client 목록."
  value       = local.confidential_client_ids
  sensitive   = true
}
