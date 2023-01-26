locals {
  instance_name_prefix       = terraform.workspace
  instance_type              = var.instance_type
  jump_host_instance_type    = "cx11"
  masters_instance_type      = "cx11"
  minions_instance_type      = "cx11"
  instance_image             = "debian-11"
  instance_enable_ipv6       = true
  instance_enable_dynamic_ip = true
  ssh_public_key_name        = "${local.instance_name_prefix}_key"
  ssh_public_key_file        = var.ssh_public_key_file
  raw_ssh_user               = "root"
  private_subnet_cidr        = "192.168.42.0/24"
  private_subnet_gw          = "192.168.42.1"
}

resource "hcloud_ssh_key" "admin" {
  name       = local.ssh_public_key_name
  public_key = file(local.ssh_public_key_file)
}

resource "hcloud_firewall" "server" {
  name   = "${terraform.workspace}-sre"

  rule {
    direction = "in"
    protocol = "tcp"
    port = 22
    source_ips = [
      "0.0.0.0/0",
      "::/0"
    ]
  }
}

resource "hcloud_server" "sre" {
  name        = "${local.instance_name_prefix}-sre"
  server_type = local.jump_host_instance_type
  image       = local.instance_image
  location    = "ash"

  auto_delete = true
  public_net {
    ipv4_enabled = local.instance_enable_dynamic_ip
    ipv6_enabled = local.instance_enable_ipv6
  }

  firewall_ids = [hcloud_firewall.server.id]

  network {
    network_id = hcloud_network.workspace.id
  }

  user_data = templatefile("${path.module}/cloud-init.sre.yml", {})

  depends_on = [hcloud_network_subnet.workspace]
}

resource "hcloud_server" "masters" {
  count       = var.masters_count

  name        = "${local.instance_name_prefix}-master-0${count.index + 1}"
  server_type = local.masters_instance_type
  image       = local.instance_image
  location    = "ash"

  auto_delete = true
  public_net {
    ipv4_enabled = local.instance_enable_dynamic_ip
    ipv6_enabled = local.instance_enable_ipv6
  }

  firewall_ids = [hcloud_firewall.server.id]

  network {
    network_id = hcloud_network.workspace.id
  }

  user_data = templatefile("${path.module}/cloud-init.yml", { private_subnet_gw = local.private_subnet_gw, sre_ip = scaleway_instance_server.sre.private_ip })

  depends_on = [hcloud_network_subnet.workspace]
}

resource "hcloud_server" "minions" {
  count       = var.minions_count

  name        = "${local.instance_name_prefix}-minions-0${count.index + 1}"
  server_type = local.minions_instance_type
  image       = local.instance_image
  location    = "ash"

  auto_delete = true
  public_net {
    ipv4_enabled = local.instance_enable_dynamic_ip
    ipv6_enabled = local.instance_enable_ipv6
  }

  firewall_ids = [hcloud_firewall.server.id]

  network {
    network_id = hcloud_network.workspace.id
  }

  user_data = templatefile("${path.module}/cloud-init.yml", { private_subnet_gw = local.private_subnet_gw, sre_ip = scaleway_instance_server.sre.private_ip })

  depends_on = [hcloud_network_subnet.workspace]
}

resource "hcloud_network" "workspace" {
  name = terraform.workspace
  ip_range = local.private_subnet_cidr
}

resource "hcloud_network_route" "workspace" {
  network_id  = hcloud_network.workspace.id
  destination = "0.0.0.0/0"
  gateway     = local.private_subnet_gw
}

resource "hcloud_network_subnet" "workspace" {
  network_id = hcloud_network.workspace.id
  type = "cloud"
  network_zone = "us-west"
  ip_range = local.private_subnet_cidr
}