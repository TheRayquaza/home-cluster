variable "keycloak_url" {
  description = "Keycloak server URL (e.g. https://keycloak.internal.rayq.app)"
  type        = string
  default     = "https://keycloak.internal.rayq.app"
}

variable "keycloak_admin_user" {
  description = "Keycloak admin username"
  type        = string
  default     = "admin"
}

variable "keycloak_admin_password" {
  description = "Keycloak admin password"
  type        = string
  sensitive   = true
}

variable "users" {
  description = "Users to create in the home realm. groups: argocd-admins, vault-admins, kommande-admins, games-admins"
  type = map(object({
    email    = string
    password = string
    groups   = optional(list(string), [])
  }))
  default = {}
}
