data "netris_inventory_profile" "my-profile" {
  name = "my-profile"
}

data "netris_site" "santa_clara" {
  name = "Santa Clara"
}

data "netris_tenant" "admin" {
  name = "Admin"
}

resource "netris_softgate" "my-softgate" {
  name = "my-softgate"
  tenantid = netris_tenant.admin.id
  siteid = netris_site.santa_clara.id
  description = "Softgate 1"
  profileid = data.netris_inventory_profile.my-profile.id
  mainip = "auto"
  mgmtip = "192.0.2.11"
  depends_on = [
    netris_subnet.my-subnet-mgmt,
    netris_subnet.my-subnet-loopback,
  ]
}
