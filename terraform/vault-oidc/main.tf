terraform {
  required_providers {
    vault = {
      source  = "hashicorp/vault"
      version = "~> 5.0"
    }
  }
}

provider "vault" {
  address         = var.vault_address
  token           = var.vault_root_token
  skip_tls_verify = true
}

# ==========================================
# OIDC Auth Backend (Keycloak)
# ==========================================

resource "vault_jwt_auth_backend" "oidc" {
  path               = "oidc"
  type               = "oidc"
  oidc_discovery_url = "${var.keycloak_url}/realms/home"
  oidc_client_id     = "vault"
  oidc_client_secret = var.vault_oidc_client_secret
  default_role       = "default"
}

resource "vault_jwt_auth_backend_role" "default" {
  backend        = vault_jwt_auth_backend.oidc.path
  role_name      = "default"
  role_type      = "oidc"
  bound_audiences = ["vault"]
  user_claim     = "sub"
  groups_claim   = "groups"
  allowed_redirect_uris = [
    "https://vault.internal.rayq.app/ui/vault/auth/oidc/oidc/callback",
    "https://vault.internal.rayq.app/oidc/callback",
  ]
  token_policies = ["default"]
}

# ==========================================
# Admin group → full access policy
# ==========================================

resource "vault_policy" "admin" {
  name = "admin"

  policy = <<EOT
path "*" {
  capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}
EOT
}

resource "vault_identity_group" "admins" {
  name     = "admins"
  type     = "external"
  policies = [vault_policy.admin.name]
}

# Maps Keycloak "vault-admins" group to Vault "admins" identity group
resource "vault_identity_group_alias" "admins_keycloak" {
  name           = "vault-admins"
  mount_accessor = vault_jwt_auth_backend.oidc.accessor
  canonical_id   = vault_identity_group.admins.id
}
