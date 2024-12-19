resource "netris_switch" "my-switch01" {
  name        = "my-switch01"
  tenantid    = data.netris_tenant.admin.id
  siteid      = netris_site.santa-clara.id
  description = "Switch 01"
  nos         = "cumulus_nvue"
  asnumber    = "auto"
  profileid   = netris_inventory_profile.my-profile.id
  mainip      = "auto"
  mgmtip      = "auto"
  portcount   = 16
  tags  = ["foo", "bar"]
  depends_on = [
    netris_subnet.my-subnet-mgmt,
    netris_subnet.my-subnet-loopback,
  ]
}

resource "netris_switch" "my-switch02" {
  name        = "my-switch02"
  tenantid    = data.netris_tenant.admin.id
  siteid      = netris_site.santa-clara.id
  description = "Switch 02"
  nos         = "dell_sonic"
  asnumber    = "auto"
  profileid   = netris_inventory_profile.my-profile.id
  mainip      = "auto"
  mgmtip      = "auto"
  portcount   = 16
  depends_on = [
    netris_subnet.my-subnet-mgmt,
    netris_subnet.my-subnet-loopback,
  ]
}
