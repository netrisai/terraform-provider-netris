terraform {
  required_providers {
    netris = {
      source  = "netrisai/netris"
    }
  }
  required_version = ">= 0.13"
}

provider "netris" {
  address = "https://dev.netris.ai"
  login = "netris"
  password = "newNet0ps"
}

resource "netris_bgp" "artash-testt" {
  name = "artash-testt"
  site = "Yerevan"
  softgate = "sg01"
  neighboras =23456
  transport = {
    type = "vnet"
    name = "artash-test"
    vlanid = 4
  }
  localip = "1.0.32.1/24"
  remoteip = "1.0.32.2/24"
  description = "someDesc"
  state = "enabled"
  terminateonswitch = {
    enabled = "false"
    switchname = "spine1"
  }
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