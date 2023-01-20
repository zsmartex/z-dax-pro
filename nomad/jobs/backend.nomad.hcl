# Simple example to deploy traefik with consul connect enabled.
# For simplicity the job includes traefik as well as the backend service.
# Please note that traefik currently only supports connect for HTTP.
job "backend" {
  datacenters = ["dc1"]

  group "backend" {
    count = 2

    network {
      mode = "bridge"
    }

    service {
      name = "whoami"
      port = 80
      tags = [
        "traefik.enable=true",
        "traefik.http.routers.whoami.rule=Host(`whoami.example.com`)"
      ]

      connect {
        sidecar_service {}
      }
    }

    # Note: For increased security the service should only listen on localhost
    # Otherwise it could be reachable from the outside world without going through connect
    task "whoami" {
      driver = "docker"
      config {
        image = "containous/whoami"
      }
    }
  }
}
