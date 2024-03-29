resource "netris_route" "route-terraform-test" {
  description = "Terraform Test"
  prefix      = "10.0.0.0/25"
  nexthop     = "198.18.51.111"
  siteid      = netris_site.santa-clara.id
  state       = "active"
  # hwids = [netris_switch.my-switch01.id]
}

resource "netris_route" "route-terraform-test-in-my-vpc" {
  description = "Terraform Test In MY VPC"
  prefix      = "10.0.0.0/25"
  nexthop     = "198.18.51.111"
  siteid      = netris_site.santa-clara.id
  state       = "active"
  # hwids = [netris_switch.my-switch01.id]
  vpcid = netris_vpc.my-vpc.id
}
