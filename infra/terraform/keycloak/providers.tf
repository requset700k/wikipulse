provider "keycloak" {
  client_id = var.keycloak_admin_client_id
  username  = var.keycloak_admin_username
  password  = var.keycloak_admin_password
  realm     = var.keycloak_admin_realm
  url       = var.keycloak_url

  tls_insecure_skip_verify = var.keycloak_tls_insecure_skip_verify
}
