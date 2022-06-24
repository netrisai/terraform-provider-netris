resource "netris_acl" "my_acl" {
  name      = "my-acl"
  action    = "permit"
  proto     = "all"
  srcprefix = "203.0.113.128/26"
  dstprefix = "203.0.113.0/25"
  comment   = "Terraform Test"
  reverse   = true
  # established = 0
  # icmptype = 1
  # srcportgroup = "portgroup-terraform-test"
  # dstportfrom = 1
  # dstportto = 100
  # validuntil = "2026-01-02T23:15:05Z"
  depends_on = [netris_subnet.my-subnet-common]
}

resource "netris_acl" "my_acl2" {
  name      = "my-acl2"
  action    = "permit"
  proto     = "tcp"
  srcprefix = "198.18.51.0/24"
  dstprefix = "100.71.56.0/24"
  comment   = "Terraform Test 2"
  # reverse = true
  established = 1
  # icmptype = 1
  srcportgroup = "my_portgroup"
  dstportfrom  = 1
  dstportto    = 65535
  # validuntil = "2026-01-02T23:15:05Z"
  depends_on = [netris_subnet.my-subnet-roh, netris_portgroup.my_portgroup]
}
