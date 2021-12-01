# resource "netris_permission_group" "my-group" {
#   name = "my-group"
#   description = "Terraform Test Permission group"
#   groups   = [
# "services:view",
# 		"services.instances:view",
# 		"services.vnet:view",
# 		"services.acl:view",
# 		"services.acl:external-acl",
# 		"services.acltwozero:view",
# 		"services.aclportgroups:view",
# 		"services.loadbalancer:view",
# 		"services.l4loadbalancer:view",

# 		"net:view",
# 		"net.topology:view",
# 		"net.inventory:view",
# 		"net.inventoryprofiles:view",
# 		"net.switchports:view",
# 		"net.sites:view",
# 		"net.ebgp:view",
# 		"net.ebgpobjects:view",
# 		"net.ebgproutemaps:view",
# 		"net.ipam:view",
# 		"net.nat:view",
# 		"net.vpn:view",
# 		"net.routes:view",
# 		"net.lookinglass:view",

# 		"accounts:view",
# 		"accounts.users:view",
# 		"accounts.tenants:view",
# 		"accounts.userroles:view",
# 		"accounts.permissiongroups:view",

# 		"settings:view",
# 		"settings.general:view",
# 		"settings.whitelist:view",
# 		"settings.authentication:view",
# 		"settings.checks:view",

# 		"api.docs:view",
#   ]
# }
