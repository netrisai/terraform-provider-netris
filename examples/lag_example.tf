resource "netris_lag" "lag-terraform-switch-01" {
  description = "Terraform LAG"
  tenantid    = data.netris_tenant.admin.id
  mtu         = 9000                           # Optional
  lacp        = "off"                          # Optional
  extension = {                                # Optional
     extensionname = "aggsw1"
     vlanrange = "10-13"
  }
  members = [                                  # at least one member port is required
    "swp8@my-switch01",
    "swp9@my-switch01",
  ]
}