resource "netris_server" "my-server01" {
  name        = "my-server01"
  tenantid    = data.netris_tenant.admin.id
  siteid      = netris_site.santa-clara.id
  description = "Server 01"
  # mainip      = "auto"
  # mgmtip      = "auto"
  portcount = 2
  # asnumber    = "auto"
  # customdata  = "custom data"
  depends_on = [netris_subnet.my-subnet-mgmt, netris_subnet.my-subnet-loopback]
}

resource "netris_server" "my-server02" {
  name        = "my-server02"
  tenantid    = data.netris_tenant.admin.id
  siteid      = netris_site.santa-clara.id
  description = "Server 02"
  # mainip      = "auto"
  # mgmtip      = "auto"
  portcount = 2
  # asnumber    = "auto"
  # customdata  = "custom data"
  depends_on = [netris_subnet.my-subnet-mgmt, netris_subnet.my-subnet-loopback]
}
