variable "network" {
  default = "dev"
}

// Nomad & Consul
variable "cn_network" {
  default = var.network
}

variable "cn_nomad_cluster_name" {
  default = "nomad_cluster.local"
}

variable "cn_nomad_client_nodes" {
  default = 3
}

variable "cn_consul_version" {
  default = "1.12.2"
}

variable "cn_nomad_version" {
  default = "1.3.1"
}

variable "cn_nomad_client_config" {
  default = "${data("nomad_config")}/client.hcl"
}