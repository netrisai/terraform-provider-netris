resource "netris_permission_group" "my-group" {
  name = "my-group"
  description = "Terraform Test Permission group"
  groups   = [
    "services:view",
    "services.loadbalancer:edit",
    "services.acl:external-acl",
    "services.acltwozero:edit",
    "net:edit",
    "net.ipam:view",
    "accounts:view",
    "settings:view",
    "settings.checks:edit",
  ]
}
