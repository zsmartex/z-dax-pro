output "consul_servers" {
  value = flatten([
    for index, node in hcloud_server.server_node : [
      for server in local.servers :
      {host = "${node.ipv4_address}", 
        host_name = "${node.name}", 
        private_ip = "${server.private_ip}",
        server_id= node.id
      } if server.name == node.name
    ] if node.labels["group"] == "consul"
  ])
}

output "nomad_servers" {
  value = flatten([
    for index, node in hcloud_server.server_node : [
      for server in local.servers :
      {host = "${node.ipv4_address}", 
        host_name = "${node.name}", 
        private_ip = "${server.private_ip}",
        server_id= node.id
      } if server.name == node.name
    ] if node.labels["group"] == "nomad-server"
  ])
}

output "database_servers" {
  value = flatten([
    for index, node in hcloud_server.server_node : [
      for server in local.servers :
      {host = "${node.ipv4_address}", 
        host_name = "${node.name}", 
        private_ip = "${server.private_ip}",
        server_id= node.id
      } if server.name == node.name
    ] if node.labels["group"] == "database-server"
  ])
}

output "vault_servers" {
  value = flatten([
    for index, node in hcloud_server.server_node : [
      for server in local.servers :
      {host = "${node.ipv4_address}", 
        host_name = "${node.name}", 
        private_ip = "${server.private_ip}",
        server_id= node.id
      } if server.name == node.name
    ] if node.labels["group"] == "vault"
  ])
}

output "client_servers" {
  value = flatten([
    for index, node in hcloud_server.server_node : [
      for server in local.servers :
      {host = "${node.ipv4_address}", 
        host_name = "${node.name}", 
        private_ip = "${server.private_ip}",
        server_id= node.id
      } if server.name == node.name
    ] if node.labels["group"] == "client"
  ])
}

output "o11y_servers" {
  value = flatten([
    for index, node in hcloud_server.server_node : [
      for server in local.servers :
      {host = "${node.ipv4_address}", 
        host_name = "${node.name}", 
        private_ip = "${server.private_ip}",
        server_id = node.id
      } if server.name == node.name
     ] if node.labels["group"] == "observability"
  ])
}

output "consul_volumes" {
  value = flatten([
    for index, attachment in hcloud_volume_attachment.consul : [
      {mount = "/mnt/HC_Volume_${attachment.volume_id}", 
      path = "/opt/consul",
      name = "",
      server_id = attachment.server_id,
      is_nomad = false
      }
    ] 
  ])
}

output "client_volumes" {
 value = flatten([
    for index, attachment in hcloud_volume_attachment.client_volumes : [
      for vol in var.client_volumes :
      {mount = "/mnt/HC_Volume_${attachment.volume_id}", 
      path = vol.path,
      name = vol.name,
      is_nomad = true,
      server_id = attachment.server_id} if hcloud_volume.client_volumes[index].name == vol.name
    ] 
  ])
}

output "database_volumes" {
 value = flatten([
    for index, attachment in hcloud_volume_attachment.database_volumes : [
      for vol in var.database_volumes :
      {mount = "/mnt/HC_Volume_${attachment.volume_id}", 
      path = vol.path,
      name = vol.name,
      is_nomad = true,
      server_id = attachment.server_id} if hcloud_volume.database_volumes[index].name == vol.name
    ] 
  ])
}
