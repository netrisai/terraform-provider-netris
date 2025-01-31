resource "netris_l4lb" "my_l4lb" {
  name     = "my-l4lb"
  tenantid = data.netris_tenant.admin.id
  siteid   = netris_site.santa-clara.id
  # state = "active"
  protocol = "tcp"
  frontend = "203.0.113.150"
  port     = 8443
  backend  = ["198.18.51.100:443", "198.18.51.101:443"]
  check = {
    type        = "http"
    timeout     = 3000
    requestPath = "/"
  }
  depends_on = [netris_subnet.my-subnet-load-balancer, netris_subnet.my-subnet-vnet]
}

resource "netris_l4lb" "my_l4lb-in-vpc" {
  name     = "my-l4lb-in-vpc"
  tenantid = data.netris_tenant.admin.id
  siteid   = netris_site.santa-clara.id
  # state = "active"
  protocol = "tcp"
  port     = 8443
  backend  = ["198.18.51.102:443", "198.18.51.103:443"]
  check = {
    type        = "http"
    timeout     = 3000
    requestPath = "/"
  }
  vpcid      = netris_vpc.my-vpc.id
  depends_on = [netris_subnet.my-subnet-load-balancer, netris_subnet.my-subnet-vnet-in-my-vpc]
}
