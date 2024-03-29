---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "netris_dhcp_option_set Resource - terraform-provider-netris"
subcategory: ""
description: |-
  Creates and manages dhcp option sets
---

# netris_dhcp_option_set

Define a new dhcp option set resource in Netris.
## Example Usages
```hcl
resource "netris_dhcp_option_set" "dhcp_option_set" {
  name         = "my-dhcp-option-set"
  description  = "My DHCP Option Set"
  domainsearch = "dhcp.example.local"
  dnsservers   = ["1.1.1.1", "8.8.8.8"]
  ntpservers   = ["0.pool.ntp.org", "132.163.96.5"]
  leasetime    = 86400
  standardtoption {
    code  = 67
    value = "ipxe.efi"
  }
  customoption {
    code  = 239
    type  = "string"
    value = "http://192.0.2.55/script.sh"
  }
}
```


<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **name** (String) The name of DHCP Option Set

### Optional


- **description** (String) DHCP Option Set description
- **domainsearch** (String) The domain search that should be used as a suffix when resolving hostnames via the DNS
- **dnsservers** (List of String) List of IP addresses of DNS servers. Example `["1.1.1.1", "8.8.8.8"]`
- **ntpservers** (List of String) List of domain names or IP addresses of NTP servers. Example `["0.pool.ntp.org", "132.163.96.5"]`
- **leasetime** (Number) The amount of time (in seconds) a network device can use an IP address before being required to renew the lease. Default value is `86400`


- **standardtoption** (Block List) User-defined additional DHCP Options (see [below for nested schema](#nestedblock--standardtoption))
- **customoption** (Block List) User-defined additional custom DHCP Options (see [below for nested schema](#nestedblock--customoption))

<a id="nestedblock--standardtoption"></a>
### Nested Schema for `standardtoption`

Required:

- **code** (Number) DHCP Option code
- **value** (String) DHCP Option value

<a id="nestedblock--customoption"></a>
### Nested Schema for `customoption`

Required:

- **code** (Number) Custom DHCP Option code
- **type** (String) Custom DHCP Option type. Possible values: `string`, `boolean`, `uin8`, `uint16`, `uint32`, `int8`, `int16`, `int32`, `ipv4-address`, `fqdn`
- **value** (String) Custom DHCP Option value
