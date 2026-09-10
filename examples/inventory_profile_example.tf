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
    # optimisebgpoverlay    = true
    # optimisebgpoverlayhypervisor = true
    # unnumberedbgpunderlay = false
    # automaticlinkaggregation = true
    mclag = true
    # serverbasedesi = false
    fabrictype = "ew-plane1"
  }
  gpuclustersettings {
    # aggregatel3vpnprefix = true
    asicmonitoring       = false
    congestioncontrol    = false
    hwmp                 = false
    qosandroce           = true
    roceadaptiverouting  = true
    refarch              = "none"
  }

  ztpsettings {
    nos_image = "cumulus-linux-5.16.1-mlx-amd64.bin"
    password = "JiKqjtMxDFowRgfMmPMvj_Hw8ivHu6Y!Qym"
  }

  netqsettings {
    server_addrs = ["192.0.2.11"]
    server_port  = 32708
  }

  syslog_destinations {
    enabled     = true
    use_rfc5424 = true

    servers {
      host     = "syslog.example.com"
      port     = 514
      protocol = "TCP"
      severity = "Informational"
    }

    servers {
      host     = "192.0.2.10"
      port     = 514
      protocol = "UDP"
      severity = "Error"
    }
  }

}
