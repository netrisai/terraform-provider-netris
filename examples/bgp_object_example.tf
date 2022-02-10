resource "netris_bgp_object" "my-bgp-object" {
  name = "my-bgp-object"
  type = "ipv4"
  value = "permit 10.10.0.0/22 ge 26 le 26"
}

resource "netris_bgp_object" "my-bgp-object-multiline" {
  name = "my-bgp-object-multiline"
  type = "ipv4"
  value = <<EOF
permit 10.10.10.0/24 le 27
permit 4.4.4.0/24
EOF
}

resource "netris_bgp_object" "my-bgp-object-community" {
  name = "my-bgp-object-community"
  type = "community"
  value = "permit 23456:501"
}
