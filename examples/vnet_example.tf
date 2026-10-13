resource "netris_vnet" "my-vnet" {
  name     = "my-vnet"
  tenantid = data.netris_tenant.admin.id
  state    = "active"
  # tags  = ["foo", "bar"]
  vlanid = "auto"
  sites {
    id = netris_site.santa-clara.id
    gateways {
      prefix = "198.18.51.1/24"
      # dhcp   = "enabled"
    }
    gateways {
      prefix = "2001:db8:acad::fffe/64"
    }
    interface {
      name = "swp8@my-switch01"
    }
    interface {
      name = "swp8@my-switch02"
    }
    interface {
      id = netris_lag.lag2-mc.id
    }
    interfacetag {
      tag        = "foo"
      accessmode = true
    }
    interfacetag {
      tag = "bar"
      # accessmode = true
    }
  }
  ipv6nd {
    routeradvertisement {
      mode                  = "enabled"
      routerlifetime        = "1800"
      advertisementinterval = "600"
      managedconfig         = true
      otherconfig           = true
    }
    prefixadvertisement {
      enabled           = true
      preferredlifetime = "604800"
      validlifetime     = "2592000"
      autoconfig        = true
    }
    rdnss {
      enabled    = true
      dnsservers = ["2001:db8:acad::1", "2001:db8:acad::2"]
      lifetime   = "1800"
      # infinite = true # when true, lifetime is ignored
    }
  }
  depends_on = [
    netris_switch.my-switch01,
    netris_switch.my-switch02,
    netris_subnet.my-subnet-vnet,
    netris_subnet.my-subnet-vnetv6,
  ]
}

resource "netris_vnet" "my-vnet-in-my-vpc" {
  name     = "my-vnet-in-my-vpc"
  tenantid = data.netris_tenant.admin.id
  state    = "active"
  # tags  = ["foo", "bar"]
  vpcid   = netris_vpc.my-vpc.id
  vxlanid = 456
  ipfamily = "ipv4"
  sites {
    id = netris_site.santa-clara.id
    gateways {
      prefix = "198.18.51.1/24"
      # dhcp   = "enabled"
    }
    gateways {
      prefix = "2001:db8:acad::fffe/64"
    }
    interface {
      name   = "swp9@my-switch01"
      vlanid = 1050
    }
    interface {
      name = "swp9@my-switch02"
    }
  }
  depends_on = [
    netris_switch.my-switch01,
    netris_switch.my-switch02,
    netris_subnet.my-subnet-vnet-in-my-vpc,
    netris_subnet.my-subnet-vnetv6-in-my-vpc,
  ]
}

resource "netris_vnet" "my-vnet2" {
  name     = "my-vnet2"
  tenantid = data.netris_tenant.admin.id
  state    = "active"
  # tags  = ["foo", "bar"]
  # vpcid   = netris_vpc.my-vpc.id
  # vxlanid = 0
  ipfamily = "ipv4"
  dhcprelay {
    enabled = true
    vpcid = 1
    primaryaddr = "192.168.10.1"
    secondaryaddr = "192.168.12.1"
  }
  # dhcpv6relay {
  #   enabled = true
  #   vpcid = 1
  #   primaryaddr = "2001:db8::1"
  #   secondaryaddr = "2001:db8::2"
  # }
  sites {
    id = netris_site.santa-clara.id
    gateways {
      prefix = "198.18.52.1/24"
      # dhcp   = "enabled"
    }
    interface {
      name   = "swp12@my-switch02"
      vlanid = 1058
    }
  }
  depends_on = [
    netris_switch.my-switch01,
    netris_switch.my-switch02,
    netris_subnet.my-subnet-vnet2,
  ]
}

# Dynamic VRF assignment uses a Dynamic VPC Group instead of vpcid.
resource "netris_vnet" "dynamic-access" {
  name              = "dynamic-access"
  tenantid          = 1
  l3vpn             = true
  dynamicvpcgroupid = netris_dynamic_vpc_group.access.id
  sites {
    id = 1
    interface {
      id = 1201
    }
  }
}
