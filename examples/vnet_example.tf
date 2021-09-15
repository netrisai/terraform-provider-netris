terraform {
  required_providers {
    netris = {
      source  = "netrisai/netris"
    }
  }
  required_version = ">= 0.13"
}

provider "netris" {
  # address = ""                         # overwrite env: NETRIS_ADDRESS
  # login = ""                           # overwrite env: NETRIS_LOGIN
  # password = ""                        # overwrite env: NETRIS_PASSWORD
}

resource "netris_vnet" "my-vnet" {
  name = "my-vnet"
  owner = "Admin"
  state = "active"
  gateways {
        prefix = "109.23.0.6/24"
  }
  ports {
        name = "swp9@Yerevan-Spine1"
        vlanid = 1050
  }
}
