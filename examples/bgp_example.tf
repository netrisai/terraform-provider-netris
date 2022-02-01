# data "netris_site" "santa-clara"{
#     name = "Santa Clara"
# }

# data "netris_port" "swp1-sg-test"{
#     name = "swp1@sg-test"
# }

# resource "netris_bgp" "my-bgp" {
#    name = "my-bgp"
#    siteid = data.netris_site.santa-clara.id
#    hardware = "sg-test"
#    neighboras = 23456
#    portid = data.netris_port.swp1-sg-test.id
#    vlanid = 3000
#    localip = "192.0.0.2/30"
#    remoteip = "192.0.0.1/30"
#    description = "My First BGP"
#    state = "enabled"
#    multihop = {
#      neighboraddress = "8.8.8.9"
#      updatesource = "1.0.32.22"
#      hops = "5"
#    }
#    bgppassword = "somestrongpass"
#    allowasin = 5
#    defaultoriginate = false
#    prefixinboundmax = 1000
#    inboundroutemap = "my-in-rm"
#    outboundroutemap = "my-out-rm"
#    localpreference  = 100
#    weight = 0
#    prependinbound = 2
#    prependoutbound = 1
#    prefixlistinbound = ["deny 127.0.0.0/8 le 32", "permit 0.0.0.0/0 le 24"]
#    prefixlistoutbound = ["permit 192.168.0.0/23"]
#    sendbgpcommunity = ["65501:777"]
# }
