resource "netris_user_role" "terrraform-userrole" {
  name = "terrraform"
  pgroup = "my-group"
  tenantids = [netris_tenant.my-tenant.id]
}
