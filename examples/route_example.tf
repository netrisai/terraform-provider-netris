data "netris_site" "santa-clara"{
    name = "Santa Clara"
}

resource "netris_route" "route-terraform-test" {
    description = "Terraform Test"
    prefix = "10.0.0.0/25"
    nexthop = "10.0.1.5"
    siteid = data.netris_site.santa-clara.id
    state = "active"
    # hwids = [netris_switch.my-switch1.id]
}
