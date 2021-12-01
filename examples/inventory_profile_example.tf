# resource "netris_inventory_profile" "my-profile" {
#   name = "my-profile"
#   description = "My First Inventory Profile"
#   ipv4ssh = ["10.0.10.0/24", "172.16.16.16"]
#   ipv6ssh = ["2001:DB8::/32"]
#   timezone = "America/Los_Angeles"
#   ntpservers = ["0.pool.ntp.org", "132.163.96.5"]
#   dnsservers = ["1.1.1.1", "8.8.8.8"]
#   customrule {
#     sourcesubnet = "10.10.10.0/24"
#     srcport = ""
#     dstport = "22"
#     protocol = "udp"
#   }
# }
