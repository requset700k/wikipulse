resource "keycloak_realm" "cledyu" {
  realm   = var.realm_name
  enabled = true

  display_name = var.realm_display_name

  registration_allowed     = false
  remember_me              = false
  reset_password_allowed   = true
  login_with_email_allowed = true
  duplicate_emails_allowed = false
  edit_username_allowed    = false
  revoke_refresh_token     = true

  access_token_lifespan    = "15m"
  sso_session_idle_timeout = "30m"
  sso_session_max_lifespan = "8h"
}

resource "keycloak_realm_events" "cledyu" {
  realm_id = keycloak_realm.cledyu.id

  events_enabled       = true
  admin_events_enabled = true

  events_listeners = ["jboss-logging"]

  enabled_event_types = [
    "LOGIN",
    "LOGIN_ERROR",
    "LOGOUT",
    "REGISTER",
    "UPDATE_PASSWORD",
    "CLIENT_LOGIN",
    "CLIENT_LOGIN_ERROR",
  ]
}
