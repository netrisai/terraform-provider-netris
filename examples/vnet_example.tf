resource "netris_vnet" "my-vnet" {
  name     = "my-vnet"
  tenantid = data.netris_tenant.admin.id
  state    = "active"
  sites {
    id = netris_site.santa-clara.id
    gateways {
      prefix = "198.18.51.1/24"
    }
    gateways {
      prefix = "2001:db8:acad::fffe/64"
    }
    interface {
      name   = "swp8@my-switch01"
      vlanid = 1050
    }
    interface {
      name = "swp8@my-switch02"
      lacp = "on"
    }
  }
  depends_on = [
    netris_switch.my-switch01,
    netris_switch.my-switch02,
    netris_subnet.my-subnet-vnet,
    netris_subnet.my-subnet-vnetv6,
  ]
}
