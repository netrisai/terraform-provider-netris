resource "netris_link" "sg1_to_sw1" {
  ports = [
    "swp1@my-softgate01",
    "swp16@my-switch01"
  ]
  depends_on = [netris_softgate.my-softgate01, netris_switch.my-switch01]
}

resource "netris_link" "sg2_to_sw2" {
  ports = [
    "swp1@my-softgate02",
    "swp16@my-switch02"
  ]
  depends_on = [netris_softgate.my-softgate02, netris_switch.my-switch02]
}

resource "netris_link" "sw1_to_sw2" {
  ports = [
    "swp15@my-switch01",
    "swp15@my-switch02"
  ]
  depends_on = [netris_switch.my-switch01, netris_switch.my-switch02]
}

resource "netris_link" "srv1_to_sw1" {
  ports = [
    "eth1@my-server01",
    "swp4@my-switch01"
  ]
  ipv4 = [
    "172.16.32.1/30",
    "172.16.32.2/30"
  ]
  ipv6 = [
    "fc00::c82a:75ff:fe66:84a0/127",
    "fc00::c82a:75ff:fe66:84a1/127"
  ]
  depends_on = [netris_switch.my-switch01, netris_server.my-server01]
}

resource "netris_link" "srv1_to_sw2" {
  ports = [
    "eth2@my-server01",
    "swp4@my-switch02"
  ]
  # ipv4 = [
  #   "172.16.32.5/30",
  #   "172.16.32.6/30"
  # ]
  # ipv6 = [
  #   "fc00::c82a:75ff:fe66:84b0/127",
  #   "fc00::c82a:75ff:fe66:84b1/127"
  # ]
  depends_on = [netris_switch.my-switch02, netris_server.my-server01]
}

resource "netris_link" "srv2_to_sw1" {
  ports = [
    "eth1@my-server02",
    "swp5@my-switch01"
  ]
  ipv4 = [
    "172.16.33.1/24",
    "172.16.33.2/24"
  ]
  ipv6 = [
    "fc00:c0::0/64",
    "fc00:c0::1/64"
  ]
  depends_on = [netris_switch.my-switch01, netris_server.my-server02]
}

resource "netris_link" "srv2_to_sw2" {
  ports = [
    "eth2@my-server02",
    "swp5@my-switch02"
  ]
  # ipv4 = [
  #   "172.16.33.5/24",
  #   "172.16.33.6/24"
  # ]
  # ipv6 = [
  #   "fc00:c0::2/64",
  #   "fc00:c0::3/64"
  # ]
  depends_on = [netris_switch.my-switch02, netris_server.my-server02]
}


resource "netris_link" "sw1_to_sw2_mc" {
  ports = [
    "swp7@my-switch01",
    "swp7@my-switch02"
  ]
  mclag {
    sharedipv4addr = "198.51.100.50"
    anycastmacaddr = "44:38:39:ff:00:f0"
  }
  depends_on = [netris_switch.my-switch01, netris_switch.my-switch02]
}
