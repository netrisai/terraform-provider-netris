resource "netris_controller" "controller" {
  name = "controller"
  tenant = "Admin"
  site = "Santa Clara"
  description = "Terraform Test Controller"
  mainip = "auto"
  depends_on = [
    netris_subnet.my-subnet-common,
  ]
}
