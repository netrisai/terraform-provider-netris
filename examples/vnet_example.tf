# resource "netris_vnet" "my-vnet" {
#   name = "my-vnet"
#   owner = "Admin"
#   state = "active"
#   sites{
#     name = "Santa Clara"
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