resource "netris_user" "terrraform-user" {
  username = "terrraform-user"
  fullname = "Terraform"
  email = "test@test.test"
  emailcc = "test@test.test"
  phone = "123456789"
  company = "Company Inc."
  position = "Developer"
  userrole = "netris"
  pgroup = "netris"
  tenants = ["netris"]
}