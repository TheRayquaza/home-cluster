variable "cloudflare_api_key" {
  type      = string
  sensitive = true
}

variable "cloudflare_email" {
  type = string
}

variable "zone_id" {
  type = string
}

variable "dns_records" {
  description = "A map of DNS records to create"
  type = map(object({
    name    = string
    content   = string
    type    = string
    proxied = optional(bool, false)
    ttl     = optional(number, 3600)
  }))
}
