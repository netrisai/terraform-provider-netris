# resource "netris_softgate" "my-softgate" {
#   name = "my-softgate"
#   tenantid = netris_tenant.admin.id
#   siteid = netris_site.santa_clara.id
#   description = "Softgate 1"
#   profile = ""
#   mainip = "198.51.100.11"
#   mgmtip = "192.0.2.11"
#   depends_on = [
#     netris_subnet.my-subnet-mgmt,
#     netris_subnet.my-subnet-loopback,
#   ]
# }