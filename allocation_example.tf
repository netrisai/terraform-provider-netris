resource "netris_ipam" "artash-ipam-test" {
  name = "artash-ipam-test"
  prefix = "10.0.0.0/24"
  tenant = "Admin"
}
