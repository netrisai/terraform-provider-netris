---
page_title: "netris_dynamic_vpc_group Resource - terraform-provider-netris"
subcategory: ""
description: |-
  Creates and manages a Dynamic VPC Group.
---

# netris_dynamic_vpc_group

Creates a Dynamic VPC Group with existing VPC members and hosted users or third-party RADIUS servers.

## Example Usages

```hcl
resource "netris_dynamic_vpc_group" "access" {
  name            = "access"
  resourcegroupid = netris_tenant.my-tenant.id
  radiusserver    = "netris-hosted"

  member {
    vpcid = netris_vpc.my-vpc.id
    user {
      username = "engineer"
      password = var.radius_password
    }
  }
}
```

## Schema

### Required

- **resourcegroupid** (Number) Resource Group owning the group and member VPCs. Requires replacement.
- **radiusserver** (String) `netris-hosted` or `third-party`. Requires replacement.
- **member** (Block Set, Min: 1) Full desired VPC members. Removing a member also removes its users. (see [below for nested schema](#nestedblock--member))

### Optional

- **name** (String) User-assigned name, up to 255 characters.
- **tags** (Set of String) Tags.
- **server** (Block List, Max: 3) Ordered RADIUS servers. Third-party groups require 1–3 servers and cannot contain users. Hosted groups cannot contain servers. (see [below for nested schema](#nestedblock--server))

### Read-Only

- **id** (String) Numeric group ID.
- **displayid** (String) Controller display ID.
- **radiusclients** (Block List) Clients derived from group V-Nets and Server Clusters. (see [below for nested schema](#nestedblock--radiusclients))

<a id="nestedblock--member"></a>
### Nested Schema for `member`

Required:

- **vpcid** (Number) Existing VPC ID. Must be unique within the group.

Optional:

- **user** (Block Set) Hosted users. Removing a user block deletes the user. (see [below for nested schema](#nestedblock--member--user))

Read-Only:

- **vpcname** (String) VPC name.
- **vni** (Number) VPC VNI, or `0` when unallocated.
- **vrfname** (String) VPC VRF name.

<a id="nestedblock--member--user"></a>
### Nested Schema for `member.user`

Required:

- **username** (String) Unique across the group. 1–63 ASCII letters, digits, dots, underscores, hyphens or `@`, beginning with a letter or digit. `DEFAULT` is reserved case-insensitively.

Optional:

- **password** (String, Sensitive) Required for a new username. 8–128 printable ASCII characters excluding spaces, double quotes and backslashes. Empty preserves an existing password.

Read-Only:

- **id** (Number) User ID.

<a id="nestedblock--server"></a>
### Nested Schema for `server`

Required:

- **host** (String) IPv4 or IPv6 address.

Optional:

- **authport** (Number) Authentication port. 1–65535. Defaults to `1812`.
- **acctport** (Number) Accounting port. 1–65535. Defaults to `1813`.
- **secret** (String, Sensitive) Up to 512 characters; required on create. Empty on update preserves an existing secret when the controller can match the endpoint or its previous position.

<a id="nestedblock--radiusclients"></a>
### Nested Schema for `radiusclients`

Read-Only:

- **nasidentifier** (String) RADIUS NAS identifier.
- **nasip** (String) RADIUS NAS IP address.
- **switchid** (Number) Switch ID.
- **switchname** (String) Switch name.

## Import

```shell
terraform import netris_dynamic_vpc_group.access 123
```

Imported credentials are empty; the API masks passwords and secrets. Configured values remain in Terraform state. A group referenced by a V-Net or Server Cluster cannot be deleted; deleting an unused group preserves its member VPCs.
