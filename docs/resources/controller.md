---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "netris_controller Resource - terraform-provider-netris"
subcategory: ""
description: |-
  Creates and manages controllers
---

# netris_controller

Create and manage a controller resource in the inventory.
## Example Usages
```hcl
data "netris_tenant" "admin" {
  name = "Admin"
}

data "netris_site" "santa-clara" {
  name = "Santa Clara"
}

resource "netris_controller" "controller" {
  name = "controller"
  tenantid = data.netris_tenant.admin.id
  siteid = data.netris_site.santa-clara.id
  description = "Terraform Test Controller"
  mainip = "auto"
  depends_on = [
    netris_subnet.my-subnet-loopback,
  ]
}
```


<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **mainip** (String) A unique IP address which will be used as a loopback address of this unit. Valid value is ip address (example `198.51.100.10`) or `auto`. If set `auto` the controller will assign an ip address automatically from subnets with relevant purpose.
- **name** (String) User assigned name of controllers.
- **siteid** (Number) The site ID where this controller belongs.
- **tenantid** (Number) ID of tenant. Users of this tenant will be permitted to edit this unit.

### Optional

- **description** (String) Controller description.
