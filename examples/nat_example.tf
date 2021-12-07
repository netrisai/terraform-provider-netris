# resource "netris_nat" "nat-terraform-test" {
#     name = "nat-terraform-test"
#     state = "enabled"
#     comment = "Terraform Test"
#     siteid = netris_site.santa-clara.id
#     action = "DNAT"
#     protocol = "tcp"
#     srcaddress = "0.0.0.0/0"
#     srcport = "25"
#     dstaddress = "10.10.10.0/24"
#     dstport = "25"
#     dnattoip = "1.2.3.4/32"
#     dnattoport = 100
#     snattoip = "2.0.2.1"
#     snattopool = "2.0.2.1/32"
# }