terraform {
  required_providers {
    netris = {
      source  = "netrisai/netris"
      version = ">= 3.6.0"
    }
  }
  required_version = ">= 0.13"
}

provider "netris" {
  # address = "http://example.com"           # overwrite env: NETRIS_ADDRESS
  # login = "netris"                         # overwrite env: NETRIS_LOGIN
  # password = "newNet0ps"                   # overwrite env: NETRIS_PASSWORD
}
