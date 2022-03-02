resource "netris_user" "terrraform-user" {
  username = "terraform"
  fullname = "Terraform"
  email = "terraform@netris.ai"
  emailcc = "devops@netris.ai"
  phone = "6504570097"
  company = "Netris, Inc."
  position = "DevOps Engineer"
  userrole = ""
  pgroup = "my-group"
  tenants {
    id = -1
    edit = false
  }
  tenants {
    id = netris_tenant.my-tenant.id
  }
  depends_on = [
    netris_tenant.my-tenant,
    netris_permission_group.my-group
  ]
}
