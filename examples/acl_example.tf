# resource "netris_acl" "test-acl" {
#     name = "test-acl"
#     action = "permit"
#     comment = "Terraform Test"
#     established = 1
#     icmptype = 1
#     proto = "tcp"
#     reverse = true
#     srcprefix = "192.0.2.0/24"
#     srcportgroup = "portgroup-terraform-test"
#     dstprefix = "0.0.0.0/0"
#     dstportfrom = 1
#     dstportto = 100
#     validuntil = "2026-01-02T23:15:05Z"
# }
