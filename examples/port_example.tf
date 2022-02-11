resource "netris_port" "swp8_sw01" {
  name = "swp8"
  description = "swp8 my-vnet"
  switchid = netris_switch.my-switch01.id
  tenantid = data.netris_tenant.admin.id
  # breakout = "4x10"
  mtu = 9050
  autoneg = "off"
  speed = "40g"
  # extension = {
  #   extensionname = "vlans10-14"
  #   vlanrange = "10-14"
  # }
}

resource "netris_port" "swp8_sw02" {
  name = "swp8"
  description = "swp8 my-vnet"
  switchid = netris_switch.my-switch02.id
  tenantid = data.netris_tenant.admin.id
  # breakout = "4x10"
  mtu = 9050
  autoneg = "off"
  speed = "40g"
  # extension = {
  #   extensionname = "vlans10-14"
  #   vlanrange = "10-14"
  # }
}
