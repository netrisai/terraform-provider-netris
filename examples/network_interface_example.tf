resource "netris_network_interface" "swp9_sw01" {
  name        = "swp9"
  description = "my port swp9"
  nodeid      = netris_switch.my-switch01.id
  tenantid    = data.netris_tenant.admin.id
  # breakout = "4x10"
  mtu     = 9050
  autoneg = "off"
  speed   = "40g"
  # extension = {
  #   extensionname = "vlans10-14"
  #   vlanrange = "10-14"
  # }
}

resource "netris_network_interface" "swp9_sw02" {
  name        = "swp9"
  description = "my port swp9"
  nodeid      = netris_switch.my-switch02.id
  tenantid    = data.netris_tenant.admin.id
  # breakout = "4x10"
  mtu     = 9050
  autoneg = "off"
  speed   = "40g"
  # extension = {
  #   extensionname = "vlans10-14"
  #   vlanrange = "10-14"
  # }
}
