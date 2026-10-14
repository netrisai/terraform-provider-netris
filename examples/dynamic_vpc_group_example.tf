variable "radius_password" {
  type      = string
  sensitive = true
}

resource "netris_dynamic_vpc_group" "access" {
  name            = "access"
  resourcegroupid = netris_tenant.my-tenant.id
  radiusserver    = "netris-hosted"

  member {
    vpcid = netris_vpc.my-vpc.id
    user {
      username = "engineer"
      password = var.radius_password
    }
  }
}

data "netris_dynamic_vpc_group" "access" {
  groupid = netris_dynamic_vpc_group.access.id
}
