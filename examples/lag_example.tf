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
