variable "nomad_address" {
  type = string
  sensitive = true
}

variable "nomad_ca_file" {
  type = string
}

variable "nomad_cert_file" {
  type = string
}

variable "nomad_key_file" {
  type = string
}

variable "consul_token" {
  type = string
  sensitive = true
}

variable "vault_address" {
  type      = string
  sensitive = true
}

variable "vault_token" {
  type      = string
  sensitive = true
}

# the root username of postgresql
variable "postgres_username" {
  type      = string
  sensitive = true
}

# the root password of postgresql
variable "postgres_password" {
  type      = string
  sensitive = true
}
