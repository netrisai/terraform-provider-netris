# resource "netris_port" "swp4-artash-spine" {
#   name = "swp4"
#   description = "swp4 - Artash"
#   switchid = 860
#   tenantid = 128
#   breakout = "manual"
#   mtu = 9005
#   autoneg = "default"
#   speed = "1g"
#   extension = {
#     extensionname = "extname"
#     vlanrange = "10-14"
#   }
# }