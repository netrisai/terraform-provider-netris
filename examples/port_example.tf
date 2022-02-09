resource "netris_port" "swp10_sw01" {
  name = "swp10"
  description = "swp10 - Some description"
  switchid = netris_switch.sw01-nyc.id
  tenantid = data.netris_tenant.admin.id
  # breakout = "4x10"
  mtu = 9005
  autoneg = "on"
  speed = "10g"
  # extension = {
  #   extensionname = "vlans10-14"
  #   vlanrange = "10-14"
  # }
}
