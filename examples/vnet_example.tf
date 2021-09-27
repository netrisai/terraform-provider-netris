resource "netris_vnet" "artash-test" {
      name = "artash-test"
      owner = "Artash"
      state = ""
      sites{
            name = "Artash"
            gateways {
                  prefix = "99.0.3.1/24"
            }
            ports {
                  name = "swp10@artash-spine-1"
                  vlanid = 3000
            }
      }
}
