resource "netris_allocation" "my-allocation-mgmt" {
  name = "my-allocation-mgmt"
  prefix = "192.0.2.0/24"
  tenant = "Admin"
}

resource "netris_allocation" "my-allocation-loopback" {
  name = "my-allocation-loopback"
  prefix = "198.51.100.0/24"
  tenant = "Admin"
}

resource "netris_allocation" "my-allocation-common" {
  name = "my-allocation-common"
  prefix = "203.0.113.0/24"
  tenant = "Admin"
}
