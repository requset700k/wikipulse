variable "keycloak_url" {
  description = "Base URL for the Keycloak deployment."
  type        = string
  default     = "https://keycloak.cledyu.local"
}

variable "keycloak_admin_realm" {
  description = "Realm used to authenticate the Terraform provider."
  type        = string
  default     = "master"
}

variable "keycloak_admin_client_id" {
  description = "Admin client ID used by the Terraform provider."
  type        = string
  default     = "admin-cli"
}

variable "keycloak_admin_username" {
  description = "Keycloak admin username for Terraform."
  type        = string
  sensitive   = true
}

variable "keycloak_admin_password" {
  description = "Keycloak admin password for Terraform."
  type        = string
  sensitive   = true
}

variable "keycloak_tls_insecure_skip_verify" {
  description = "Skip TLS verification for the Keycloak provider. Keep false after Cledyu Root CA is trusted."
  type        = bool
  default     = false
}

variable "realm_name" {
  description = "Business realm name for Cledyu."
  type        = string
  default     = "cledyu"
}

variable "realm_display_name" {
  description = "Human-readable display name for the Cledyu realm."
  type        = string
  default     = "Cledyu"
}

variable "oidc_clients" {
  description = "OIDC clients managed in the Cledyu realm."
  type = map(object({
    name                            = string
    access_type                     = string
    standard_flow_enabled           = bool
    direct_access_grants_enabled    = bool
    implicit_flow_enabled           = bool
    service_accounts_enabled        = bool
    pkce_code_challenge_method      = optional(string)
    valid_redirect_uris             = optional(list(string), [])
    valid_post_logout_redirect_uris = optional(list(string), [])
    web_origins                     = optional(list(string), [])
    root_url                        = optional(string)
    base_url                        = optional(string)
    admin_url                       = optional(string)
  }))
}

variable "oidc_client_secrets" {
  description = "Client secrets for confidential OIDC clients. Store real values in a secure tfvars source."
  type        = map(string)
  default     = {}
  sensitive   = true
}

variable "team_members" {
  description = "Team member users to create and map to groups."
  type = map(object({
    username           = string
    email              = string
    first_name         = string
    last_name          = string
    groups             = list(string)
    temporary_password = optional(bool, true)
    enabled            = optional(bool, true)
    email_verified     = optional(bool, true)
  }))
}

variable "team_member_initial_passwords" {
  description = "Temporary bootstrap passwords for team members. Store real values in a secure tfvars source."
  type        = map(string)
  sensitive   = true
}

variable "master_super_admins" {
  description = "Master realm super-admin users (ADR-0001 §13). lifecycle.ignore_changes 로 비번 reset 강제 방지."
  type = map(object({
    username   = string
    email      = string
    first_name = string
    last_name  = string
  }))
  default = {}
}

variable "master_admin_initial_passwords" {
  description = "Master realm super-admin 의 초기 임시 비번. 1Password 보관 후 첫 로그인 시 변경 강제."
  type        = map(string)
  default     = {}
  sensitive   = true
}
