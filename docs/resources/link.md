---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "netris_link Resource - terraform-provider-netris"
subcategory: ""
description: |-
  Creates and manages Links
---

# netris_link

Define the links in the network. Once the links have been defined, the network is automatically configured as long as physical connectivity is in place and Netris Agents can communicate with Netris Controller.

~> **Note:** Link require hardware to exist prior to resource creation. Use `depends_on` to set an explicit dependency on the hardware (switch/softgate).


## Example Usages
```hcl
resource "netris_link" "sg_to_sw" {
  ports = [
    "swp1@my-softgate",
    "swp8@my-switch"
  ]
  depends_on = [netris_softgate.my-softgate, netris_switch.my-switch]
}
```



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **ports** (List of String) List of two ports.
