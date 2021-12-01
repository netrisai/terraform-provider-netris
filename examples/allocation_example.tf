data "netris_tenant" "admin"{
  name = "Admin"
}

resource "netris_allocation" "my-allocation-mgmt" {
  name = "my-allocation-mgmt"
  prefix = "192.0.2.0/24"
  tenantid = data.netris_tenant.admin.id
  depends_on = [
    netris_site.santa-clara,
  ]
}

resource "netris_allocation" "my-allocation-loopback" {
  name = "my-allocation-loopback"
  prefix = "198.51.100.0/24"
  tenantid = data.netris_tenant.admin.id
  depends_on = [
    netris_site.santa-clara,
  ]
}

resource "netris_allocation" "my-allocation-common" {
  name = "my-allocation-common"
  prefix = "203.0.113.0/24"
  tenantid = data.netris_tenant.admin.id
  depends_on = [
    netris_site.santa-clara,
  ]
}
