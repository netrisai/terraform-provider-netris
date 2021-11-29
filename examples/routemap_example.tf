data "netris_bgp_object" "ipv4"{
    name = "ipv4_prefix_list_test"
}
data "netris_bgp_object" "ipv6"{
    name = "myIPv6"
}
data "netris_bgp_object" "lgc"{
    name = "large_community"
}
resource "netris_routemap" "routemap-terraform-test" {
    name = "routemap-terraform-test"
    sequence{
        description = "Terraform Test Seq"
        policy = "permit"
        match{
            type = "ipv4_prefix_list"
            objectid = data.netris_bgp_object.ipv4.itemid
        }
        match{
            type = "ipv4_next_hop"
            objectid = data.netris_bgp_object.ipv4.itemid
        }
    }
    sequence{
        description = "Terraform Test Seq 2"
        policy = "permit"
        match{
            type = "ipv6_prefix_list"
            objectid = data.netris_bgp_object.ipv6.itemid
        }
        match{
            type = "large_community"
            objectid = data.netris_bgp_object.lgc.itemid
        }
        match{
            type = "med"
            value = "6"
        }
        action{
            type = "set"
            parameter = "community"
            value = "0:10"
        }
        action{
            type = "goto"
            parameter = "community"
            value = "0:11"
        }
    }
}