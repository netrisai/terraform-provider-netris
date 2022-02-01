# data "netris_site" "default-site"{
#     name = "Default"
# }

# data "netris_tenant" "default-tenant"{
#     name = "Admin"
# }

# resource "netris_l4lb" "terraform-test" {
#     name = "terraform-test"
#     tenantid = data.netris_tenant.default-tenant.id
#     siteid = data.netris_site.default-site.id
#     state = "active"
#     protocol = "tcp"
#     frontend = "10.0.4.2"
#     port = 456
#     check = {
#         type = "http"
#         timeout = 3000
#         requestPath =  "/"
#     }
#     backend = ["10.10.10.1:45"]
# }
