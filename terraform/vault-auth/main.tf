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
}

# ==========================================
# Enable Kubernetes Auth Method
# ==========================================

resource "vault_auth_backend" "kubernetes" {
  type = "kubernetes"
  path = "kubernetes"
  
  description = "Kubernetes auth backend for External Secrets Operator"
}

resource "vault_kubernetes_auth_backend_config" "kubernetes" {
  backend            = vault_auth_backend.kubernetes.path
  kubernetes_host    = var.kubernetes_host
  kubernetes_ca_cert = var.kubernetes_ca_cert
  token_reviewer_jwt = var.vault_auth_token
}

# ==========================================
# External Secrets Operator Role
# ==========================================

resource "vault_policy" "external_secrets" {
  name = "external-secrets-policy"

  policy = <<EOT
# Read access to all secrets paths used by ExternalSecrets

# List capability for the secret mount
path "secret/metadata/*" {
  capabilities = ["list"]
}

# Allow reading the actual secret values
path "kv/data/*" {
  capabilities = ["read"]
}

# Allow listing (optional, but helpful for UI/debugging)
path "kv/metadata/*" {
  capabilities = ["list", "read"]
}

EOT
}

resource "vault_kubernetes_auth_backend_role" "flux_system" {
  backend                          = vault_auth_backend.kubernetes.path
  role_name                        = "flux-system"
  bound_service_account_names      = ["vault-auth"]
  bound_service_account_namespaces = ["kube-system"]
  token_ttl                        = 3600
  token_max_ttl                    = 86400
  token_policies                   = [vault_policy.external_secrets.name]
  audience                         = null
}
