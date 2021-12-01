# resource "netris_acl" "acl-terraform-test" {
#     name = "acl-terraform-test"
#     action = "permit"
#     comment = "Terraform Test"
#     established = 1
#     icmptype = 1
#     proto = "tcp"
#     reverse = true
#     srcprefix = "99.0.1.0/24"
#     srcportgroup = "anahittest"
#     dstprefix = "0.0.0.0/0"
#     dstportfrom = 1
#     dstportto = 100
#     validuntil = "2006-01-02T23:15:05Z"
# }