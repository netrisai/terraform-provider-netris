resource "netris_acltwozero" "terraform-test" {
  name = "terraform-testt"
  privacy = "public"
  tenantid = 1
  state = "enabled"
  publishers {
    instanceids = [294]
    lbvips = []
    prefixes = ["10.10.10.0/24", "20.10.10.0/24", "30.10.10.0/24"]
    protocol {
      name = "TCPP"
      protocol = "tcp"
      port = "46"
      portgroupid = 22
    }
  }
  subscribers {
    instanceids = [294]
    prefix {
      prefix = "20.10.10.0/25"
      comment = "test-prefix"
    }
    prefix {
      prefix = "30.10.10.0/24"
      comment = "test-prefix-2"
    }
  }
}