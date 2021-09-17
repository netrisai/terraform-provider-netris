resource "netris_subnet" "artash-terraform" {
  name = "artash-terraform"
  prefix = "33.0.2.0/24"
  tenant = "Admin"
  purpose = "management"
  defaultgateway = "33.0.2.2"
  sites = ["Yerevan"]
}
