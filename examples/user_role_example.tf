data "netris_tenant" "my-tenant" {
  name = "my-tenant"
}
resource "netris_user_role" "terrraform-userrole" {
  name = "terrraform"
  pgroup = "my-group"
  tenantids = [data.netris_tenant.my-tenant.id]
}
