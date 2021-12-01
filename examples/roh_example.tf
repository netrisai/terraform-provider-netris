# resource "netris_roh" "roh-terraform-test-1" {
#     name = "roh-terraform-test-1"
#     tenant = "Admin"
#     site = "Santa Clara"
#     type = "physical"
#     routingprofile = "default"
#     unicastips = ["7.0.0.34/9"]
#     anycastips = ["7.0.0.35/9"]
#     ports = ["swp5@leaf1"]
# }

# resource "netris_roh" "roh-terraform-test-2" {
#     name = "roh-terraform-test-2"
#     tenant = "Admin"
#     site = "Santa Clara"
#     type = "hypervisor"
#     unicastips = ["7.0.0.46/9"]
#     anycastips = ["7.0.0.7/9", "7.0.0.8/9"]
#     ports = ["swp3@leaf1"]
#     inboundprefixlist = ["permit 7.0.0.0/9 le 25", "permit 10.0.0.0/24 le 28"]
# }