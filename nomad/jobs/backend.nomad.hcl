job "backend" {
  datacenters = ["dc1"]

  group "demo" {
    count = 3

    network {
      port  "http"{
        to = 80
      }
    }

    service {
      name = "backend"
      port = "http"

      tags = [
        "traefik.enable=true",
        "traefik.http.routers.http.rule=Path(`/`)",
      ]

      check {
        type     = "http"
        path     = "/"
        interval = "2s"
        timeout  = "2s"
      }
    }

    task "server" {
      driver = "docker"

      config {
        image = "nginx"
        ports = ["http"]
      }
    }
  }
}