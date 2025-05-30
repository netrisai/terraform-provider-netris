---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "netris_network_interface Resource - terraform-provider-netris"
subcategory: ""
description: |-
  Manages Network Interfaces
---

# netris_network_interface

Network Interfaces can be directly managed by this resource.

~> **WARNING:** This resource won't create any new network interfaces. It's just for configuring already existing network interfaces.

## Example Usages

```hcl
data "netris_tenant" "admin" {
  name = "Admin"
}

resource "netris_network_interface" "swp10_my-switch" {
  name = "swp10"
  description = "swp10 - Some description"
  nodeid = netris_switch.my-switch.id
  tenantid = data.netris_tenant.admin.id
  # breakout = "4x10"
  mtu = 9005
  autoneg = "on"
  speed = "10g"
  # extension = {
  #   extensionname = "vlans10-14"
  #   vlanrange = "10-14"
  # }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **name** (String) Network Interface's exact name
- **nodeid** (Number) The node ID to whom this network interface belongs
- **tenantid** (Number) ID of tenant. Users of this tenant will be permitted to manage network interface

### Optional

- **autoneg** (String) Toggle auto negotiation. Possible values: `default`, `on`, `off`. Default value is `default`
- **breakout** (String) Toggle breakout. Possible values: `off`, `disabled`, `1x10`,`1x25`,`1x40`,`1x50`,`1x100`,`1x200`,`1x400`,`1x800`,`2x10`,`2x25`,`2x40`,`2x50`,`2x100`,`2x200`,`2x400`,`4x10`,`4x25`,`4x50`,`4x100`,`4x200`,`8x10`,`8x25`,`8x50`,`8x100`. Default value is `off`.
- **description** (String) Network Interface desired description
- **extension** (Map of String) Network Interface extension configurations.
- **mtu** (Number) MTU must be integer between 68 and 9216. Default value is `9000`
- **speed** (String) Toggle interface speed, make sure that current node supports the configured speed. Possibe values: `auto`, `1g`, `10g`, `25g`, `40g`, `50g`, `100g`, `200g`, `400g`. Default value is `auto`
