# resource "netris_switch" "my-switch" {
#   name = "my-switch"
#   tenantid = netris_tenant.admin.id
#   siteid = netris_site.santa-clara.id
#   description = "Terraform Test"
#   nos = "cumulus_linux"
#   asnumber = 4280000000
#   profile = "my-profile"
#   mainip = "auto"
#   mgmtip = "auto"
#   macaddress = ""
#   portcount = 16
# }