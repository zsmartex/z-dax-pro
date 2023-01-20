provider "nomad" {
  address = "http://localhost:4646"
  version = "~> 1.4"
}

resource "nomad_job" "traefik" {
  jobspec = file("${path.module}/jobs/traefik.nomad.hcl")
}

resource "nomad_job" "backend" {
  jobspec = file("${path.module}/jobs/backend.nomad.hcl")
}
