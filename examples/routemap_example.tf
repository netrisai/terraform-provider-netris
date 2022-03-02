resource "netris_routemap" "routemap-out" {
  name = "routemap-out"
  sequence {
    description = "routemap-out seq 5"
    policy      = "permit"
    match {
      type     = "ipv4_prefix_list"
      objectid = netris_bgp_object.my-bgp-object-multiline.id
    }
    action {
      type      = "set"
      parameter = "community"
      value     = "23456:1001"
    }
  }
}

resource "netris_routemap" "routemap-in" {
  name = "routemap-in"
  sequence {
    description = "routemap-out seq 5"
    policy      = "permit"
    match {
      type     = "ipv4_prefix_list"
      objectid = netris_bgp_object.my-bgp-object.id
    }
    match {
      type  = "med"
      value = "600"
    }
    action {
      type  = "goto"
      value = "10"
    }
  }
  sequence {
    description = "routemap-out seq 10"
    policy      = "permit"
    match {
      type     = "community"
      objectid = netris_bgp_object.my-bgp-object-community.id
    }
    action {
      type      = "set"
      parameter = "local_preference"
      value     = "90"
    }
    action {
      type      = "set"
      parameter = "as_path"
      value     = "2"
    }
  }
}
