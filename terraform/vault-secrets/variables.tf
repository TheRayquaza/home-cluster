variable "vault_address" {
  description = "Vault server address"
  type        = string
  default     = "http://localhost:8200"
}

variable "vault_root_token" {
  description = "Vault root token"
  type        = string
  sensitive   = true
}


# Cloudflare variables
variable "cloudflare_api_token" {
  type      = string
  sensitive = true
}

variable "cloudflare_email" {
  type      = string
  sensitive = true
}

# Keycloak variables — see terraform/keycloak/ for client secret values

variable "keycloak_admin_password" {
  description = "Keycloak admin password (set before first deploy)"
  type        = string
  sensitive   = true
}

variable "keycloak_postgresql_password" {
  description = "PostgreSQL password for Keycloak (set before first deploy)"
  type        = string
  sensitive   = true
}

variable "keycloak_argocd_client_secret" {
  description = "OIDC client secret for ArgoCD (from terraform/keycloak output)"
  type        = string
  sensitive   = true
  default     = ""
}

variable "keycloak_vault_client_secret" {
  description = "OIDC client secret for Vault (from terraform/keycloak output)"
  type        = string
  sensitive   = true
  default     = ""
}

variable "keycloak_kommande_client_secret" {
  description = "OIDC client secret for Kommande (from terraform/keycloak output)"
  type        = string
  sensitive   = true
  default     = ""
}

variable "keycloak_games_client_secret" {
  description = "OIDC client secret for Games (from terraform/keycloak output)"
  type        = string
  sensitive   = true
  default     = ""
}

