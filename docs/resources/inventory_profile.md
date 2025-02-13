---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "netris_inventory_profile Resource - terraform-provider-netris"
subcategory: ""
description: |-
  Creates and manages inventory profiles
---

# netris_inventory_profile

Inventory profiles allow security hardening of inventory devices. By default all traffic flow destined to switch/SoftGate is allowed. As soon as the inventory profile is attached to a device it denies all traffic destined to the device except Netris-defined and user-defined custom flows.
## Example Usages
```hcl
resource "netris_inventory_profile" "my-profile" {
  name        = "my-profile"
  description = "My First Inventory Profile"
  ipv4ssh     = ["100.71.56.0/24", "203.0.113.0/24"]
  ipv6ssh     = ["2001:db8:acad::/64"]
  timezone    = "America/Los_Angeles"
  ntpservers  = ["0.pool.ntp.org", "132.163.96.5"]
  dnsservers  = ["1.1.1.1", "8.8.8.8"]
  customrule {
    description  = "my custom rule"
    sourcesubnet = "10.0.0.0/8"
    srcport      = ""
    dstport      = "8443"
    protocol     = "tcp"
  }
  fabricsettings {
    # optimisebgpoverlay    = true
    # unnumberedbgpunderlay = false
    # automaticlinkaggregation = false
    mclag = true
  }
  gpuclustersettings {
    aggregatel3vpnprefix = true
    asicmonitoring       = false
    congestioncontrol    = false
    qosandroce           = true
    roceadaptiverouting  = true
  }
}
```


<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **name** (String) The name of inventory profile
- **ipv4ssh** (List of String) List of IPv4 subnets allowed to ssh. Example `["10.0.10.0/24", "172.16.16.16"]`

### Optional

- **customrule** (Block List) Custom Rules configuration block. User defined rules to allow certain traffic. (see [below for nested schema](#nestedblock--customrule))
- **fabricsettings** (Block List) Fabric Settings. (see [below for nested schema](#nestedblock--fabricsettings))
- **gpuclustersettings** (Block List) GPU Cluster Specific Settings. Switch Fabric optimizations for GPU clusters. (see [below for nested schema](#nestedblock--gpuclustersettings))
- **description** (String) Inventory profile description
- **dnsservers** (List of String) List of IP addresses of DNS servers. Example `["1.1.1.1", "8.8.8.8"]`
- **ipv6ssh** (List of String) List of IPv6 subnets allowed to ssh. Example `["2001:DB8::/32"]`
- **ntpservers** (List of String) List of domain names or IP addresses of NTP servers. Example `["0.pool.ntp.org", "132.163.96.5"]`
- **timezone** (String) Devices using this inventory profile will adjust their system time to the selected timezone. Valid value is a name from the TZ [database](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones).

<a id="nestedblock--customrule"></a>
### Nested Schema for `customrule`

Required:

- **dstport** (String) Destination port. 1-65535, or empty for any.
- **protocol** (String) Protocol. Valid value is `udp`, `tcp` or `any`.
- **sourcesubnet** (String) Source Subnet. Example `10.0.0.0/8`
- **srcport** (String) Source port. 1-65535, or empty for any.

Optional: 

- **description** (String) Custom rule's description.


<a id="nestedblock--fabricsettings"></a>
### Nested Schema for `fabricsettings`

Optional:

- **optimisebgpoverlay** (Boolean) Optimize BGP Overlay for leaf-spine topology. When checked, overlay BGP updates will be optimized for large scale. Each leaf switch (based on name) will form its overlay BGP sessions only with two spine switches (with the lowest IDs). Otherwise, Overlay BGP sessions will be configured on p2p links alongside underlay. Default value is `false`.
- **optimisebgpoverlayhypervisor** (Boolean) Required for BGP/EVPN VXLAN integration with compute hypervisor networking. This optimization makes sure that a large number of hypervisor virtual networking EVPN prefixes do not overflow switch TCAM. Default value is `false`.
- **unnumberedbgpunderlay** (Boolean) When checked, BGP underlay sessions will be configured using p2p IPv4 addresses configured on link objects in the Netris controller. Otherwise, BGP unnumbered method is used and p2p ipv6 link-local addresses are used for BGP sessions. Default value is `false`.
- **automaticlinkaggregation** (Boolean) Automatically configure non-backbone switch ports under a single legged link aggregation (agg) interface. This allows for active/standby multihoming if LACP is enabled on the server side. Active/Active multihoming with EVPN-MH will be automatically configured on Nvidia Spectrum-2 and higher switch models. Default value is `false`.
- **mclag** (Boolean) Enabling MC-LAG functionality will disable any EVPN-MH functionality. Two multihoming methods are not supported simultaneously on the same switches. Default value is `false`.


<a id="nestedblock--gpuclustersettings"></a>
### Nested Schema for `gpuclustersettings`

Optional:

- **qosandroce** (Boolean) Optimize for RDMA over Converged Ethernet. Default value is `false`.
- **roceadaptiverouting** (Boolean) Enable Adaptive Routing for RoCE. Default value is `false`.
- **congestioncontrol** (Boolean) Enable Zero Touch RoCE Congestion Control. Default value is `false`.
- **asicmonitoring** (Boolean) Enable ASIC monitoring: Histograms and Telemetry Snapshots. Default value is `false`.
- **aggregatel3vpnprefix** (Boolean) Minimize prefix updates over BGP Overlay for L3VPN p2p links in rail-optimized topology and IP addressing schemes. Default value is `false`.
