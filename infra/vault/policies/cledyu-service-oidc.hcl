path "cledyu/data/oidc/web" {
  capabilities = ["read"]
}

path "cledyu/data/oidc/api" {
  capabilities = ["read"]
}

path "cledyu/data/oidc/tutor" {
  capabilities = ["read"]
}

path "cledyu/metadata/oidc/*" {
  capabilities = ["read", "list"]
}
