resource "netris_acltwozero" "terraform-acltwozero" {
  name = "terraform-acltwozero"
  privacy = "public"
  tenantid = data.netris_tenant.admin.id
  state = "enabled"
  publishers {
    instanceids = [netris_roh.roh_srv1.id]
    lbvips = []
    prefixes = ["192.0.2.0/24"]
    protocol {
      name = "TCP"
      protocol = "tcp"
      port = "80"
      # portgroupid = netris_portgroup.my_portgroup.id
    }
  }
  subscribers {
    instanceids = [netris_roh.roh_srv2.id]
    prefix {
      prefix = "198.51.100.0/25"
      comment = "test-prefix"
    }
    prefix {
      prefix = "203.0.113.0/24"
      comment = "test-prefix-2"
    }
  }
}
