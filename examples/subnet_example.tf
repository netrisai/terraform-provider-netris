resource "netris_subnet" "my-subnet-mgmt" {
  name           = "my-subnet-mgmt"
  prefix         = "192.0.2.0/24"
  tenantid       = data.netris_tenant.admin.id
  purpose        = "management"
  defaultgateway = "192.0.2.1"
  siteids        = [netris_site.santa-clara.id]
  depends_on = [
    netris_allocation.my-allocation-mgmt,
  ]
}

resource "netris_subnet" "my-subnet-loopback" {
  name     = "my-subnet-loopback"
  prefix   = "198.51.100.0/24"
  tenantid = data.netris_tenant.admin.id
  purpose  = "loopback"
  siteids  = [netris_site.santa-clara.id]
  depends_on = [
    netris_allocation.my-allocation-loopback,
  ]
}

resource "netris_subnet" "my-subnet-common" {
  name     = "my-subnet-common"
  prefix   = "203.0.113.0/25"
  tenantid = data.netris_tenant.admin.id
  purpose  = "common"
  siteids  = [netris_site.santa-clara.id]
  depends_on = [
    netris_allocation.my-allocation-common,
  ]
}

resource "netris_subnet" "my-subnet-load-balancer" {
  name     = "my-subnet-load-balancer"
  prefix   = "203.0.113.128/26"
  tenantid = data.netris_tenant.admin.id
  purpose  = "load-balancer"
  siteids  = [netris_site.santa-clara.id]
  depends_on = [
    netris_allocation.my-allocation-common,
  ]
}

resource "netris_subnet" "my-subnet-nat" {
  name     = "my-subnet-nat"
  prefix   = "203.0.113.192/26"
  tenantid = data.netris_tenant.admin.id
  purpose  = "nat"
  siteids  = [netris_site.santa-clara.id]
  depends_on = [
    netris_allocation.my-allocation-common,
  ]
}

resource "netris_subnet" "my-subnet-roh" {
  name     = "my-subnet-roh"
  prefix   = "10.171.56.0/24"
  tenantid = data.netris_tenant.admin.id
  purpose  = "common"
  siteids  = [netris_site.santa-clara.id]
  depends_on = [
    netris_allocation.my-allocation-roh,
  ]
}

resource "netris_subnet" "my-subnet-vnet" {
  name     = "my-subnet-vnet"
  prefix   = "172.28.51.0/24"
  tenantid = data.netris_tenant.admin.id
  purpose  = "common"
  siteids  = [netris_site.santa-clara.id]
  depends_on = [
    netris_allocation.my-allocation-vnet,
  ]
}

resource "netris_subnet" "my-subnet-vnetv6" {
  name     = "my-subnet-vnetV6"
  prefix   = "2001:db8:acad::/64"
  tenantid = data.netris_tenant.admin.id
  purpose  = "common"
  siteids  = [netris_site.santa-clara.id]
  depends_on = [
    netris_allocation.my-allocation-vnetv6,
  ]
}
