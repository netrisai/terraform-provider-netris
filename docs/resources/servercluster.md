---
page_title: "netris_servercluster Resource - terraform-provider-netris"
subcategory: ""
description: |-
  Creates and manages ServerCluster
---

# netris_servercluster

Creates a Server Cluster from a Server Cluster Template. Only tags and server membership can be updated in place.

## Example Usages

```hcl
resource "netris_servercluster" "dynamic-access" {
  name              = "dynamic-access"
  adminid           = 1
  siteid            = 1
  templateid        = 10
  servers           = [101, 102]
  dynamicvpcgroupid = netris_dynamic_vpc_group.access.id
}
```

A Dynamic VPC Group applies only to untagged L3VPN template objects; UFM/NVLink objects are ignored. Deleting the cluster preserves its group and member VPCs.

Use `vpcmapping` for per-template-object selection:

```hcl
vpcmapping {
  postfix           = "access"
  dynamicvpcgroupid = netris_dynamic_vpc_group.access.id
}
```

Map each VPC-backed template postfix exactly once. Group mappings cannot also specify an existing VPC `id`, `group`, or `name`. Without `dynamicvpcgroupid`, `id = 0` creates a VPC.

## Schema

### Required

- **name** (String) Server Cluster name.
- **adminid** (Number) Admin tenant ID.
- **siteid** (Number) Site ID.
- **templateid** (Number) Server Cluster Template ID.
- **servers** (Set of Number) Server IDs.

### Optional

- **vpcid** (Number) Existing VPC ID. Conflicts with `dynamicvpcgroupid`. Omit both and `vpcmapping` to create a new VPC.
- **dynamicvpcgroupid** (Number) Existing Dynamic VPC Group ID. Conflicts with configured `vpcid` and `vpcmapping`. If omitted, adopts the group returned by the controller.
- **vpcmapping** (Block Set) Per-object selection, with required `postfix` and optional `id`, `group`, `name`, or `dynamicvpcgroupid`.
- **vlan** (Block Set) Required `postfix` and `vlan_id` values for template V-Nets whose VLAN mode is `specify`.
- **tags** (Set of String) Tags.

## Import

Import by Server Cluster name (or numeric API ID):

```sh
terraform import netris_servercluster.dynamic-access dynamic-access
```

Configure the actual `dynamicvpcgroupid` for a single-group import. For mixed clusters, use returned mappings with actual group IDs and resolved VPC IDs; original auto-create recipes cannot be recovered.
