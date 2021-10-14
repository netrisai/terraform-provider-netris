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
