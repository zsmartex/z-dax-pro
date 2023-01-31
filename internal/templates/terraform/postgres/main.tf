resource "nomad_job" "name" {
  jobspec = file("${path.module}/_postgres.nomad.hcl")
  hcl2 {
    enabled = true

    default_user = var.postgres_username
    default_password = var.postgres_password
  }
}

resource "vault_mount" "postgres" {
  path = "postgres"
  type = "database"
}

resource "vault_database_secret_backend_connection" "peatio" {
  backend       = vault_mount.db.path
  name          = "peatio"
  allowed_roles = [
    "peatio",
    "zsmartex"
  ]

  postgresql {
    connection_url = "postgres://{{username}}:{{password}}@database.service.consul:5432/peatio"
    username       = var.postgres_username
    password       = var.postgres_password
  }
}

resource "postgresql_database" "peatio" {
  name              = "peatio"
  connection_limit  = -1
  allow_connections = true
}

resource "postgresql_role" "static_role_peatio" {
  name            = "zsmartex-peatio"
  login           = true
  create_database = false

  lifecycle {
    ignore_changes = [
      password
    ]
  }

  depends_on = [postgresql_database.peatio]
}

resource "vault_database_secret_backend_role" "peatio" {
  backend             = vault_mount.db.path
  name                = "peatio"
  db_name             = vault_database_secret_backend_connection.postgres.name
  creation_statements = ["CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}';"]
}

resource "vault_database_secret_backend_static_role" "peatio" {
  backend             = vault_mount.db.path
  name                = postgresql_role.static_role_peatio.name
  db_name             = vault_database_secret_backend_connection.peatio.name
  username            = postgresql_role.static_role_peatio.name
  rotation_period     = "3600"
  rotation_statements = ["ALTER USER \"{{name}}\" WITH PASSWORD '{{password}}';"]
}

#=========Barong database===========#

resource "vault_database_secret_backend_connection" "barong" {
  backend       = vault_mount.db.path
  name          = "barong"
  allowed_roles = [
    "barong",
    "zsmartex"
  ]

  postgresql {
    connection_url = "postgres://{{username}}:password@database.service.consul:5432/barong"
  }
}

resource "postgresql_database" "barong" {
  name              = "barong"
  connection_limit  = -1
  allow_connections = true
}

resource "postgresql_role" "static_role_barong" {
  name            = "zsmartex-barong"
  login           = true
  create_database = false

  lifecycle {
    ignore_changes = [
      password
    ]
  }

  depends_on = [postgresql_database.barong]
}

resource "vault_database_secret_backend_role" "peatio" {
  backend             = vault_mount.db.path
  name                = "peatio"
  db_name             = vault_database_secret_backend_connection.postgres.name
  creation_statements = ["CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}';"]
}

resource "vault_database_secret_backend_static_role" "static-role-peatio" {
  backend             = vault_mount.db.path
  name                = postgresql_role.static_role_barong.name
  db_name             = vault_database_secret_backend_connection.peatio.name
  username            = postgresql_role.static_role_barong.name
  rotation_period     = "3600"
  rotation_statements = ["ALTER USER \"{{name}}\" WITH PASSWORD '{{password}}';"]
}
