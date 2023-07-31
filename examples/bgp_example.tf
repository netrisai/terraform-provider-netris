data "netris_network_interface" "swp14_sw1" {
  name       = "swp14@my-switch01"
  depends_on = [netris_switch.my-switch01]
}

data "netris_network_interface" "swp14_sw2" {
  name       = "swp14@my-switch02"
  depends_on = [netris_switch.my-switch02]
}

data "netris_network_interface" "swp13_sw1" {
  name       = "swp13@my-switch01"
  depends_on = [netris_switch.my-switch01]
}

data "netris_network_interface" "swp13_sw2" {
  name       = "swp13@my-switch02"
  depends_on = [netris_switch.my-switch02]
}


resource "netris_bgp" "my-bgp-isp1" {
  name             = "my-bgp-isp1"
  siteid           = netris_site.santa-clara.id
  hardware         = "my-softgate01"
  neighboras       = 23456
  portid           = data.netris_network_interface.swp14_sw1.id
  vlanid           = 3000
  localip          = "172.19.25.2/30"
  remoteip         = "172.19.25.1/30"
  description      = "My ISP1 BGP"
  inboundroutemap  = netris_routemap.routemap-in.id
  outboundroutemap = netris_routemap.routemap-out.id
  #  state = "enabled"
  #  multihop = {
  #    neighboraddress = "185.54.21.5"
  #    updatesource = "198.51.100.11/32"
  #    hops = "5"
  #  }
  #  bgppassword = "somestrongpass"
  #  allowasin = 5
  #  defaultoriginate = false
  #  prefixinboundmax = 1000
  #  localpreference  = 100
  #  weight = 0
  #  prependinbound = 2
  #  prependoutbound = 1
  #  prefixlistinbound = ["deny 127.0.0.0/8 le 32", "permit 0.0.0.0/0 le 24"]
  #  prefixlistoutbound = ["permit 192.0.2.0/24", "permit 198.51.100.0/24 le 25"]
  #  sendbgpcommunity = ["65501:777"]
  depends_on = [netris_softgate.my-softgate01]
}

resource "netris_bgp" "my-bgp-isp2" {
  name        = "my-bgp-isp2"
  siteid      = netris_site.santa-clara.id
  hardware    = "my-softgate02"
  neighboras  = 64600
  portid      = data.netris_network_interface.swp14_sw2.id
  localip     = "172.19.35.2/30"
  remoteip    = "172.19.35.1/30"
  description = "My ISP2 BGP"
  #  inboundroutemap = netris_routemap.routemap-in.id
  #  outboundroutemap = netris_routemap.routemap-out.id
  #  state = "enabled"
  #  multihop = {
  #    neighboraddress = "185.54.21.5"
  #    updatesource = "198.51.100.11/32"
  #    hops = "5"
  #  }
  #  bgppassword = "somestrongpass"
  #  allowasin = 5
  #  defaultoriginate = false
  #  prefixinboundmax = 1000
  #  localpreference  = 100
  #  weight = 0
  #  prependinbound = 2
  prependoutbound    = 2
  prefixlistinbound  = ["deny 127.0.0.0/8 le 32", "permit 0.0.0.0/0 le 24"]
  prefixlistoutbound = ["permit 192.0.2.0/24", "permit 198.51.100.0/24 le 25", "permit 203.0.113.0/24 le 26"]
  #  sendbgpcommunity = ["65501:777"]
  depends_on = [netris_softgate.my-softgate02]
}


resource "netris_bgp" "my-bgp-isp1-in-my-vpc" {
  name             = "my-bgp-isp1-in-my-vpc"
  siteid           = netris_site.santa-clara.id
  hardware         = "my-softgate01"
  neighboras       = 23456
  portid           = data.netris_network_interface.swp13_sw1.id
  vlanid           = 3003
  localip          = "172.19.25.2/30"
  remoteip         = "172.19.25.1/30"
  description      = "My ISP1 BGP"
  vpcid            = netris_vpc.my-vpc.id
  inboundroutemap  = netris_routemap.routemap-in.id
  outboundroutemap = netris_routemap.routemap-out.id
  #  state = "enabled"
  #  multihop = {
  #    neighboraddress = "185.54.21.5"
  #    updatesource = "198.51.100.11/32"
  #    hops = "5"
  #  }
  #  bgppassword = "somestrongpass"
  #  allowasin = 5
  #  defaultoriginate = false
  #  prefixinboundmax = 1000
  #  localpreference  = 100
  #  weight = 0
  #  prependinbound = 2
  #  prependoutbound = 1
  #  prefixlistinbound = ["deny 127.0.0.0/8 le 32", "permit 0.0.0.0/0 le 24"]
  #  prefixlistoutbound = ["permit 192.0.2.0/24", "permit 198.51.100.0/24 le 25"]
  #  sendbgpcommunity = ["65501:777"]
  depends_on = [netris_softgate.my-softgate01]
}

resource "netris_bgp" "my-bgp-isp2-in-my-vpc" {
  name        = "my-bgp-isp2-in-my-vpc"
  siteid      = netris_site.santa-clara.id
  hardware    = "my-softgate02"
  neighboras  = 64600
  portid      = data.netris_network_interface.swp13_sw2.id
  localip     = "172.19.35.2/30"
  remoteip    = "172.19.35.1/30"
  description = "My ISP2 BGP"
  vpcid       = netris_vpc.my-vpc.id
  #  inboundroutemap = netris_routemap.routemap-in.id
  #  outboundroutemap = netris_routemap.routemap-out.id
  #  state = "enabled"
  #  multihop = {
  #    neighboraddress = "185.54.21.5"
  #    updatesource = "198.51.100.11/32"
  #    hops = "5"
  #  }
  #  bgppassword = "somestrongpass"
  #  allowasin = 5
  #  defaultoriginate = false
  #  prefixinboundmax = 1000
  #  localpreference  = 100
  #  weight = 0
  #  prependinbound = 2
  prependoutbound    = 2
  prefixlistinbound  = ["deny 127.0.0.0/8 le 32", "permit 0.0.0.0/0 le 24"]
  prefixlistoutbound = ["permit 192.0.2.0/24", "permit 198.51.100.0/24 le 25", "permit 203.0.113.0/24 le 26"]
  #  sendbgpcommunity = ["65501:777"]
  depends_on = [netris_softgate.my-softgate02]
}
