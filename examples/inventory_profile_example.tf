resource "netris_inventory_profile" "my-profile" {
  name        = "my-profile"
  description = "My First Inventory Profile"
  ipv4ssh     = ["100.71.56.0/24", "203.0.113.0/24"]
  ipv6ssh     = ["2001:db8:acad::/64"]
  timezone    = "America/Los_Angeles"
  ntpservers  = ["0.pool.ntp.org", "132.163.96.5"]
  dnsservers  = ["1.1.1.1", "8.8.8.8"]
  customrule {
    description  = "my custom rule"
    sourcesubnet = "10.0.0.0/8"
    srcport      = ""
    dstport      = "8443"
    protocol     = "tcp"
  }
  fabricsettings {
    optimisebgpoverlay    = true
    unnumberedbgpunderlay = false
  }
  gpuclustersettings {
    aggregatel3vpnprefix = true
    asicmonitoring       = false
    congestioncontrol    = false
    qosandroce           = true
    roceadaptiverouting  = true
  }
}
