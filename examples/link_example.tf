resource "netris_link" "sg_to_sw" {
  ports = [
    "swp1@my-softgate",
    "swp8@my-switch"
  ]
  depends_on = [netris_softgate.my-softgate, netris_switch.my-switch]
}
