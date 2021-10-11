resource "netris_subnet" "my-subnet-mgmt" {
  name = "my-subnet-mgmt"
  prefix = "192.0.2.0/24"
  tenant = "Admin"
  purpose = "management"
  defaultgateway = "192.0.2.1"
  sites = ["Santa Clara"]
  depends_on = [
    netris_allocation.my-allocation-mgmt,
  ]
}

resource "netris_subnet" "my-subnet-loopback" {
  name = "my-subnet-loopback"
  prefix = "198.51.100.0/24"
  tenant = "Admin"
  purpose = "loopback"
  defaultgateway = ""
  sites = ["Santa Clara"]
  depends_on = [
    netris_allocation.my-allocation-loopback,
  ]
}

resource "netris_subnet" "my-subnet-common" {
  name = "my-subnet-common"
  prefix = "203.0.113.0/24"
  tenant = "Admin"
  purpose = "common"
  defaultgateway = ""
  sites = ["Santa Clara"]
  depends_on = [
    netris_allocation.my-allocation-common,
  ]
}
