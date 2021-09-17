resource "netris_vnet" "my-vnet" {
      name = "my-vnet"
      owner = "Admin"
      state = "active"
      sites{
            name = "Yerevan"
            gateways {
                  prefix = "109.23.0.6/24"
            }
            ports {
                  name = "swp9@yerevan-spine1"
                  vlanid = 1051
            }
      }
      sites {
            name = "Test"
            gateways {
                  prefix = "66.66.66.1/24"
            }
      }
}
