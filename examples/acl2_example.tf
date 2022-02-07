# resource "netris_acltwozero" "terraform-test" {
#   name = "terraform-test"
#   privacy = "public"
#   tenantid = 1
#   state = "enabled"
#   publishers {
#     instanceids = [netris_roh.roh-terraform-test-1.id]
#     lbvips = []
#     prefixes = ["192.0.2.0/24"]
#     protocol {
#       name = "TCP"
#       protocol = "tcp"
#       port = "80"
#       portgroupid = 22
#     }
#   }
#   subscribers {
#     instanceids = [netris_roh.roh-terraform-test-2.id]
#     prefix {
#       prefix = "198.51.100.0/25"
#       comment = "test-prefix"
#     }
#     prefix {
#       prefix = "203.0.113.0/24"
#       comment = "test-prefix-2"
#     }
#   }
# }
