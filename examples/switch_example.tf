# data "netris_site" "default"{
#     name = "Default"
# }

# resource "netris_switch" "my-switch" {
#   name = "my-switch"
#   tenant = "Admin"
#   siteid = data.netris_site.default.id
#   description = "Terraform Test"
#   nos = "cumulus_linux"
#   asnumber = 4280000000
#   profile = "my-profile"
#   mainip = "auto"
#   mgmtip = "auto"
#   macaddress = ""
#   portcount = 16
# }