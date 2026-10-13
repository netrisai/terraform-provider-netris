/*
Copyright 2021. Netris, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package vnet

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/netrisai/netriswebapi/http"
	"github.com/netrisai/netriswebapi/v2/types/ipam"
	"github.com/netrisai/netriswebapi/v2/types/vnet"
	"github.com/netrisai/terraform-provider-netris/netris/subnet"

	api "github.com/netrisai/netriswebapi/v2"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: "Creates and manages Vnets",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the vnet",
			},
			"tenantid": {
				Required:    true,
				Type:        schema.TypeInt,
				ForceNew:    true,
				Description: "ID of tenant. Users of this tenant will be permitted to edit this unit.",
			},
			"state": {
				Optional:     true,
				Default:      "active",
				ValidateFunc: validateState,
				Type:         schema.TypeString,
				Description:  "V-Net state. Allowed values: `active` or `disabled`. Default value is `active`",
			},
			"ipfamily": {
				Optional:     true,
				Default:      "dual",
				ValidateFunc: validateIPFamily,
				Type:         schema.TypeString,
				Description:  "IP address family for the V-Net. Allowed values: `dual`, `ipv4`, or `ipv6`. Default value is `dual`.",
			},
			"vlanid": {
				Optional:     true,
				ValidateFunc: validateVlanID,
				Type:         schema.TypeString,
				Description:  "VLAN ID",
			},
			"sites": {
				Required:    true,
				Type:        schema.TypeList,
				Description: "Block of per site vnet configuration",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "The site ID. Ports from these sites will be allowed to participate in the V-Net. (Multi-site vnet would require backbone connectivity between sites).",
						},
						"interface": {
							Optional:    true,
							Type:        schema.TypeSet,
							Description: "Block of interfaces",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "Switch port ID",
									},
									"name": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Switch port name. Example: `swp5@my-sw01`",
									},
									"vlanid": {
										Default:     1,
										Type:        schema.TypeString,
										Optional:    true,
										Description: "VLAN tag for current port. If vlanid is not set - means port untagged",
									},
									"untagged": {
										Type:         schema.TypeString,
										Optional:     true,
										Description:  "Only when global 'vlanid' is specified",
										ValidateFunc: validateUntagged,
									},
									"lacp": {
										ValidateFunc: validateLACP,
										Default:      "off",
										Type:         schema.TypeString,
										Deprecated:   "You no longer need to specify this option. Create a V-Net with global VlanID, and Netris will automatically establish a Link Aggregation Group (LAG) using EVPN Multihoming.",
										Optional:     true,
										Description:  "LAG mode. Allows for active-standby dual-homing, assuming LAG configuration on the remote end. Valid value is `on` or `off`. Default value is `off`.",
									},
								},
							},
						},
						"interfacetag": {
							Optional:    true,
							Type:        schema.TypeSet,
							Description: "Network Interface Tags help referencing one or more network interface objects to one or more V-Net objects. Network interfaces with matching tags will appear in a given V-Net.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"tag": {
										Type:        schema.TypeString,
										Description: "Any tag. Example: `zone1`",
										Required:    true,
									},
									"accessmode": {
										Type:        schema.TypeBool,
										Default:     false,
										Optional:    true,
										Description: "Default value is `false`.",
									},
								},
							},
						},
						"ports": {
							Optional:    true,
							Type:        schema.TypeSet,
							Description: "Block of ports",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "Switch port ID",
									},
									"name": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Switch port name. Example: `swp5@my-sw01`",
									},
									"vlanid": {
										Default:     1,
										Type:        schema.TypeString,
										Optional:    true,
										Description: "VLAN tag for current port. If vlanid is not set - means port untagged",
									},
									"untagged": {
										Type:         schema.TypeString,
										Optional:     true,
										Description:  "Only when global 'vlanid' is specified",
										ValidateFunc: validateUntagged,
									},
									"lacp": {
										ValidateFunc: validateLACP,
										Default:      "off",
										Type:         schema.TypeString,
										Deprecated:   "You no longer need to specify this option. Create a V-Net with global VlanID, and Netris will automatically establish a Link Aggregation Group (LAG) using EVPN Multihoming.",
										Optional:     true,
										Description:  "LAG mode. Allows for active-standby dual-homing, assuming LAG configuration on the remote end. Valid value is `on` or `off`. Default value is `off`.",
									},
								},
							},
						},
						"gateways": {
							Optional:    true,
							Type:        schema.TypeSet,
							Description: "Block of gateways",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"prefix": {
										ValidateFunc: validateGateway,
										Type:         schema.TypeString,
										Required:     true,
										Description:  "The address will be serving as anycast default gateway for selected subnet. Example: `203.0.113.1/25`",
									},
									"vlanid": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"dhcp": {
										ValidateFunc: validateDHCP,
										Type:         schema.TypeString,
										Default:      "disabled",
										Optional:     true,
									},
									"dhcpoptionsetid": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
									},
									"dhcpstartip": {
										ValidateFunc: validateGateway,
										Type:         schema.TypeString,
										Optional:     true,
										Computed:     true,
									},
									"dhcpendip": {
										ValidateFunc: validateGateway,
										Type:         schema.TypeString,
										Optional:     true,
										Computed:     true,
									},
								},
							},
						},
					},
				},
			},
			"tags": {
				Computed: true,
				Optional: true,
				Type:     schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"l3vpn": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Enable L3VPN. Required to be true when dynamicvpcgroupid is configured.",
			},
			"dynamicvpcgroupid": {
				Type:          schema.TypeInt,
				Optional:      true,
				ForceNew:      true,
				ValidateFunc:  validation.IntAtLeast(1),
				ConflictsWith: []string{"vpcid"},
				Description:   "Dynamic VPC Group ID. Requires an untagged L3VPN V-Net without gateways or DHCP relay. Conflicts with vpcid.",
			},
			"dynamicvrf": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the controller enables dynamic VRF assignment.",
			},
			"vpcid": {
				ForceNew:    true,
				Optional:    true,
				Type:        schema.TypeInt,
				Description: "ID of VPC. If not specified, the V-Net will be created in the VPC marked as a default.",
			},
			"vxlanid": {
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeInt,
				Description: "VXLAN ID. If not specified will be generated automatically.",
			},
			"dhcprelay": {
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Type:        schema.TypeList,
				Description: "DHCP Relay configuration. Enabling DHCP Relay disables DHCP configuration under Gateways.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Enable DHCP Relay for this V-Net. Enabling it disables DHCP configuration under Gateways. Default value is `false`.",
						},
						"vpcid": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "ID of the VPC where the DHCP Relay servers reside.",
						},
						"primaryaddr": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Primary DHCP Relay address.",
						},
						"secondaryaddr": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Secondary DHCP Relay address.",
						},
					},
				},
			},
			"dhcpv6relay": {
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Type:        schema.TypeList,
				Description: "DHCPv6 Relay configuration. Enabling DHCPv6 Relay requires an IPv6 Gateway on the V-Net.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Enable DHCPv6 Relay for this V-Net. Requires an IPv6 Gateway on the V-Net. Default value is `false`.",
						},
						"vpcid": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "ID of the VPC where the DHCPv6 Relay servers reside.",
						},
						"primaryaddr": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Primary DHCPv6 Relay address.",
						},
						"secondaryaddr": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Secondary DHCPv6 Relay address.",
						},
					},
				},
			},
			"ipv6nd": ipv6NDSchema(true),
		},
		Create: resourceCreate,
		Read:   resourceRead,
		Update: resourceUpdate,
		Delete: resourceDelete,
		Exists: resourceExists,
		Importer: &schema.ResourceImporter{
			State: resourceImport,
		},
		CustomizeDiff: customizeDiff,
	}
}

func ipv6NDSchema(computed bool) *schema.Schema {
	return &schema.Schema{
		Optional:    true,
		Computed:    computed,
		MaxItems:    1,
		Type:        schema.TypeList,
		Description: "IPv6 Neighbor Discovery configuration.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"routeradvertisement": {
					Optional: true,
					Computed: computed,
					MaxItems: 1,
					Type:     schema.TypeList,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"mode": {
								Type:         schema.TypeString,
								Optional:     true,
								Computed:     computed,
								ValidateFunc: validateIPv6NDRAMode,
								Description:  "Router Advertisement mode. Allowed values: `default`, `enabled`, or `disabled`.",
							},
							"routerlifetime": {
								Type:         schema.TypeString,
								Optional:     true,
								Computed:     computed,
								ValidateFunc: validateNonNegativeIntString,
								Description:  "Router lifetime in seconds. Use an empty value to use the platform default.",
							},
							"advertisementinterval": {
								Type:         schema.TypeString,
								Optional:     true,
								Computed:     computed,
								ValidateFunc: validateNonNegativeIntString,
								Description:  "Interval between unsolicited Router Advertisements in seconds. Use an empty value to use the platform default.",
							},
							"managedconfig": {
								Type:        schema.TypeBool,
								Optional:    true,
								Computed:    computed,
								Description: "Managed address configuration flag.",
							},
							"otherconfig": {
								Type:        schema.TypeBool,
								Optional:    true,
								Computed:    computed,
								Description: "Other configuration flag.",
							},
						},
					},
				},
				"prefixadvertisement": {
					Optional: true,
					Computed: computed,
					MaxItems: 1,
					Type:     schema.TypeList,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"enabled": {
								Type:        schema.TypeBool,
								Optional:    true,
								Computed:    computed,
								Description: "Enable Prefix Information in Router Advertisements.",
							},
							"preferredlifetime": {
								Type:         schema.TypeString,
								Optional:     true,
								Computed:     computed,
								ValidateFunc: validateNonNegativeIntString,
								Description:  "Preferred lifetime of the advertised prefix in seconds. Use an empty value to use the platform default.",
							},
							"validlifetime": {
								Type:         schema.TypeString,
								Optional:     true,
								Computed:     computed,
								ValidateFunc: validateNonNegativeIntString,
								Description:  "Valid lifetime of the advertised prefix in seconds. Use an empty value to use the platform default.",
							},
							"autoconfig": {
								Type:        schema.TypeBool,
								Optional:    true,
								Computed:    computed,
								Description: "Autonomous address configuration flag.",
							},
						},
					},
				},
				"rdnss": {
					Optional: true,
					Computed: computed,
					MaxItems: 1,
					Type:     schema.TypeList,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"enabled": {
								Type:        schema.TypeBool,
								Optional:    true,
								Computed:    computed,
								Description: "Advertise Recursive DNS Server entries.",
							},
							"dnsservers": {
								Type:        schema.TypeList,
								Optional:    true,
								Computed:    computed,
								Description: "IPv6 DNS servers to advertise.",
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
							},
							"lifetime": {
								Type:         schema.TypeString,
								Optional:     true,
								Computed:     computed,
								ValidateFunc: validateNonNegativeIntString,
								Description:  "Lifetime of advertised RDNSS (DNS Server) entries in seconds. Use an empty value to use the platform default. Ignored when `infinite` is set.",
							},
							"infinite": {
								Type:        schema.TypeBool,
								Optional:    true,
								Computed:    computed,
								Description: "Advertise the RDNSS (DNS Server) lifetime as infinite. When true, `lifetime` is ignored.",
							},
						},
					},
				},
			},
		},
	}
}

// strVal dereferences a *string, returning "" for a nil pointer.
func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// customizeDiff normalizes a disabled dhcprelay/dhcpv6relay block. When
// enabled = false the other properties are meaningless, so they are forced to
// their zero values in the plan. This matches what read stores and what
// getDhcpRelay/getDhcpv6Relay send (the relay object with enabled = false and
// null properties), so a disabled block left in the configuration - even with
// stale vpcid/addresses - does not produce a perpetual diff. The blocks are
// Computed so SetNew is permitted on them.
func customizeDiff(d *schema.ResourceDiff, m interface{}) error {
	// SDK v1 can know a set while its nested access-mode values are unknown.
	// Defer only the affected collection, without relying on internal set hashes.
	known := func(key string) bool {
		if !d.NewValueKnown(key) {
			return false
		}
		for _, nestedKey := range d.GetChangedKeysPrefix(key + ".") {
			if strings.HasSuffix(nestedKey, ".#") || strings.HasSuffix(nestedKey, ".vlanid") || strings.HasSuffix(nestedKey, ".untagged") || strings.HasSuffix(nestedKey, ".accessmode") {
				if !d.NewValueKnown(nestedKey) {
					return false
				}
			}
		}
		return true
	}
	if err := validateDynamicVPCGroup(d.Get, known); err != nil {
		return err
	}

	for _, key := range []string{"dhcprelay", "dhcpv6relay"} {
		relays := d.Get(key).([]interface{})
		if len(relays) == 0 || relays[0] == nil {
			continue
		}
		r := relays[0].(map[string]interface{})
		if !r["enabled"].(bool) {
			err := d.SetNew(key, []interface{}{
				map[string]interface{}{
					"enabled":       false,
					"vpcid":         0,
					"primaryaddr":   "",
					"secondaryaddr": "",
				},
			})
			if err != nil {
				return err
			}
		}
	}

	if nd := firstBlock(d.Get("ipv6nd")); nd != nil {
		if ra := firstBlock(nd["routeradvertisement"]); ra != nil && (blockString(ra, "mode") == "default" || blockString(ra, "mode") == "disabled") {
			mode := blockString(ra, "mode")
			ra["routerlifetime"] = ""
			ra["advertisementinterval"] = ""
			ra["managedconfig"] = false
			ra["otherconfig"] = false

			if pa := firstBlock(nd["prefixadvertisement"]); pa != nil {
				pa["enabled"] = false
				pa["preferredlifetime"] = ""
				pa["validlifetime"] = ""
				pa["autoconfig"] = false
			}

			if r := firstBlock(nd["rdnss"]); r != nil {
				r["enabled"] = false
				r["dnsservers"] = []string{}
				r["lifetime"] = ""
				r["infinite"] = false
			}

			ra["mode"] = mode
			if err := d.SetNew("ipv6nd", []interface{}{nd}); err != nil {
				return err
			}
		}
	}

	return nil
}

// getDhcpRelay builds a DHCP Relay payload from the resource data. It returns
// nil only when no dhcprelay block is configured at all (the field is then
// omitted from the request). When the block is present but disabled, it returns
// an object with enabled = false and all other properties set to null, which is
// how the controller is told to disable the relay - sending the field as null
// (omitting it) does not disable anything.
func getDhcpRelay(d *schema.ResourceData) *vnet.VNetDhcpRelay {
	relays := d.Get("dhcprelay").([]interface{})
	if len(relays) == 0 || relays[0] == nil {
		return nil
	}
	r := relays[0].(map[string]interface{})
	if !r["enabled"].(bool) {
		return &vnet.VNetDhcpRelay{Enabled: false}
	}
	primary := r["primaryaddr"].(string)
	secondary := r["secondaryaddr"].(string)
	return &vnet.VNetDhcpRelay{
		Enabled:       true,
		Vpc:           &vnet.IDName{ID: r["vpcid"].(int)},
		PrimaryAddr:   &primary,
		SecondaryAddr: &secondary,
	}
}

// getDhcpv6Relay builds a DHCPv6 Relay payload from the resource data. It
// behaves exactly like getDhcpRelay but for the dhcpv6relay block: it returns
// nil only when no block is configured, an object with enabled = false and null
// properties when the block is present but disabled, and a fully populated
// object when enabled.
func getDhcpv6Relay(d *schema.ResourceData) *vnet.VNetDhcpv6Relay {
	relays := d.Get("dhcpv6relay").([]interface{})
	if len(relays) == 0 || relays[0] == nil {
		return nil
	}
	r := relays[0].(map[string]interface{})
	if !r["enabled"].(bool) {
		return &vnet.VNetDhcpv6Relay{Enabled: false}
	}
	primary := r["primaryaddr"].(string)
	secondary := r["secondaryaddr"].(string)
	return &vnet.VNetDhcpv6Relay{
		Enabled:       true,
		Vpc:           &vnet.IDName{ID: r["vpcid"].(int)},
		PrimaryAddr:   &primary,
		SecondaryAddr: &secondary,
	}
}

func intStringPtr(v string) *int {
	if v == "" {
		return nil
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return nil
	}
	return &i
}

func ptrIntString(i *int) string {
	if i == nil {
		return ""
	}
	return strconv.Itoa(*i)
}

func secondsObj(v string) *vnet.VNetIPv6NDSeconds {
	return &vnet.VNetIPv6NDSeconds{Seconds: intStringPtr(v)}
}

func lifetimeObj(v string, infinite bool) *vnet.VNetIPv6NDLifetime {
	seconds := intStringPtr(v)
	if infinite {
		seconds = nil
	}
	return &vnet.VNetIPv6NDLifetime{Seconds: seconds, Infinite: infinite}
}

func secondsStr(s *vnet.VNetIPv6NDSeconds) string {
	if s == nil {
		return ""
	}
	return ptrIntString(s.Seconds)
}

func lifetimeStr(l *vnet.VNetIPv6NDLifetime) string {
	if l == nil {
		return ""
	}
	return ptrIntString(l.Seconds)
}

func lifetimeInfinite(l *vnet.VNetIPv6NDLifetime) bool {
	if l == nil {
		return false
	}
	return l.Infinite
}

func firstBlock(v interface{}) map[string]interface{} {
	blocks, ok := v.([]interface{})
	if !ok || len(blocks) == 0 || blocks[0] == nil {
		return nil
	}
	block, ok := blocks[0].(map[string]interface{})
	if !ok {
		return nil
	}
	return block
}

func blockString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func blockBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}

func getIPv6ND(d *schema.ResourceData) *vnet.VNetIPv6ND {
	nd := firstBlock(d.Get("ipv6nd"))
	if nd == nil {
		return nil
	}

	ra := firstBlock(nd["routeradvertisement"])
	if ra == nil {
		ra = map[string]interface{}{}
	}
	pa := firstBlock(nd["prefixadvertisement"])
	if pa == nil {
		pa = map[string]interface{}{}
	}
	rdnss := firstBlock(nd["rdnss"])
	if rdnss == nil {
		rdnss = map[string]interface{}{}
	}

	mode := "default"
	if v := blockString(ra, "mode"); v != "" {
		mode = v
	}

	if mode == "default" || mode == "disabled" {
		return &vnet.VNetIPv6ND{
			RouterAdvertisement: &vnet.VNetIPv6NDRouterAdvertisement{
				Mode: mode,
			},
			PrefixAdvertisement: &vnet.VNetIPv6NDPrefixAdvertisement{},
			RDNSS:               &vnet.VNetIPv6NDRDNSS{},
		}
	}

	dnsServers := []string{}
	if servers, ok := rdnss["dnsservers"].([]interface{}); ok {
		for _, server := range servers {
			dnsServers = append(dnsServers, server.(string))
		}
	}

	return &vnet.VNetIPv6ND{
		RouterAdvertisement: &vnet.VNetIPv6NDRouterAdvertisement{
			Mode:                  mode,
			RouterLifetime:        secondsObj(blockString(ra, "routerlifetime")),
			AdvertisementInterval: secondsObj(blockString(ra, "advertisementinterval")),
			ManagedConfig:         blockBool(ra, "managedconfig"),
			OtherConfig:           blockBool(ra, "otherconfig"),
		},
		PrefixAdvertisement: &vnet.VNetIPv6NDPrefixAdvertisement{
			Enabled:           blockBool(pa, "enabled"),
			PreferredLifetime: secondsObj(blockString(pa, "preferredlifetime")),
			ValidLifetime:     secondsObj(blockString(pa, "validlifetime")),
			Autoconfig:        blockBool(pa, "autoconfig"),
		},
		RDNSS: &vnet.VNetIPv6NDRDNSS{
			Enabled:    blockBool(rdnss, "enabled"),
			DNSServers: dnsServers,
			Lifetime:   lifetimeObj(blockString(rdnss, "lifetime"), blockBool(rdnss, "infinite")),
		},
	}
}

func flattenIPv6ND(nd *vnet.VNetIPv6ND) []map[string]interface{} {
	if nd == nil {
		return nil
	}
	ra := nd.RouterAdvertisement
	if ra == nil {
		ra = &vnet.VNetIPv6NDRouterAdvertisement{Mode: "default"}
	}
	pa := nd.PrefixAdvertisement
	if pa == nil {
		pa = &vnet.VNetIPv6NDPrefixAdvertisement{}
	}
	rdnss := nd.RDNSS
	if rdnss == nil {
		rdnss = &vnet.VNetIPv6NDRDNSS{}
	}

	if ra.Mode == "default" || ra.Mode == "disabled" {
		pa = &vnet.VNetIPv6NDPrefixAdvertisement{}
		rdnss = &vnet.VNetIPv6NDRDNSS{}
		ra = &vnet.VNetIPv6NDRouterAdvertisement{Mode: ra.Mode}
	}

	return []map[string]interface{}{
		{
			"routeradvertisement": []map[string]interface{}{
				{
					"mode":                  ra.Mode,
					"routerlifetime":        secondsStr(ra.RouterLifetime),
					"advertisementinterval": secondsStr(ra.AdvertisementInterval),
					"managedconfig":         ra.ManagedConfig,
					"otherconfig":           ra.OtherConfig,
				},
			},
			"prefixadvertisement": []map[string]interface{}{
				{
					"enabled":           pa.Enabled,
					"preferredlifetime": secondsStr(pa.PreferredLifetime),
					"validlifetime":     secondsStr(pa.ValidLifetime),
					"autoconfig":        pa.Autoconfig,
				},
			},
			"rdnss": []map[string]interface{}{
				{
					"enabled":    rdnss.Enabled,
					"dnsservers": rdnss.DNSServers,
					"lifetime":   lifetimeStr(rdnss.Lifetime),
					"infinite":   lifetimeInfinite(rdnss.Lifetime),
				},
			},
		},
	}
}

func DiffSuppress(k, old, new string, d *schema.ResourceData) bool {
	return true
}

func resourceCreate(d *schema.ResourceData, m interface{}) error {
	if err := validateDynamicVPCGroup(d.Get, func(string) bool { return true }); err != nil {
		return err
	}
	l3vpn := d.Get("l3vpn").(bool)
	groupID := d.Get("dynamicvpcgroupid").(int)

	clientset := m.(*api.Clientset)

	sites := d.Get("sites").([]interface{})
	vlanid := d.Get("vlanid").(string)
	vnetTypeOne := false

	var sitesList []map[string]interface{}
	for _, site := range sites {
		sitesList = append(sitesList, site.(map[string]interface{}))
	}

	siteIDs := []vnet.VNetAddSite{}
	members := []vnet.VNetAddPort{}
	gatewayList := []vnet.VNetAddGateway{}
	portTagsList := []vnet.VNetPortTag{}

	tagsList := d.Get("tags").(*schema.Set).List()
	vpcid := d.Get("vpcid").(int)
	vxlanid := d.Get("vxlanid").(int)
	tags := []string{}
	for _, tag := range tagsList {
		tags = append(tags, tag.(string))
	}

	for _, site := range sitesList {
		if siteID, ok := site["id"]; ok {
			siteIDs = append(siteIDs, vnet.VNetAddSite{ID: siteID.(int)})
		}
		if gws, ok := site["gateways"]; ok {
			gateways := gws.(*schema.Set).List()

			for _, gw := range gateways {
				gateway := gw.(map[string]interface{})
				gwAdd := vnet.VNetAddGateway{
					Prefix: gateway["prefix"].(string),
					Vlan:   gateway["vlanid"].(string),
				}
				if dhcp := gateway["dhcp"].(string); dhcp == "enabled" {
					gwAdd.DHCPEnabled = true
					gwAdd.DHCPLeaseCount = 2
					if gateway["dhcpstartip"].(string) != "" {
						gwAdd.DHCP = &vnet.VNetGatewayDHCP{
							OptionSet: vnet.IDName{ID: gateway["dhcpoptionsetid"].(int)},
							Start:     gateway["dhcpstartip"].(string),
							End:       gateway["dhcpendip"].(string),
						}
					}
				}
				gatewayList = append(gatewayList, gwAdd)
			}
		}
		if p, ok := site["interface"]; ok {
			ports := p.(*schema.Set).List()
			if len(ports) > 0 {
				for _, p := range ports {
					port := p.(map[string]interface{})
					accessMode := false
					if port["untagged"].(string) == "yes" {
						accessMode = true
					}
					vID := vlanid
					if v := port["vlanid"].(string); v != "1" || vlanid == "" {
						vID = v
						vnetTypeOne = true
						if vID == "1" {
							accessMode = true
						}
					}
					if vlanid != "" && port["untagged"].(string) != "no" {
						accessMode = true
					}
					members = append(members, vnet.VNetAddPort{
						AccessMode: accessMode,
						ID:         port["id"].(int),
						Name:       port["name"].(string),
						Vlan:       vID,
						State:      "active",
					})
				}
			} else if p, ok := site["ports"]; ok {
				ports := p.(*schema.Set).List()
				for _, p := range ports {
					port := p.(map[string]interface{})
					accessMode := false
					if port["untagged"].(string) == "yes" {
						accessMode = true
					}
					vID := vlanid
					if v := port["vlanid"].(string); v != "1" || vlanid == "" {
						vID = v
						vnetTypeOne = true
						if vID == "1" {
							accessMode = true
						}
					}
					if vlanid != "" && port["untagged"].(string) != "no" {
						accessMode = true
					}
					members = append(members, vnet.VNetAddPort{
						AccessMode: accessMode,
						ID:         port["id"].(int),
						Name:       port["name"].(string),
						Vlan:       vID,
						State:      "active",
					})
				}
			}
		}
		if ifTags, ok := site["interfacetag"]; ok {
			ifaceTags := ifTags.(*schema.Set).List()

			for _, ifTag := range ifaceTags {
				ifaceTag := ifTag.(map[string]interface{})
				ifTagAdd := vnet.VNetPortTag{
					Name:       ifaceTag["tag"].(string),
					AccessMode: ifaceTag["accessmode"].(bool),
				}
				portTagsList = append(portTagsList, ifTagAdd)
			}
		}
	}

	var vlanidInterface any

	if vnetTypeOne || (groupID > 0 && vlanid == "" && len(members) == 0) {
		vlanidInterface = 0
	} else {
		if vlanid == "disabled" {
			vlanidInterface = 0
		} else {
			vlanidInterface = vlanid
		}
	}

	if groupID > 0 {
		for _, port := range members {
			if !port.AccessMode {
				return fmt.Errorf("dynamicvpcgroupid requires untagged member ports")
			}
		}
	}
	vnetAdd := &vnet.VNetAdd{
		L3VPN:        &l3vpn,
		Name:         d.Get("name").(string),
		Tenant:       vnet.VNetAddTenant{ID: d.Get("tenantid").(int)},
		GuestTenants: []vnet.VNetAddTenant{},
		Sites:        siteIDs,
		State:        d.Get("state").(string),
		IPFamily:     d.Get("ipfamily").(string),
		Gateways:     gatewayList,
		Ports:        members,
		Vlan:         vlanidInterface,
		Tags:         tags,
		VxlanID:      vxlanid,
		PortTags:     portTagsList,
	}

	if groupID > 0 {
		vnetAdd.DynamicVPCGroup = &vnet.DynamicVPCGroupRef{ID: groupID}
	} else if vpcid > 0 {
		vnetAdd.Vpc = &vnet.IDName{ID: vpcid}
	}

	vnetAdd.DhcpRelay = getDhcpRelay(d)
	vnetAdd.Dhcpv6Relay = getDhcpv6Relay(d)
	vnetAdd.IPv6ND = getIPv6ND(d)

	js, _ := json.Marshal(vnetAdd)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.VNet().Add(vnetAdd)
	if err != nil {
		log.Println("[DEBUG]", err)
		return err
	}

	js, _ = json.Marshal(reply)
	log.Println("[DEBUG]", string(js))

	log.Println("[DEBUG]", string(reply.Data))

	idStruct := struct {
		ID int `json:"id"`
	}{}

	data, err := reply.Parse()
	if err != nil {
		log.Println("[DEBUG]", err)
		return err
	}

	err = http.Decode(data.Data, &idStruct)
	if err != nil {
		log.Println("[DEBUG]", err)
		return err
	}

	log.Println("[DEBUG] ID:", idStruct.ID)

	if reply.StatusCode != 200 {
		return fmt.Errorf(string(reply.Data))
	}

	d.SetId(strconv.Itoa(idStruct.ID))
	if err := d.Set("l3vpn", l3vpn); err != nil {
		return err
	}
	return d.Set("dynamicvrf", groupID > 0)
}

func resourceRead(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	id, _ := strconv.Atoi(d.Id())
	vnetresp, err := clientset.VNet().GetByID(id)
	if err != nil {
		return nil
	}

	groupID := 0
	if vnetresp.DynamicVPCGroup != nil {
		groupID = vnetresp.DynamicVPCGroup.ID
	}
	if groupID > 0 || d.Get("dynamicvpcgroupid").(int) > 0 {
		err = d.Set("dynamicvpcgroupid", groupID)
		if err != nil {
			return err
		}
	}
	err = d.Set("l3vpn", vnetresp.L3VPN)
	if err != nil {
		return err
	}
	err = d.Set("dynamicvrf", vnetresp.DynamicVrf)
	if err != nil {
		return err
	}
	currentVpcId := d.Get("vpcid").(int)

	d.SetId(strconv.Itoa(vnetresp.ID))
	err = d.Set("name", vnetresp.Name)
	if err != nil {
		return err
	}
	err = d.Set("tenantid", vnetresp.Tenant.ID)
	if err != nil {
		return err
	}
	err = d.Set("state", vnetresp.State)
	if err != nil {
		return err
	}
	err = d.Set("ipfamily", vnetresp.IPFamily)
	if err != nil {
		return err
	}
	err = d.Set("ipv6nd", flattenIPv6ND(vnetresp.IPv6ND))
	if err != nil {
		return err
	}

	var subnets []*ipam.IPAM
	if currentVpcId > 0 {
		subnets, err = clientset.IPAM().GetSubnetsByVPC(currentVpcId)
	} else {
		subnets, err = clientset.IPAM().GetSubnets()
	}
	if err != nil {
		return err
	}

	hostsList := make(map[int][]*ipam.Host)

	sitesT := d.Get("sites").([]interface{})
	var sitesList []map[string]interface{}
	for _, site := range sitesT {
		sitesList = append(sitesList, site.(map[string]interface{}))
	}

	tportVlanIDMap := make(map[string]string)
	tportPortID := make(map[int]string)

	tPorts := make(map[string]map[string]interface{})
	tPortsId := make(map[int]map[string]interface{})
	interfaces := groupID > 0 && len(sitesT) == 0
	gatewayMap := make(map[string]map[string]interface{})
	ifaceTagsMap := make(map[string]map[string]interface{})

	for _, site := range sitesList {
		if gws, ok := site["gateways"]; ok {
			gateways := gws.(*schema.Set).List()
			for _, g := range gateways {
				gw := g.(map[string]interface{})
				gatewayMap[gw["prefix"].(string)] = gw
			}
		}
		if ifTags, ok := site["interfacetag"]; ok {
			ifaceTags := ifTags.(*schema.Set).List()
			for _, ifT := range ifaceTags {
				ifTag := ifT.(map[string]interface{})
				ifaceTagsMap[ifTag["tag"].(string)] = ifTag
			}
		}
		if p, ok := site["interface"]; ok {
			ports := p.(*schema.Set).List()
			if len(ports) > 0 {
				interfaces = true
				for _, p := range ports {
					port := p.(map[string]interface{})
					tportVlanIDMap[port["name"].(string)] = port["vlanid"].(string)
					tportPortID[port["id"].(int)] = port["vlanid"].(string)
					tPorts[port["name"].(string)] = port
					if port["id"].(int) != 0 {
						tPortsId[port["id"].(int)] = port
					}
				}
			} else if p, ok := site["ports"]; ok {
				ports := p.(*schema.Set).List()
				for _, p := range ports {
					port := p.(map[string]interface{})
					tportVlanIDMap[port["name"].(string)] = port["vlanid"].(string)
					tportPortID[port["id"].(int)] = port["vlanid"].(string)
					tPorts[port["name"].(string)] = port
					if port["id"].(int) != 0 {
						tPortsId[port["id"].(int)] = port
					}
				}
			}
		}
	}

	var sites []map[string]interface{}
	for _, site := range vnetresp.Sites {
		s := make(map[string]interface{})
		portList := make([]interface{}, 0)
		for _, port := range vnetresp.Ports {
			if port.Site.ID == site.ID {
				m := make(map[string]interface{})
				name := fmt.Sprintf("%s@%s", port.Port, port.SwitchName)
				if vlanFromTf, ok := tportVlanIDMap[name]; ok {
					m["name"] = name
					if vlanFromTf == "1" && vnetresp.Vlan != 0 && d.Get("vlanid").(string) != "" {
						port.Vlan = "1"
					}
				}
				if vlanFromTfId, ok := tportPortID[port.ID]; ok {
					m["id"] = port.ID
					if vlanFromTfId == "1" && vnetresp.Vlan != 0 && d.Get("vlanid").(string) != "" {
						port.Vlan = "1"
					}
				}
				if groupID > 0 && len(sitesT) == 0 {
					m["id"] = port.ID
					if port.Vlan != "1" && port.Untagged {
						m["untagged"] = "yes"
					}
				}
				m["vlanid"] = port.Vlan
				m["lacp"] = "off"
				if tPort, exists := tPorts[name]; exists {
					if tPortUntagged, ok := tPort["untagged"]; ok && tPortUntagged != "" {
						if groupID > 0 && vnetresp.Vlan == 0 && port.Vlan == "1" {
							m["untagged"] = tPortUntagged
							if tPortUntagged == "yes" {
								m["vlanid"] = tPort["vlanid"]
							}
						} else if port.Untagged {
							m["untagged"] = "yes"
						} else {
							m["untagged"] = "no"
						}
					}
				}
				if tPortid, exists := tPortsId[port.ID]; exists {
					if tPortUntagged, ok := tPortid["untagged"]; ok && tPortUntagged != "" {
						if groupID > 0 && vnetresp.Vlan == 0 && port.Vlan == "1" {
							m["untagged"] = tPortUntagged
							if tPortUntagged == "yes" {
								m["vlanid"] = tPortid["vlanid"]
							}
						} else if port.Untagged {
							m["untagged"] = "yes"
						} else {
							m["untagged"] = "no"
						}
					}
				}
				portList = append(portList, m)
			}
		}
		gatewayList := make([]interface{}, 0)
		for _, gateway := range vnetresp.Gateways {
			siteID := 0
			ip, ipNet, err := net.ParseCIDR(gateway.Prefix)
			if err != nil {
				return err
			}
			var hosts []*ipam.Host
			var ok bool
			subnet := subnet.GetByPrefix(subnets, ipNet.String())
			if hosts, ok = hostsList[subnet.ID]; !ok {
				var err error
				hosts, err = clientset.IPAM().GetHosts(subnet.ID)
				if err != nil {
					return err
				}
				hostsList[subnet.ID] = hosts
			}

			for _, host := range hosts {
				if ip.String() == host.Address {
					if len(subnet.Sites) > 0 {
						siteID = subnet.Sites[0].ID
					}
				}
			}
			if siteID == site.ID {
				if m, ok := gatewayMap[gateway.Prefix]; ok {
					m["prefix"] = gateway.Prefix
					m["vlanid"] = gateway.Vlan
					m["dhcp"] = "disabled"
					if gateway.DHCPEnabled {
						m["dhcp"] = "enabled"
						if m["dhcpstartip"].(string) != "" {
							m["dhcpoptionsetid"] = gateway.DHCP.OptionSet.ID
							m["dhcpstartip"] = gateway.DHCP.Start
							m["dhcpendip"] = gateway.DHCP.End
						}
					}
					gatewayList = append(gatewayList, m)
				}
			}
		}

		interfacePortsList := make([]interface{}, 0)
		for _, portTag := range vnetresp.PortTags {
			if m, ok := ifaceTagsMap[portTag.Name]; ok {
				m["tag"] = portTag.Name
				m["accessmode"] = portTag.AccessMode
				interfacePortsList = append(interfacePortsList, m)
			}
		}

		s["id"] = site.ID
		if interfaces {
			s["interface"] = portList
		} else {
			s["ports"] = portList
		}

		s["gateways"] = gatewayList
		s["interfacetag"] = interfacePortsList
		sites = append(sites, s)
	}

	if vnetresp.Vlan > 0 && d.Get("vlanid").(string) != "auto" {
		err = d.Set("vlanid", strconv.Itoa(vnetresp.Vlan))
		if err != nil {
			return err
		}
	}

	err = d.Set("sites", sites)
	if err != nil {
		return err
	}

	err = d.Set("tags", vnetresp.Tags)
	if err != nil {
		return err
	}

	err = d.Set("vxlanid", vnetresp.VxlanID)
	if err != nil {
		return err
	}

	if currentVpcId > 0 {
		err = d.Set("vpcid", vnetresp.Vpc.ID)
		if err != nil {
			return err
		}
	}

	var dhcpRelay []map[string]interface{}
	if vnetresp.DhcpRelay != nil {
		relayConfigured := len(d.Get("dhcprelay").([]interface{})) > 0
		if vnetresp.DhcpRelay.Enabled {
			vpcID := 0
			if vnetresp.DhcpRelay.Vpc != nil {
				vpcID = vnetresp.DhcpRelay.Vpc.ID
			}
			dhcpRelay = append(dhcpRelay, map[string]interface{}{
				"enabled":       true,
				"vpcid":         vpcID,
				"primaryaddr":   strVal(vnetresp.DhcpRelay.PrimaryAddr),
				"secondaryaddr": strVal(vnetresp.DhcpRelay.SecondaryAddr),
			})
		} else if relayConfigured {
			// Relay is disabled but a block is present in the configuration.
			// Store it normalized (other properties zeroed) so it matches the
			// planned value and does not produce a perpetual diff.
			dhcpRelay = append(dhcpRelay, map[string]interface{}{
				"enabled":       false,
				"vpcid":         0,
				"primaryaddr":   "",
				"secondaryaddr": "",
			})
		}
	}
	err = d.Set("dhcprelay", dhcpRelay)
	if err != nil {
		return err
	}

	var dhcpv6Relay []map[string]interface{}
	if vnetresp.Dhcpv6Relay != nil {
		relayConfigured := len(d.Get("dhcpv6relay").([]interface{})) > 0
		if vnetresp.Dhcpv6Relay.Enabled {
			vpcID := 0
			if vnetresp.Dhcpv6Relay.Vpc != nil {
				vpcID = vnetresp.Dhcpv6Relay.Vpc.ID
			}
			dhcpv6Relay = append(dhcpv6Relay, map[string]interface{}{
				"enabled":       true,
				"vpcid":         vpcID,
				"primaryaddr":   strVal(vnetresp.Dhcpv6Relay.PrimaryAddr),
				"secondaryaddr": strVal(vnetresp.Dhcpv6Relay.SecondaryAddr),
			})
		} else if relayConfigured {
			// Relay is disabled but a block is present in the configuration.
			// Store it normalized (other properties zeroed) so it matches the
			// planned value and does not produce a perpetual diff.
			dhcpv6Relay = append(dhcpv6Relay, map[string]interface{}{
				"enabled":       false,
				"vpcid":         0,
				"primaryaddr":   "",
				"secondaryaddr": "",
			})
		}
	}
	err = d.Set("dhcpv6relay", dhcpv6Relay)
	if err != nil {
		return err
	}

	return nil
}

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	if err := validateDynamicVPCGroup(d.Get, func(string) bool { return true }); err != nil {
		return err
	}
	l3vpn := d.Get("l3vpn").(bool)
	groupID := d.Get("dynamicvpcgroupid").(int)

	clientset := m.(*api.Clientset)

	sites := d.Get("sites").([]interface{})
	vlanid := d.Get("vlanid").(string)
	vnetTypeOne := false
	vxlanid := d.Get("vxlanid").(int)

	var sitesList []map[string]interface{}
	for _, site := range sites {
		sitesList = append(sitesList, site.(map[string]interface{}))
	}

	id, _ := strconv.Atoi(d.Id())
	v, err := clientset.VNet().GetByID(id)
	if err != nil {
		return nil
	}

	tagsList := d.Get("tags").(*schema.Set).List()
	tags := []string{}
	for _, tag := range tagsList {
		tags = append(tags, tag.(string))
	}

	siteIDs := []vnet.VNetUpdateSite{}
	members := []vnet.VNetUpdatePort{}
	gatewayList := []vnet.VNetUpdateGateway{}
	portTagsList := []vnet.VNetPortTag{}
	apiPorts := make(map[string]vnet.VNetDetailedPort)
	for _, p := range v.Ports {
		apiPorts[fmt.Sprintf("%s@%s", p.Port, p.SwitchName)] = p
	}

	existingVlanForAuto := ""
	for _, site := range sitesList {
		if siteID, ok := site["id"]; ok {
			siteIDs = append(siteIDs, vnet.VNetUpdateSite{ID: siteID.(int)})
		}
		if gws, ok := site["gateways"]; ok {
			gateways := gws.(*schema.Set).List()

			for _, gw := range gateways {
				gateway := gw.(map[string]interface{})
				gwUpdate := vnet.VNetUpdateGateway{
					Prefix: gateway["prefix"].(string),
					Vlan:   gateway["vlanid"].(string),
				}
				if dhcp := gateway["dhcp"].(string); dhcp == "enabled" {
					gwUpdate.DHCPEnabled = true
					gwUpdate.DHCPLeaseCount = 2
					if gateway["dhcpstartip"].(string) != "" {
						gwUpdate.DHCP = &vnet.VNetGatewayDHCP{
							OptionSet: vnet.IDName{ID: gateway["dhcpoptionsetid"].(int)},
							Start:     gateway["dhcpstartip"].(string),
							End:       gateway["dhcpendip"].(string),
						}
					}
				}
				gatewayList = append(gatewayList, gwUpdate)
			}
		}
		if p, ok := site["interface"]; ok {
			ports := p.(*schema.Set).List()
			if len(ports) > 0 {
				for _, p := range ports {
					port := p.(map[string]interface{})
					accessMode := false
					if port["untagged"].(string) == "yes" {
						accessMode = true
					}
					vID := vlanid
					if v := port["vlanid"].(string); v != "1" || vlanid == "" {
						vID = v
						vnetTypeOne = true
						if vID == "1" {
							accessMode = true
						}
					}
					if vlanid != "" && port["untagged"].(string) != "no" {
						accessMode = true
					}
					if portID, ok := apiPorts[port["name"].(string)]; ok {
						vl := vID
						if vlanid == "auto" {
							vl = portID.Vlan
							existingVlanForAuto = portID.Vlan
						}
						members = append(members, vnet.VNetUpdatePort{
							AccessMode: accessMode,
							ID:         portID.ID,
							Vlan:       vl,
							State:      "active",
						})
					} else {
						members = append(members, vnet.VNetUpdatePort{
							AccessMode: accessMode,
							Name:       port["name"].(string),
							ID:         port["id"].(int),
							Vlan:       vID,
							State:      "active",
						})
					}
				}
			} else if p, ok := site["ports"]; ok {
				ports := p.(*schema.Set).List()
				for _, p := range ports {
					port := p.(map[string]interface{})
					accessMode := false
					if port["untagged"].(string) == "yes" {
						accessMode = true
					}
					vID := vlanid
					if v := port["vlanid"].(string); v != "1" || vlanid == "" {
						vID = v
						vnetTypeOne = true
						if vID == "1" {
							accessMode = true
						}
					}
					if vlanid != "" && port["untagged"].(string) != "no" {
						accessMode = true
					}
					if portID, ok := apiPorts[port["name"].(string)]; ok {
						vl := vID
						if vlanid == "auto" {
							vl = portID.Vlan
							existingVlanForAuto = portID.Vlan
						}
						members = append(members, vnet.VNetUpdatePort{
							AccessMode: accessMode,
							ID:         portID.ID,
							Vlan:       vl,
							State:      "active",
						})
					} else {
						members = append(members, vnet.VNetUpdatePort{
							AccessMode: accessMode,
							Name:       port["name"].(string),
							ID:         port["id"].(int),
							Vlan:       vID,
							State:      "active",
						})
					}
				}
			}
		}
		if ifTags, ok := site["interfacetag"]; ok {
			ifaceTags := ifTags.(*schema.Set).List()

			for _, ifTag := range ifaceTags {
				ifaceTag := ifTag.(map[string]interface{})
				ifTagAdd := vnet.VNetPortTag{
					Name:       ifaceTag["tag"].(string),
					AccessMode: ifaceTag["accessmode"].(bool),
				}
				portTagsList = append(portTagsList, ifTagAdd)
			}
		}
	}

	log.Println("[DEBUG] ExistingVlanForAuto", existingVlanForAuto)
	if existingVlanForAuto != "" {
		newMembers := []vnet.VNetUpdatePort{}
		for _, m := range members {
			m.Vlan = existingVlanForAuto
			newMembers = append(newMembers, m)
		}
		members = newMembers
	}

	var vlanidInterface any

	if vnetTypeOne || (groupID > 0 && vlanid == "" && len(members) == 0) {
		vlanidInterface = 0
	} else {
		if vlanid == "disabled" {
			vlanidInterface = 0
		} else {
			vlanidInterface = vlanid
		}
	}

	if groupID > 0 {
		for i := range members {
			if !members[i].AccessMode {
				return fmt.Errorf("dynamicvpcgroupid requires untagged member ports")
			}
			members[i].Untagged = true
		}
	}
	vnetUpdate := &vnet.VNetUpdate{
		L3VPN:        &l3vpn,
		Name:         d.Get("name").(string),
		GuestTenants: []vnet.VNetUpdateGuestTenant{},
		Sites:        siteIDs,
		State:        d.Get("state").(string),
		IPFamily:     d.Get("ipfamily").(string),
		Gateways:     gatewayList,
		Ports:        members,
		Vlan:         vlanidInterface,
		Tags:         tags,
		VxlanID:      vxlanid,
		PortTags:     portTagsList,
		DhcpRelay:    getDhcpRelay(d),
		Dhcpv6Relay:  getDhcpv6Relay(d),
		IPv6ND:       getIPv6ND(d),
	}

	js, _ := json.Marshal(vnetUpdate)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.VNet().Update(id, vnetUpdate)
	if err != nil {
		log.Println("[DEBUG]", err)
		return err
	}

	if reply.StatusCode != 200 {
		return fmt.Errorf(string(reply.Data))
	}

	return nil
}

func resourceDelete(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	id, _ := strconv.Atoi(d.Id())
	reply, err := clientset.VNet().Delete(id)
	if err != nil {
		return err
	}

	if reply.StatusCode != 200 {
		return fmt.Errorf(string(reply.Data))
	}

	d.SetId("")
	return nil
}

func resourceExists(d *schema.ResourceData, m interface{}) (bool, error) {
	clientset := m.(*api.Clientset)

	id, _ := strconv.Atoi(d.Id())
	vnet, _ := clientset.VNet().GetByID(id)

	if vnet == nil {
		return false, nil
	}
	if vnet.ID > 0 {
		return true, nil
	}

	return false, nil
}

func resourceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	clientset := m.(*api.Clientset)

	vnets, _ := clientset.VNet().Get()
	name := d.Id()
	for _, vnet := range vnets {
		if vnet.Name == name {
			d.SetId(strconv.Itoa(vnet.ID))
			return []*schema.ResourceData{d}, nil
		}
	}

	return []*schema.ResourceData{d}, nil
}
