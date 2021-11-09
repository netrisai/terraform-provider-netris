# resource "netris_l4lb" "terraform-test" {
#     name = "terraform-test"
#     owner = "Admin"
#     site = "Santa Clara"
#     state = "active"
#     protocol = "tcp"
#     frontend = "10.0.4.2"
#     port = 456
#     check = {
#         type = "http"
#         timeout = 3000
#         requestPath =  "/"
#     }
# }