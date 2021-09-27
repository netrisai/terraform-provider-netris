resource "netris_switch" "artash-sww" {
      name = "artash-sww"
      tenant = "Artash"
      site = "Artash"
      description = "Terraform Test"
      nos = "cumulus_linux"
      asnumber = 4280000000
      profile = "YerevanUltimate"
      mainip = "auto"
      mgmtip = "auto"
      macaddress = ""
      portcount = 16
      links{
            localport = "swp13@artash-sww"
            remoteport = "swp14@artash-spine-1"
      }
      links{
            localport = "swp15@artash-sww"
            remoteport = "swp16@artash-spine-1"
      }
}
