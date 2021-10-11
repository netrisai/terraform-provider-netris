resource "netris_softgate" "my-softgate" {
  name = "my-softgate"
  tenant = "Admin"
  site = "Santa Clara"
  description = "Softgate 1"
  # profile = ""
  mainip = "198.51.100.11"
  mgmtip = "192.0.2.11"
  # links{
  #   localport = "swp1@my-softgate"
  #   remoteport = "swp8@my-spine-1"
  # }
  depends_on = [
    netris_subnet.my-subnet-mgmt,
    netris_subnet.my-subnet-loopback,
  ]
}
