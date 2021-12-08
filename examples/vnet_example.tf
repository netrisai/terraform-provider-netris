# resource "netris_vnet" "my-vnet" {
#   name = "my-vnet"
#   tenantid = netris_tenant.admin.id
#   state = "active"
#   sites{
#     id = netris_site.santa-clara.id
#     gateways {
#       prefix = "203.0.113.1/24"
#     }
#     ports {
#       name = "swp1@my-softgate"
#     }
#   }
#   depends_on = [
#     netris_softgate.my-softgate,
#   ]
# }