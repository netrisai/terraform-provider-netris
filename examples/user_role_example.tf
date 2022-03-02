resource "netris_user_role" "terrraform-userrole" {
  name = "terrraform"
  description = "Terraform Test user role"
  pgroup = "my-group"
  tenantids = [netris_tenant.my-tenant.id]
  depends_on = [netris_permission_group.my-group]
}
