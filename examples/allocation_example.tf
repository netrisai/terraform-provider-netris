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
  prefix   = "100.71.56.0/24"
  tenantid = data.netris_tenant.admin.id
}

resource "netris_allocation" "my-allocation-vnet" {
  name     = "my-allocation-vnet"
  prefix   = "198.18.51.0/24"
  tenantid = data.netris_tenant.admin.id
}

resource "netris_allocation" "my-allocation-vnetv6" {
  name     = "my-allocation-vnetV6"
  prefix   = "2001:db8:acad::/64"
  tenantid = data.netris_tenant.admin.id
}

resource "netris_allocation" "my-allocation-vnet-in-my-vpc" {
  name     = "my-allocation-vnet-in-my-vpc"
  prefix   = "198.18.51.0/24"
  tenantid = data.netris_tenant.admin.id
  vpcid    = netris_vpc.my-vpc.id
}

resource "netris_allocation" "my-allocation-vnetv6-in-my-vpc" {
  name     = "my-allocation-vnetV6-in-my-vpc"
  prefix   = "2001:db8:acad::/64"
  tenantid = data.netris_tenant.admin.id
  vpcid    = netris_vpc.my-vpc.id
}
