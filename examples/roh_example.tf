resource "netris_roh" "my_roh" {
  name           = "my-roh"
  tenantid       = data.netris_tenant.admin.id
  siteid         = netris_site.santa-clara.id
  type           = "physical"
  routingprofile = "default_agg"
  unicastips     = ["10.171.56.50/24"]
  anycastips     = []
  ports          = ["swp3@my-switch01", "swp3@my-switch02"]
  depends_on     = [netris_subnet.my-subnet-roh, netris_switch.my-switch01, netris_switch.my-switch01]
}

resource "netris_roh" "my_roh_anycast" {
  name           = "my-roh-anycast"
  tenantid       = data.netris_tenant.admin.id
  siteid         = netris_site.santa-clara.id
  type           = "physical"
  routingprofile = "default_agg"
  unicastips     = ["10.171.56.61/24"]
  anycastips     = ["10.171.56.60/24"]
  ports          = ["swp2@my-switch01", "swp2@my-switch02"]
  depends_on     = [netris_subnet.my-subnet-roh, netris_switch.my-switch01, netris_switch.my-switch01]
}
