output "raw_ssh_user" {
  value = local.raw_ssh_user
}

output "sre_ipv4" {
  value = hcloud_server.sre.ipv4_address
}

output "sre_ipv6" {
  value = hcloud_server.sre.ipv6_address
}

output "masters_ipv4" {
  value = hcloud_server.masters.*.network.ip
}

output "minions_ipv4" {
  value = hcloud_server.minions.*.network.ip
}

output "private_network_id" {
  value = hcloud_network.workspace.id
}
