 resource "netris_bgp" "my-bgp" {
   name = "artash-terraform"
   site = "Artash"
   hardware = "artash-spine-1"
   neighboras = 23456
   transport = {
     type = "port"
     name = "swp3@artash-spine-1"
   }
   localip = "99.0.1.1/24"
   remoteip = "99.0.1.2/24"
   description = "BGP for Terraform test"
   state = "enabled"
    multihop = {
      neighboraddress = "8.8.8.9"
      updatesource = "1.0.32.22"
      hops = "5"
    }
    bgppassword = "somestrongpass"
    allowasin = 5
    defaultoriginate = false
    prefixinboundmax = 1000
    inboundroutemap = "my-in-rm"
    outboundroutemap = "my-out-rm"
    localpreference  = 100
    weight = 0
    prependinbound = 2
    prependoutbound = 1
    prefixlistinbound = ["deny 127.0.0.0/8 le 32", "permit 0.0.0.0/0 le 24"]
    prefixlistoutbound = ["permit 192.168.0.0/23"]
    sendbgpcommunity = ["65501:777"]
 }
