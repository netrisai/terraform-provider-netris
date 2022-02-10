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
  tenants = ["my-tenant"]
}
