resource "netris_controller" "controller" {
  name        = "my-controller"
  tenantid    = data.netris_tenant.admin.id
  siteid      = netris_site.santa-clara.id
  description = "Controller"
  mainip      = "auto"
  depends_on = [
    netris_subnet.my-subnet-loopback,
  ]
}
