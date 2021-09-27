resource "netris_softgate" "artash-softgate" {
      name = "artash-softgate"
      tenant = "Artash"
      site = "Artash"
      description = "Artash Terraform Test"
      profile = "YerevanUltimate"
      mainip = "auto"
      mgmtip = "auto"
      links{
            localport = "swp1@artash-softgate"
            remoteport = "swp8@artash-spine-1"
      }
}
