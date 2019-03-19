variable "azure_config_version" {
  description = <<EOF
(internal) This declares the version of the AWS configuration variables.
It has no impact on generated assets but declares the version contract of the configuration.
EOF

  default = "0.1"
}

variable "azure_region" {
  type        = "string"
  description = "The target Azure region for the cluster."
}

variable "azure_bootstrap_vm_type" {
  type        = "string"
  description = "Instance type for the bootstrap node. Example: `Standard_DS4_v3`."
}

# variable "azure_master_vm_type" {
#   type        = "string"
#   description = "Instance type for the master node(s). Example: `Standard_DS4_v3`."
# }

# variable "azure_vm_image" {
#   type        = "string"
#   description = "VM Image for all nodes  Example: `image-foobar123`."
# }

variable "azure_extra_tags" {
  type = "map"

  description = <<EOF
(optional) Extra Azure tags to be applied to created resources.

Example: `{ "key" = "value", "foo" = "bar" }`
EOF

  default = {}
}

# variable "azure_master_root_volume_type" {
#   type        = "string"
#   description = "The type of volume for the root block device of master nodes."
# }

# variable "azure_master_root_volume_size" {
#   type        = "string"
#   description = "The size of the volume in gigabytes for the root block device of master nodes."
# }



# variable "azure_master_availability_zones" {
#   type        = "list"
#   description = "The availability zones in which to create the masters. The length of this list must match master_count."
# }