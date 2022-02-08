resource "netris_acl" "my_acl" {
  name = "my-acl"
  action = "permit"
  proto = "all"
  reverse = true
  srcprefix = "192.0.2.0/24"
  dstprefix = "0.0.0.0/0"
  # comment = "Terraform Test"
  # established = 1
  # icmptype = 1
  # srcportgroup = "portgroup-terraform-test"
  # dstportfrom = 1
  # dstportto = 100
  # validuntil = "2026-01-02T23:15:05Z"
}
