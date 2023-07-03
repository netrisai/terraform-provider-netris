resource "netris_vpc" "my-vpc" {
  name = "my-vpc"
  tenantid = netris_tenant.my-tenant.id
  # guesttenantid {
  #   id = data.netris_tenant.admin.id
  # }
  # tags = ["foo", "bar"]
}
