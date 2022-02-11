resource "netris_controller" "controller" {
  name = "controller"
  tenantid = netris_tenant.admin.id
  siteid = netris_site.santa-clara.id
  description = "Terraform Test Controller"
  mainip = "auto"
  depends_on = [
    netris_subnet.my-subnet-loopback,
  ]
}
