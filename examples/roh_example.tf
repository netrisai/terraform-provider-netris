data "netris_site" "santa_clara"{
  name = "Santa Clara"
}

data "netris_tenant" "admin"{
  name = "Admin"
}

resource "netris_roh" "my_roh" {
  name = "my-roh"
  tenantid = data.netris_tenant.admin.id
  siteid = data.netris_site.santa_clara.id
  type = "physical"
  routingprofile = "default_agg"
  unicastips = ["192.168.2.50/24"]
  anycastips = []
  ports = ["swp3@my-switch1", "swp3@my-switch2"]
  depends_on = [netris_subnet.roh, netris_switch.my-switch1, netris_switch.my-switch1]
}

resource "netris_roh" "my_roh_anycast" {
  name = "my-roh-anycast"
  tenantid = data.netris_tenant.admin.id
  siteid = data.netris_site.santa_clara.id
  type = "physical"
  routingprofile = "default_agg"
  unicastips = ["192.168.2.61/24"]
  anycastips = ["192.168.2.60/24"]
  ports = ["swp2@my-switch1", "swp2@my-switch2"]
  depends_on = [netris_subnet.roh, netris_switch.my-switch1, netris_switch.my-switch1]
}
