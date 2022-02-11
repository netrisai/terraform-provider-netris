data "netris_site" "santa_clara"{
    name = "Santa Clara"
}

data "netris_tenant" "admin"{
  name = "Admin"
}

resource "netris_l4lb" "my_l4lb" {
  name = "my-l4lb"
  tenantid = data.netris_tenant.admin.id
  siteid = data.netris_site.santa_clara.id
  # state = "active"
  protocol = "tcp"
  frontend = "203.0.113.7"
  port = 31434
  backend = ["192.0.2.100:443", "192.0.2.101:443"]
  check = {
    type = "http"
    timeout = 3000
    requestPath =  "/"
  }
}
