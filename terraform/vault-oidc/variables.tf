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

variable "keycloak_url" {
  description = "Keycloak server URL (e.g. https://keycloak.internal.rayq.app)"
  type        = string
  default     = "https://keycloak.internal.rayq.app"
}

variable "vault_oidc_client_secret" {
  description = "Keycloak client secret for the vault OIDC client"
  type        = string
  sensitive   = true
}
