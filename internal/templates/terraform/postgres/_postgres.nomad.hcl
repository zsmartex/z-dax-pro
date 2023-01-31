variable "default_username" {
  type = string
}

variable "default_password" {
  type = string
}

job "postgres" {
  datacenters = ["{{.Config.DC}}"]
  type = "service"

  constraint {
    attribute = "${node.unique.name}"
    value     = "zsmartex-database-server-1"
  }

  group "postgres" {
    count = 1

    network {
      mode = "bridge"

      port "db" {
        static = 5432
      }
    }

    service {
      name = "postgres"
      tags = ["postgresql"]
      port = "db"

      connect {
        sidecar_service {}
      }

      check {
        name     = "alive"
        type     = "tcp"
        interval = "10s"
        timeout  = "2s"
      }
    }

    volume "postgres-data" {
      type = "host"
      source = "data-vol"
    }

    task "postgres" {
      driver = "docker"

      config {
        image        = "debezium/postgres:14"
      }

      env {
        POSTGRES_USER     = var.default_username
        POSTGRES_PASSWORD = var.default_password
      }

      volume_mount {
        volume      = "postgres-data"
        destination = "/postgres"
      }

      logs {
        max_files     = 5
        max_file_size = 15
      }

      resources {
        cpu    = 1000
        memory = 1024
      }
    }
    restart {
      attempts = 10
      interval = "5m"
      delay = "25s"
      mode = "delay"
    }
  }

  update {
    max_parallel = 1
    min_healthy_time = "5s"
    healthy_deadline = "3m"
    auto_revert = false
    canary = 0
  }
}
