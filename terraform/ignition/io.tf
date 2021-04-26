variable "additional_ignition_files" {
  type        = list(string)
  description = "Additional ignition files to include in the ignition config"
  default     = []
}

variable "additional_systemd_units" {
  type        = list(string)
  description = "Additional systemd units to include in the ignition config"
  default     = []
}

variable "oauth2_introspect_url" {
  type        = string
  description = "Introspection url to validate access tokens"
}

variable "oauth2_client_id" {
  type        = string
  description = "Oauth client id. Used for token validation"
}

variable "wireguard_cidrs" {
  type        = list(string)
  description = "The IP addresses and associated wireguard peer subnets in CIDR notation. The length of this list determines how many ignition configs will be generated"
}

variable "wireguard_endpoint_base" {
  type        = string
  description = "The base hostname to which wireguard peers can connect. Endpoint hostname are generated by prefixing with an index number."
}

variable "wireguard_exposed_subnets" {
  type        = list(string)
  description = "The subnets that are exposed to wireguard peers in CIDR notation"
}

variable "wiresteward_version" {
  type        = string
  description = "The version of wiresteward to deploy (see https://github.com/utilitywarehouse/wiresteward/)"
  default     = "v0.2.0"
}

variable "traefik_image" {
  type        = string
  description = "Traefik image for the proxy service to wiresteward"
  default     = "traefik:v2.3.7"
}

variable "s3fs_access_key" {
  type        = string
  description = "The aws key for the user that has permissions on the s3 bucket to save traefik certs"
}

variable "s3fs_access_secret" {
  type        = string
  description = "The aws secret for the user that has permissions on the s3 bucket to save traefik certs"
}

variable "s3fs_bucket" {
  type        = string
  description = "The aws s3 bucket to save traefik certs. Assumes that it contains a numbered dir for every wiresteward server"
}

variable "s3fs_image" {
  type    = string
  default = "quay.io/utilitywarehouse/sys-s3fs:v1.89-1"
}

locals {
  instance_count      = length(var.wireguard_cidrs)
  wireguard_endpoints = [for i in range(local.instance_count) : "${i}.${var.wireguard_endpoint_base}"]
}

output "ignition" {
  value = data.ignition_config.wiresteward.*.rendered
}

output "endpoints" {
  value = local.wireguard_endpoints
}

output "ignition_systemd" {
  value = data.ignition_config.wiresteward.*.systemd
}

output "ignition_files" {
  value = data.ignition_config.wiresteward.*.files
}
