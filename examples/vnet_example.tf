resource "netris_vnet" "my-vnet" {
  name = "my-vnet"
  owner = "Admin"
  state = "active"
  gateways {
        prefix = "109.23.0.6/24"
  }
  ports {
        name = "swp9@Yerevan-Spine1"
        vlanid = 1050
  }
}
