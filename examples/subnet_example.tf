# resource "netris_subnet" "my-subnet-mgmt" {
#   name = "my-subnet-mgmt"
#   prefix = "192.0.2.0/24"
#   tenantid = netris_tenant.admin.id
#   purpose = "management"
#   defaultgateway = "192.0.2.1"
#   siteids = [netris_site.santa-clara.id]
#   depends_on = [
#     netris_allocation.my-allocation-mgmt,
#   ]
# }

# resource "netris_subnet" "my-subnet-loopback" {
#   name = "my-subnet-loopback"
#   prefix = "198.51.100.0/24"
#   tenantid = netris_tenant.admin.id
#   purpose = "loopback"
#   defaultgateway = ""
#   siteids = [netris_site.santa-clara.id]
#   depends_on = [
#     netris_allocation.my-allocation-loopback,
#   ]
# }

# resource "netris_subnet" "my-subnet-common" {
#   name = "my-subnet-common"
#   prefix = "203.0.113.0/24"
#   tenantid = netris_tenant.admin.id
#   purpose = "common"
#   defaultgateway = ""
#   siteids = [netris_site.santa-clara.id]
#   depends_on = [
#     netris_allocation.my-allocation-common,
#   ]
# }