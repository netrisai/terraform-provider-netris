---
page_title: "netris_dynamic_vpc_group Data Source - terraform-provider-netris"
subcategory: ""
description: |-
  Reads a Dynamic VPC Group without credentials.
---

# Data Source: netris_dynamic_vpc_group

Reads a Dynamic VPC Group without credentials.

## Example Usages

```hcl
data "netris_dynamic_vpc_group" "access" {
  groupid = 123
}
```

## Schema

### Argument Reference

- **groupid** (Number) Positive numeric Dynamic VPC Group ID.

### Attribute Reference

- **id** (String) Numeric group ID.
- **name** (String) User-assigned name.
- **resourcegroupid** (Number) Owning Resource Group ID.
- **radiusserver** (String) `netris-hosted` or `third-party`.
- **tags** (Set of String) Tags.
- **displayid** (String) Controller display ID.
- **member** (Block List) VPC members and users. (see [below for nested schema](#nestedblock--member))
- **server** (Block List) Ordered RADIUS endpoints. (see [below for nested schema](#nestedblock--server))
- **radiusclients** (Block List) RADIUS clients. (see [below for nested schema](#nestedblock--radiusclients))

<a id="nestedblock--member"></a>
### Nested Schema for `member`

Attribute Reference:

- **vpcid** (Number) VPC ID.
- **vpcname** (String) VPC name.
- **vni** (Number) VPC VNI, or `0` when unallocated.
- **vrfname** (String) VPC VRF name.
- **user** (Block List) Hosted users. (see [below for nested schema](#nestedblock--member--user))

<a id="nestedblock--member--user"></a>
### Nested Schema for `member.user`

Attribute Reference:

- **id** (Number) User ID.
- **username** (String) Username.

<a id="nestedblock--server"></a>
### Nested Schema for `server`

Attribute Reference:

- **host** (String) IPv4 or IPv6 address.
- **authport** (Number) Authentication port.
- **acctport** (Number) Accounting port.

<a id="nestedblock--radiusclients"></a>
### Nested Schema for `radiusclients`

Attribute Reference:

- **nasidentifier** (String) RADIUS NAS identifier.
- **nasip** (String) RADIUS NAS IP address.
- **switchid** (Number) Switch ID.
- **switchname** (String) Switch name.
