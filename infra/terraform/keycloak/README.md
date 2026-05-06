# Cledyu Keycloak RBAC

Terraform stack for the Cledyu Keycloak realm, OIDC clients, realm roles,
groups, and team member bootstrap users.

## Secret Handling

Do not commit real `terraform.tfvars`, `oidc_client_secrets`, or
`team_member_initial_passwords`. Use 1Password or another approved secret
channel and keep state in a secured backend before applying in
production-like environments.

## Usage

```bash
cd infra/terraform/keycloak
cp terraform.tfvars.example terraform.tfvars
$EDITOR terraform.tfvars
terraform init
terraform plan
terraform apply
```

<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
| ---- | ------- |
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.5.0 |
| <a name="requirement_keycloak"></a> [keycloak](#requirement\_keycloak) | ~> 4.4 |

## Providers

| Name | Version |
| ---- | ------- |
| <a name="provider_keycloak"></a> [keycloak](#provider\_keycloak) | ~> 4.4 |

## Modules

No modules.

## Resources

| Name | Type |
| ---- | ---- |
| [keycloak_group.groups](https://registry.terraform.io/providers/mrparkers/keycloak/latest/docs/resources/group) | resource |
| [keycloak_group_roles.group_roles](https://registry.terraform.io/providers/mrparkers/keycloak/latest/docs/resources/group_roles) | resource |
| [keycloak_openid_client.clients](https://registry.terraform.io/providers/mrparkers/keycloak/latest/docs/resources/openid_client) | resource |
| [keycloak_openid_group_membership_protocol_mapper.groups](https://registry.terraform.io/providers/mrparkers/keycloak/latest/docs/resources/openid_group_membership_protocol_mapper) | resource |
| [keycloak_realm.cledyu](https://registry.terraform.io/providers/mrparkers/keycloak/latest/docs/resources/realm) | resource |
| [keycloak_realm_events.cledyu](https://registry.terraform.io/providers/mrparkers/keycloak/latest/docs/resources/realm_events) | resource |
| [keycloak_role.realm_roles](https://registry.terraform.io/providers/mrparkers/keycloak/latest/docs/resources/role) | resource |
| [keycloak_user.master_super_admins](https://registry.terraform.io/providers/mrparkers/keycloak/latest/docs/resources/user) | resource |
| [keycloak_user.team_members](https://registry.terraform.io/providers/mrparkers/keycloak/latest/docs/resources/user) | resource |
| [keycloak_user_groups.team_members](https://registry.terraform.io/providers/mrparkers/keycloak/latest/docs/resources/user_groups) | resource |
| [keycloak_user_roles.master_super_admins](https://registry.terraform.io/providers/mrparkers/keycloak/latest/docs/resources/user_roles) | resource |
| [keycloak_realm.master](https://registry.terraform.io/providers/mrparkers/keycloak/latest/docs/data-sources/realm) | data source |
| [keycloak_role.master_admin](https://registry.terraform.io/providers/mrparkers/keycloak/latest/docs/data-sources/role) | data source |
| [keycloak_role.master_default_roles](https://registry.terraform.io/providers/mrparkers/keycloak/latest/docs/data-sources/role) | data source |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| <a name="input_keycloak_admin_password"></a> [keycloak\_admin\_password](#input\_keycloak\_admin\_password) | Keycloak admin password for Terraform. | `string` | n/a | yes |
| <a name="input_keycloak_admin_username"></a> [keycloak\_admin\_username](#input\_keycloak\_admin\_username) | Keycloak admin username for Terraform. | `string` | n/a | yes |
| <a name="input_oidc_clients"></a> [oidc\_clients](#input\_oidc\_clients) | OIDC clients managed in the Cledyu realm. | <pre>map(object({<br/>    name                            = string<br/>    access_type                     = string<br/>    standard_flow_enabled           = bool<br/>    direct_access_grants_enabled    = bool<br/>    implicit_flow_enabled           = bool<br/>    service_accounts_enabled        = bool<br/>    pkce_code_challenge_method      = optional(string)<br/>    valid_redirect_uris             = optional(list(string), [])<br/>    valid_post_logout_redirect_uris = optional(list(string), [])<br/>    web_origins                     = optional(list(string), [])<br/>    root_url                        = optional(string)<br/>    base_url                        = optional(string)<br/>    admin_url                       = optional(string)<br/>  }))</pre> | n/a | yes |
| <a name="input_team_member_initial_passwords"></a> [team\_member\_initial\_passwords](#input\_team\_member\_initial\_passwords) | Temporary bootstrap passwords for team members. Store real values in a secure tfvars source. | `map(string)` | n/a | yes |
| <a name="input_team_members"></a> [team\_members](#input\_team\_members) | Team member users to create and map to groups. | <pre>map(object({<br/>    username           = string<br/>    email              = string<br/>    first_name         = string<br/>    last_name          = string<br/>    groups             = list(string)<br/>    temporary_password = optional(bool, true)<br/>    enabled            = optional(bool, true)<br/>    email_verified     = optional(bool, true)<br/>  }))</pre> | n/a | yes |
| <a name="input_keycloak_admin_client_id"></a> [keycloak\_admin\_client\_id](#input\_keycloak\_admin\_client\_id) | Admin client ID used by the Terraform provider. | `string` | `"admin-cli"` | no |
| <a name="input_keycloak_admin_realm"></a> [keycloak\_admin\_realm](#input\_keycloak\_admin\_realm) | Realm used to authenticate the Terraform provider. | `string` | `"master"` | no |
| <a name="input_keycloak_tls_insecure_skip_verify"></a> [keycloak\_tls\_insecure\_skip\_verify](#input\_keycloak\_tls\_insecure\_skip\_verify) | Skip TLS verification for the Keycloak provider. Keep false after Cledyu Root CA is trusted. | `bool` | `false` | no |
| <a name="input_keycloak_url"></a> [keycloak\_url](#input\_keycloak\_url) | Base URL for the Keycloak deployment. | `string` | `"https://keycloak.cledyu.local"` | no |
| <a name="input_master_admin_initial_passwords"></a> [master\_admin\_initial\_passwords](#input\_master\_admin\_initial\_passwords) | Master realm super-admin 의 초기 임시 비번. 1Password 보관 후 첫 로그인 시 변경 강제. | `map(string)` | `{}` | no |
| <a name="input_master_super_admins"></a> [master\_super\_admins](#input\_master\_super\_admins) | Master realm super-admin users (ADR-0001 §13). lifecycle.ignore\_changes 로 비번 reset 강제 방지. | <pre>map(object({<br/>    username   = string<br/>    email      = string<br/>    first_name = string<br/>    last_name  = string<br/>  }))</pre> | `{}` | no |
| <a name="input_oidc_client_secrets"></a> [oidc\_client\_secrets](#input\_oidc\_client\_secrets) | Client secrets for confidential OIDC clients. Store real values in a secure tfvars source. | `map(string)` | `{}` | no |
| <a name="input_realm_display_name"></a> [realm\_display\_name](#input\_realm\_display\_name) | Human-readable display name for the Cledyu realm. | `string` | `"Cledyu"` | no |
| <a name="input_realm_name"></a> [realm\_name](#input\_realm\_name) | Business realm name for Cledyu. | `string` | `"cledyu"` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| <a name="output_confidential_client_ids"></a> [confidential\_client\_ids](#output\_confidential\_client\_ids) | Confidential OIDC clients that require secure credential delivery. |
| <a name="output_groups"></a> [groups](#output\_groups) | Managed group names. |
| <a name="output_oidc_client_ids"></a> [oidc\_client\_ids](#output\_oidc\_client\_ids) | Managed OIDC client IDs. |
| <a name="output_realm_id"></a> [realm\_id](#output\_realm\_id) | Managed Cledyu realm ID. |
| <a name="output_realm_roles"></a> [realm\_roles](#output\_realm\_roles) | Managed realm role names. |
<!-- END_TF_DOCS -->
