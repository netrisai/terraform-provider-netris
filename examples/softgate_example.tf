data "netris_tenant" "admin" {
  name = "Admin"
}

resource "netris_softgate" "my-softgate01" {
  name        = "my-softgate01"
  tenantid    = data.netris_tenant.admin.id
  siteid      = netris_site.santa-clara.id
  description = "Softgate 1"
  profileid   = netris_inventory_profile.my-profile.id
  mainip      = "auto"
  mgmtip      = "auto"
  depends_on = [
    netris_subnet.my-subnet-mgmt,
    netris_subnet.my-subnet-loopback,
  ]
}

resource "netris_softgate" "my-softgate02" {
  name        = "my-softgate02"
  tenantid    = data.netris_tenant.admin.id
  siteid      = netris_site.santa-clara.id
  description = "Softgate 2"
  profileid   = netris_inventory_profile.my-profile.id
  mainip      = "auto"
  mgmtip      = "auto"
  # tags  = ["foo", "bar"]
  depends_on = [
    netris_subnet.my-subnet-mgmt,
    netris_subnet.my-subnet-loopback,
  ]
}

resource "netris_softgate" "my-softgate03" {
  name        = "my-softgate03"
  tenantid    = data.netris_tenant.admin.id
  siteid      = netris_site.santa-clara.id
  description = "Softgate 3 HS"
  profileid   = netris_inventory_profile.my-profile.id
  mainip      = "auto"
  mgmtip      = "auto"
  flavor      = "sg-hs"
  tags        = ["foo", "bar"]
  # role = "snat"
  depends_on = [
    netris_subnet.my-subnet-mgmt,
    netris_subnet.my-subnet-loopback,
  ]
}
