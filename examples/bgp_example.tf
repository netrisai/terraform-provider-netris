# resource "netris_bgp" "my-bgp" {
#   name = "my-bgp"
#   site = "Santa Clara"
#   hardware = "my-softgate"
#   neighboras = 23456
#   transport = {
#     type = "vnet"
#     name = "my-vnet"
#   }
#   localip = "192.0.0.2/30"
#   remoteip = "192.0.0.1/30"
#   description = "My First BGP"
#   state = "enabled"
#   terminateonswitch = {
#     enabled = "false"
#     # switchname = "spine1"
#   }
#   # multihop = {
#   #   neighboraddress = "8.8.8.9"
#   #   updatesource = "1.0.32.22"
#   #   hops = "5"
#   # }
#   # bgppassword = "somestrongpass"
#   # allowasin = 5
#   # defaultoriginate = false
#   # prefixinboundmax = 10000
#   # inboundroutemap = "my-in-rm"
#   # outboundroutemap = "my-out-rm"
#   # localpreference  = 100
#   # weight = 0
#   # prependinbound = 2
#   # prependoutbound = 1
#   # prefixlistinbound = ["deny 127.0.0.0/8 le 32", "permit 0.0.0.0/0 le 24"]
#   # prefixlistoutbound = ["permit 192.168.0.0/23"]
#   # sendbgpcommunity = ["65501:777"]
# }
