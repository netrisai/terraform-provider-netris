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
