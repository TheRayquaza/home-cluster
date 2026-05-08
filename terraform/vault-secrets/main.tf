terraform {
  required_providers {
    vault = {
      source  = "hashicorp/vault"
      version = "~> 5.0"
    }
  }
}

provider "vault" {
  address = var.vault_address
  token   = var.vault_root_token
  skip_tls_verify= true
}

# Enable KV v2 secrets engine if not already enabled
resource "vault_mount" "kv" {
  path        = "kv"
  type        = "kv"
  options     = { version = "2" }
  description = "KV Version 2 secret engine mount"
}

# ==========================================
# Cloudflare Secrets
# ==========================================

resource "vault_kv_secret_v2" "cloudflare" {
  mount = vault_mount.kv.path
  name  = "cloudflare"

  data_json = jsonencode({
    CLOUDFLARE_API_TOKEN = var.cloudflare_api_token
    email                = var.cloudflare_email
  })
}

# ==========================================
# Keycloak Secrets
# ==========================================
# Run twice:
#   1st apply (bootstrap): provide admin-password + postgresql-password only
#   2nd apply (post-realm): add client secrets from `terraform output` in terraform/keycloak/

resource "vault_kv_secret_v2" "keycloak" {
  mount = vault_mount.kv.path
  name  = "keycloak"

  data_json = jsonencode({
    admin-password         = var.keycloak_admin_password
    postgresql-password    = var.keycloak_postgresql_password
    argocd-client-secret   = var.keycloak_argocd_client_secret
    vault-client-secret    = var.keycloak_vault_client_secret
    kommande-client-secret = var.keycloak_kommande_client_secret
    games-client-secret    = var.keycloak_games_client_secret
  })
}
