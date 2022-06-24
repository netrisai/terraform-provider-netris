resource "netris_acltwozero" "terraform-acltwozero" {
  name     = "terraform-acltwozero"
  privacy  = "public"
  tenantid = data.netris_tenant.admin.id
  state    = "enabled"
  publishers {
    instanceids = [netris_roh.my_roh_anycast.id]
    lbvips      = []
    prefixes    = ["203.0.113.128/26"]
    protocol {
      name        = "TCP-GROUP"
      protocol    = "tcp"
      portgroupid = netris_portgroup.my_portgroup.id
    }
    protocol {
      name     = "TCP"
      protocol = "tcp"
      port     = "80"
    }
  }
  subscribers {
    instanceids = [netris_roh.my_roh.id]
    prefix {
      prefix  = "198.18.51.0/24"
      comment = "vnet-prefix"
    }
  }
}
