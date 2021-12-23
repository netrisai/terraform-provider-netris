# data "netris_site" "santa-clara"{
#     name = "Santa Clara"
# }

# resource "netris_route" "route-terraform-test" {
#     description = "Terraform Test"
#     prefix = "10.0.0.0/24"
#     nexthop = "10.0.0.5"
#     siteid = data.netris_site.santa-clara.id
#     state = "active"
#     hwids = [688, 687]
# }