data "netris_inventory_profile" "my-profile" {
  name = "my-profile"
}

data "netris_site" "santa_clara" {
  name = "Santa Clara"
}

data "netris_tenant" "admin" {
  name = "Admin"
}

resource "netris_switch" "my-switch" {
  name = "my-switch"
  tenantid = netris_tenant.admin.id
  siteid = netris_site.santa-clara.id
  description = "Switch 01"
  nos = "cumulus_linux"
  asnumber = auto
  profileid = data.netris_inventory_profile.my-profile.id
  mainip = "auto"
  mgmtip = "auto"
  portcount = 16
}
