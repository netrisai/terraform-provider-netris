terraform {
  required_providers {
    netris = {
      source  = "netrisai/netris"
    }
  }
  required_version = ">= 0.13"
}

provider "netris" {
  address = ""
  login = ""
  password = ""
}

resource "netris_vnet" "myvnet" {
  name = "myvnet"
  owner = "Admin"
  state = "active"
  gateways {
        prefix = "1.0.0.10/19"
  }
  ports {
        name = "swp9@Yerevan-Spine1"
        vlanid = 4001
  }
}
