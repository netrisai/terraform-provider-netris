resource "netris_lag" "lag1-switch-01" {
  description = "my lag1"
  tenantid    = data.netris_tenant.admin.id
  mtu         = 9008
  # lacp        = "on"
  # extension = {
  #    extensionname = "ext1"
  #    vlanrange = "10-20"
  # }
  members = [
    "swp11@my-switch01",
    "swp12@my-switch01",
  ]
  depends_on = [
    netris_switch.my-switch01,
  ]
}

resource "netris_lag" "lag2-mc" {
  description = "my mc-lag"
  tenantid    = data.netris_tenant.admin.id
  mtu         = 9008
  # lacp        = "on"
  # extension = {
  #    extensionname = "ext1"
  #    vlanrange = "10-20"
  # }
  mclagid = 10
  members = [
    "swp10@my-switch01",
    "swp10@my-switch02",
  ]
  depends_on = [
    netris_switch.my-switch01,
    netris_switch.my-switch02
  ]
}
