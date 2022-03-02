resource "netris_allocation" "my-allocation-mgmt" {
  name     = "my-allocation-mgmt"
  prefix   = "192.0.2.0/24"
  tenantid = data.netris_tenant.admin.id
}

resource "netris_allocation" "my-allocation-loopback" {
  name     = "my-allocation-loopback"
  prefix   = "198.51.100.0/24"
  tenantid = data.netris_tenant.admin.id
}

resource "netris_allocation" "my-allocation-common" {
  name     = "my-allocation-common"
  prefix   = "203.0.113.0/24"
  tenantid = data.netris_tenant.admin.id
}

resource "netris_allocation" "my-allocation-roh" {
  name     = "my-allocation-roh"
  prefix   = "10.171.56.0/24"
  tenantid = data.netris_tenant.admin.id
}

resource "netris_allocation" "my-allocation-vnet" {
  name     = "my-allocation-vnet"
  prefix   = "172.28.51.0/24"
  tenantid = data.netris_tenant.admin.id
}

resource "netris_allocation" "my-allocation-vnetv6" {
  name     = "my-allocation-vnetV6"
  prefix   = "2001:db8:acad::/64"
  tenantid = data.netris_tenant.admin.id
}
