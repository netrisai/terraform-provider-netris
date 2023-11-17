---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "netris_lag Data Source - terraform-provider-netris"
subcategory: ""
description: |-
  Data Source: LAG Network Interfaces
---

# Data Source: netris_lag

LAG Network Interfaces data.

## Example Usages

```hcl
data "netris_lag" "agg1_sw2"{
  name = "agg1@my-switch02"
  depends_on = [netris_switch.my-switch02]
}
```


<!-- schema generated by tfplugindocs -->
## Schema

### Argument Reference

- **name** (String) LAG Network Interfaces's exact name. Example `"<agg number>@<switch name>"`

### Attribute Reference

- **id** (String) The ID of this resource.
- **description** (String) LAG Network Interfaces desired description
- **tenantid** (Number) ID of tenant. Users of this tenant will be permitted to manage LAG Network Interfaces
- **mtu** (Number) MTU must be integer between 68 and 9216
- **extension** (Map of String) LAG Network Interfaces extension configurations