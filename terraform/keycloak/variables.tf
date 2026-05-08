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
