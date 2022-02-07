data "netris_site" "santa_clara"{
    name = "Santa Clara"
}

data "netris_port" "swp5_sw1"{
    name = "swp5@sw1"
}

resource "netris_bgp" "my-bgp" {
   name = "my-bgp"
   siteid = data.netris_site.santa_clara.id
   hardware = "softgate1"
   neighboras = 23456
   portid = data.netris_port.swp5_sw1.id
   vlanid = 3000
   localip = "192.0.0.2/30"
   remoteip = "192.0.0.1/30"
  #  description = "My First BGP"
  #  state = "enabled"
  #  multihop = {
  #    neighboraddress = "185.54.21.5"
  #    updatesource = "198.51.100.10"
  #    hops = "5"
  #  }
  #  bgppassword = "somestrongpass"
  #  allowasin = 5
  #  defaultoriginate = false
  #  prefixinboundmax = 1000
  #  inboundroutemap = "my-in-rm"
  #  outboundroutemap = "my-out-rm"
  #  localpreference  = 100
  #  weight = 0
  #  prependinbound = 2
  #  prependoutbound = 1
  #  prefixlistinbound = ["deny 127.0.0.0/8 le 32", "permit 0.0.0.0/0 le 24"]
  #  prefixlistoutbound = ["permit 192.0.2.0/24", "permit 198.51.100.0/24 le 25"]
  #  sendbgpcommunity = ["65501:777"]
  depends_on = [netris_softgate.softgate1, netris_switch.sw1]
}
