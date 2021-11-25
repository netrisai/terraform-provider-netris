# resource "netris_inventory_profile" "inventory_profile-terraform" {
#   name = "inventory_profile-terraform"
#   description = "Terraform Profile"
#   ipv4ssh = ["2.2.2.2", "3.3.3.3"]
#   ipv6ssh = ["2001:0db8:85a3:0000:0000:8a2e:0370:7334"]
#   timezone = "Asia/Yerevan"
#   ntpservers = ["2.2.2.2", "3.3.3.3"]
#   dnsservers = ["2.2.2.2", "3.3.3.3"]
#   customrule {
#     sourcesubnet = "10.10.10.0/24"
#     srcport = "23"
#     dstport = "80"
#     protocol = "tcp"
#   }
# }