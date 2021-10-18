resource "netris_bgp" "terraform-bgp" {
  name = "terraform-bgp"
  site = "Artash"
  hardware = "terraform-sg"
  neighboras =23456234
  transport = {
    type = "vnet"
    name = "terraform-vnet"
  }
  localip = "99.0.3.4/24"
  remoteip = "99.0.3.5/24"
  description = "Terraform Test"
  state = "enabled"
  multihop = {
    neighboraddress = "8.8.8.9"
    updatesource = "1.0.32.22"
    hops = "5"
  }
  bgppassword = "somestrongpass"
  allowasin = 5
  defaultoriginate = false
  prefixinboundmax = 10000
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