resource "netris_vnet" "my-vnet" {
  name = "my-vnet"
  tenantid = netris_tenant.admin.id
  state = "active"
  sites{
    id = netris_site.santa-clara.id
    gateways {
      prefix = "203.0.113.1/25"
    }
    gateways {
      prefix = "2001:db8:acad::fffe/64"
    }
    ports {
      name = "swp5@my-sw01"
      vlanid = 1050
    }
    ports {
      name = "swp7@my-sw02"
    }
  }
  depends_on = [
    netris_switch.my-sw01,
    netris_switch.my-sw02,
    netris_subnet.my-subnet-common,
  ]
}
