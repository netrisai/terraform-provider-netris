data "netris_bgp_object" "ipv4"{
    name = "ipv4_prefix_list"
}
data "netris_bgp_object" "ipv6"{
    name = "ipv6_prefix_list"
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
            objectid = data.netris_bgp_object.ipv4.id
        }
        match{
            type = "ipv4_next_hop"
            objectid = data.netris_bgp_object.ipv4.id
        }
        action{
            type = "goto"
            parameter = "as_path"
            value = "10"
        }
    }
    sequence{
        description = "Terraform Test Seq 2"
        policy = "permit"
        match{
            type = "ipv6_prefix_list"
            objectid = data.netris_bgp_object.ipv6.id
        }
        match{
            type = "large_community"
            objectid = data.netris_bgp_object.lgc.id
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
    }
}
