locals {
  apps_domain          = var.apps_lb_domain
  private_network_id   = var.private_network_id
  fullchain_cert_path  = var.cert_path
  cert_key_path        = var.cert_key_path
  masters_private_ipv4 = var.masters_private_ipv4
  minions_private_ipv4 = var.minions_private_ipv4
}

resource "hcloud_uploaded_certificate" "apps" {
  name = local.apps_domain
  private_key = <<EOF
${file(local.cert_key_path)}
EOF

  certificate = <<EOF
${file(local.fullchain_cert_path)}
EOF
}

resource "hcloud_load_balancer" "apps" {
  name               = local.apps_domain
  load_balancer_type = "lb11"
  location           = "ash"
}

resource "hcloud_load_balancer_network" "srvnetwork" {
  load_balancer_id = hcloud_load_balancer.apps.id
  network_id       = hcloud_network.local.private_network_id
}

resource "hcloud_load_balancer_service" "envoy" {
  load_balancer_id = hcloud_load_balancer.apps.id
  protocol         = "http"
  proxyprotocol    = false
  listen_port      = 80
  destination_port = 80

  http {
    certificates  = [hcloud_uploaded_certificate.apps.id]
    redirect_http = true
  }
}

resource "hcloud_load_balancer_target" "load_balancer_target" {
  for_each = toset(local.minions_private_ipv4)
  type             = "ip"
  load_balancer_id = hcloud_load_balancer.apps.id
  ip               = each.value
}
