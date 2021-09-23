resource "netris_switch" "artash-sww" {
      name = "artash-sww"
      tenant = "Artash"
      site = "Artash"
      description = "Terraform Test"
      nos = "cumulus_linux"
      asnumber = 4280000000
      profile = "YerevanUltimate"
      mainip = "99.0.1.5"
      mgmtip = "99.0.2.5"
      macaddress = ""
      portcount = 16
      links{
            localport = "swp6@artash-sww"
            remoteport = "swp1@testtushkanchik"
      }
}