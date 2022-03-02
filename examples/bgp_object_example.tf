resource "netris_bgp_object" "my-bgp-object" {
  name  = "my-bgp-object"
  type  = "ipv4"
  value = "permit 8.8.8.0/24"
}

resource "netris_bgp_object" "my-bgp-object-multiline" {
  name  = "my-bgp-object-multiline"
  type  = "ipv4"
  value = <<EOF
permit 192.0.2.0/24
permit 198.51.100.0/24
permit 203.0.113.0/24 le 26
EOF
}

resource "netris_bgp_object" "my-bgp-object-community" {
  name  = "my-bgp-object-community"
  type  = "community"
  value = "permit 23456:501"
}
