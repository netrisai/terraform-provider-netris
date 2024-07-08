resource "netris_servercluster" "my-servercluster1" {
  name       = "my-servercluster1"
  adminid    = data.netris_tenant.admin.id
  siteid     = netris_site.santa-clara.id
  # vpcid      = netris_vpc.my-vpc.id
  templateid = netris_serverclustertemplate.my-serverclustertemplate1.id
  tags       = ["boo", "foo"]
  servers = [
    netris_server.my-server01.id,
    netris_server.my-server02.id,
  ]
}
