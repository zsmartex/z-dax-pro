variable "hcloud_token" {
  sensitive = true
  type = string
}

variable "server_count" {
  type = number
  default = {{.Config.ClusterConfig.Servers}}
}

variable "consul_volume_size" {
  type = number
  default = {{.Config.ClusterConfig.ConsulVolumeSize}}
}

variable "database_count" {
  type = number
  default = {{.Config.ClusterConfig.DatabaseServers}}
}

variable "client_count" {
  type = number
  default = {{.Config.ClusterConfig.Clients}}
}

variable "vault_count" {
  type = number
  default = {{.Config.ClusterConfig.VaultServers}}
}

variable "separate_consul_servers"{
  type = bool
  default = {{.Config.ClusterConfig.SeparateConsulServers}}
}

variable "client_volumes" {
  type = list
  default = [{{ range $key, $value := .Config.ClusterConfig.ClientVolumes}}
   {
    name = "{{ $value.Name }}"
    client = "{{ $value.Client}}"
    path = "{{ $value.Path}}"
    size = {{ $value.Size }}
   },{{ end }}
  ]
}

variable "database_volumes" {
  type = list
  default = [{{ range $key, $value := .Config.ClusterConfig.DatabaseVolumes}}
   {
    name = "{{ $value.Name }}"
    server = "{{ $value.Server }}"
    path = "{{ $value.Path}}"
    size = {{ $value.Size }}
   },{{ end }}
  ]
}


variable "multi_instance_observability" {
  type = bool
  default = {{.Config.ObservabilityConfig.MultiInstance}}
}

variable "ssh_keys" {
  type = list
  default = [{{ range $key, $value := .Config.CloudProviderConfig.ProviderSettings.ssh_keys}}
   "{{ $value }}",{{ end }}
  ]
}

variable "base_server_name" {
  type = string
  default = "{{.Config.CloudProviderConfig.ProviderSettings.resource_names.base_server_name}}"
}

variable "load_balancer_type" {
  type = string
  default = "{{.Config.CloudProviderConfig.ProviderSettings.load_balancer_type}}"
}

variable "firewall_name" {
  type = string
  default = "{{.Config.CloudProviderConfig.ProviderSettings.resource_names.firewall_name}}"
}

variable "network_name" {
  type = string
  default = "{{.Config.CloudProviderConfig.ProviderSettings.resource_names.network_name}}"
}

variable "allow_ips" {
  type = list
  default = [{{ range $key, $value := .Config.CloudProviderConfig.ProviderSettings.allowed_ips}}
   "{{ $value }}",{{ end }}
  ]
}

variable "https_allowed_ips" {
  type = list
  default = [{{ range $key, $value := .HTTPSAllowedIPs}}
   "{{ $value }}",{{ end }}
  ]
}

variable "ssl_certificate_ids" {
  type = list
  default = [{{ range $key, $value := .Config.CloudProviderConfig.ProviderSettings.ssl_certificate_ids}}
   {{ $value }},{{ end }}
  ]
}

variable "server_instance_type"{
  type = string
  default = "{{.Config.CloudProviderConfig.ProviderSettings.server_instance_type}}"
}

variable "database_instance_type"{
  type = string
  default = "{{.Config.CloudProviderConfig.ProviderSettings.database_instance_type}}"
}

variable "client_instance_type"{
  type = string
  default = "{{.Config.CloudProviderConfig.ProviderSettings.client_instance_type}}"
}

variable "observability_instance_type"{
  type = string
  default = "{{.Config.CloudProviderConfig.ProviderSettings.observability_instance_type}}"
}
variable "location"{
  type = string
  default = "{{.Config.CloudProviderConfig.ProviderSettings.location}}"
}
