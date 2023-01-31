terraform {
  required_providers {
    vault = {
      source  = "hashicorp/vault"
      version = "3.12.0"
    }
    nomad = {
      source  = "hashicorp/nomad"
      version = "1.4.18"
    }
    postgresql = {
      source  = "cyrilgdn/postgresql"
      version = "1.18.0"
    }
  }

  backend "consul" {
    address      = "consul.service.consul"
    scheme       = "https"
    path         = "terraform/postgres"
  }
}

provider "postgresql" {
  host     = "database.consul.service"
  port     = 5432
  username = var.postgres_username
  password = var.postgres_password
  sslmode  = "disable"
}
