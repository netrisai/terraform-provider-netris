resource "netris_route" "route-terraform-test" {
    description = "Terraform Test"
    prefix = "10.0.0.0/25"
    nexthop = "172.28.51.111"
    siteid = netris_site.santa-clara.id
    state = "active"
    # hwids = [netris_switch.my-switch01.id]
}
