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

package inventoryprofile

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/netrisai/netriswebapi/v1/types/inventoryprofile"

	api "github.com/netrisai/netriswebapi/v2"
)

func DataResource() *schema.Resource {
	return &schema.Resource{
		Description: "Data Source: inventory profiles",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of inventory profile",
			},
			"description": {
				Computed:    true,
				Optional:    true,
				Type:        schema.TypeString,
				Description: "Inventory profile description",
			},
			"ipv4ssh": {
				Optional:    true,
				Type:        schema.TypeList,
				Description: "List of IPv4 subnets allowed to ssh.",
				Elem: &schema.Schema{
					ValidateFunc: validateIP,
					Type:         schema.TypeString,
				},
			},
			"ipv6ssh": {
				Optional:    true,
				Type:        schema.TypeList,
				Description: "List of IPv6 subnets allowed to ssh.",
				Elem: &schema.Schema{
					ValidateFunc: validateIP,
					Type:         schema.TypeString,
				},
			},
			"timezone": {
				ValidateFunc: validateTimeZone,
				Optional:     true,
				Type:         schema.TypeString,
				Description:  "Devices using this inventory profile will adjust their system time to the selected timezone. Valid value is a name from the TZ [database](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones).",
			},
			"ntpservers": {
				Optional:    true,
				Type:        schema.TypeList,
				Description: "List of domain names or IP addresses of NTP servers.",
				Elem: &schema.Schema{
					ValidateFunc: validateNTP,
					Type:         schema.TypeString,
				},
			},
			"dnsservers": {
				Optional:    true,
				Type:        schema.TypeList,
				Description: "List of IP addresses of DNS servers.",
				Elem: &schema.Schema{
					ValidateFunc: validateIP,
					Type:         schema.TypeString,
				},
			},
			"customrule": {
				Optional:    true,
				Type:        schema.TypeList,
				Description: "Custom Rules configuration block. User defined rules to allow certain traffic.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"sourcesubnet": {
							ValidateFunc: validateIPPrefix,
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Source Subnet.",
						},
						"srcport": {
							ValidateFunc: validatePort,
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Source port. 1-65535, or empty for any.",
						},
						"description": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Custom rule's description.",
						},
						"dstport": {
							ValidateFunc: validatePort,
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Destination port. 1-65535, or empty for any.",
						},
						"protocol": {
							ValidateFunc: validateProtocol,
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Protocol.",
						},
					},
				},
			},
			"fabricsettings": {
				Optional:    true,
				Type:        schema.TypeSet,
				Description: "Fabric Settings.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"optimisebgpoverlay": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Optimize BGP Overlay for leaf-spine topology.",
						},
						"optimisebgpoverlayhypervisor": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Required for BGP/EVPN VXLAN integration with compute hypervisor networking.",
						},
						"unnumberedbgpunderlay": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "BGP underlay using p2p IPv4 from link objects versus unnumbered.",
						},
						"automaticlinkaggregation": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Automatic single-legged link aggregation on non-backbone ports.",
						},
						"mclag": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Enable MC-LAG (disables EVPN-MH on the same switches).",
						},
						"serverbasedesi": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Enable server based ESI generation.",
						},
					},
				},
			},
			"gpuclustersettings": {
				Optional:    true,
				Type:        schema.TypeSet,
				Description: "GPU cluster specific settings.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"qosandroce": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Optimize for RDMA over Converged Ethernet.",
						},
						"roceadaptiverouting": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Enable Adaptive Routing for RoCE.",
						},
						"congestioncontrol": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Enable Zero Touch RoCE Congestion Control.",
						},
						"asicmonitoring": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Enable ASIC monitoring.",
						},
						"hwmp": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Enable HWMP (Hardware Multi Plane).",
						},
						"aggregatel3vpnprefix": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Minimize prefix updates over BGP overlay for L3VPN p2p links. Deprecated: derived automatically from `refarch`.",
						},
						"refarch": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "GPU cluster reference architecture (gpuClusterProps.refArch).",
						},
					},
				},
			},
			"snmpv2": {
				Optional:    true,
				Type:        schema.TypeSet,
				Description: "SNMPv2 Settings",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Enable SNMPv2 on inventory devices.",
						},
						"community": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "SNMPv2 read-only community string.",
						},
						"ipv4_list": {
							Optional:    true,
							Type:        schema.TypeList,
							Description: "List of IPv4 addresses/prefixes allowed to poll SNMPv2.",
							Elem: &schema.Schema{
								ValidateFunc: validateIPPrefix,
								Type:         schema.TypeString,
							},
						},
						"contact": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "SNMPv2 contact field.",
						},
						"location": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "SNMPv2 location field.",
						},
					},
				},
			},
			"ztpsettings": {
				Optional:    true,
				Type:        schema.TypeSet,
				Description: "ZTP settings for inventory profile.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"nos_image": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "NOS image file name.",
						},
						"password": {
							Type:        schema.TypeString,
							Optional:    true,
							Sensitive:   true,
							Description: "NOS admin password for ZTP.",
						},
					},
				},
			},
			"netqsettings": {
				Optional:    true,
				Type:        schema.TypeSet,
				Description: "NetQ settings for inventory profile.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Whether NetQ is enabled.",
						},
						"server_addrs": {
							Optional:    true,
							Type:        schema.TypeList,
							Description: "List of NetQ server addresses (IP addresses or domain names).",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"server_port": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "NetQ server port.",
						},
					},
				},
			},
			"syslog_destinations": {
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Description: "Syslog Destinations settings for inventory profile.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Whether syslog forwarding is enabled.",
						},
						"use_rfc5424": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Whether forwarded messages are formatted per RFC 5424.",
						},
						"servers": {
							Optional:    true,
							Type:        schema.TypeList,
							Description: "Syslog destination.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"host": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "IPv4, IPv6, or Fully Qualified Domain Name of the syslog destination.",
									},
									"port": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "Syslog destination port.",
									},
									"protocol": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Transport protocol (syslogDestinationItem.protocol).",
									},
									"severity": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Minimum severity level to forward (syslogDestinationItem.severity).",
									},
								},
							},
						},
					},
				},
			},
			"aaa": {
				Optional:    true,
				Type:        schema.TypeList,
				Description: "AAA (RADIUS) login authentication configuration for devices that use this inventory profile.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"authorder": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Order in which authentication backends are attempted.",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"radius": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "RADIUS authentication settings.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Type:        schema.TypeBool,
										Optional:    true,
										Description: "Whether RADIUS authentication is enabled.",
									},
									"server": {
										Type:        schema.TypeList,
										Optional:    true,
										Description: "RADIUS server.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"host": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "IPv4 address or Fully Qualified Domain Name of the RADIUS server.",
												},
												"port": {
													Type:        schema.TypeInt,
													Optional:    true,
													Description: "RADIUS server port.",
												},
												"priority": {
													Type:        schema.TypeInt,
													Optional:    true,
													Description: "Priority of this RADIUS server relative to others on the profile.",
												},
												"authtype": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "RADIUS authentication protocol.",
												},
												"secret": {
													Type:        schema.TypeString,
													Optional:    true,
													Sensitive:   true,
													Description: "Shared secret used to authenticate with this RADIUS server. Write-only: never returned in cleartext by the API.",
												},
											},
										},
									},
								},
							},
						},
						"local": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Local admin-account authentication settings.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Type:        schema.TypeBool,
										Optional:    true,
										Description: "Whether local admin-account authentication is enabled.",
									},
								},
							},
						},
					},
				},
			},
		},
		Read:   dataResourceRead,
		Exists: dataResourceExists,
		Importer: &schema.ResourceImporter{
			State: dataRresourceImport,
		},
	}
}

func dataResourceRead(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)
	var profile *inventoryprofile.Profile
	var ok bool
	profile, ok = findByName(d.Get("name").(string), clientset)
	if !ok {
		return fmt.Errorf("Coudn't find inventory profile '%s'", d.Get("name").(string))
	}

	d.SetId(strconv.Itoa(profile.ID))
	err := d.Set("name", profile.Name)
	if err != nil {
		return err
	}
	err = d.Set("description", profile.Description)
	if err != nil {
		return err
	}
	if (d.Get("ipv4ssh") != nil && len(d.Get("ipv4ssh").([]interface{})) > 0) || profile.Ipv4SSH != "" {
		err = d.Set("ipv4ssh", strings.Split(profile.Ipv4SSH, ","))
		if err != nil {
			return err
		}
	}
	if (d.Get("ipv6ssh") != nil && len(d.Get("ipv6ssh").([]interface{})) > 0) || profile.Ipv6SSH != "" {
		err = d.Set("ipv6ssh", strings.Split(profile.Ipv6SSH, ","))
		if err != nil {
			return err
		}
	}

	err = d.Set("timezone", effectiveTimezoneForState(profile.Timezone))
	if err != nil {
		return err
	}
	err = d.Set("ntpservers", strings.Split(profile.NTPServers, ","))
	if err != nil {
		return err
	}
	err = d.Set("dnsservers", strings.Split(profile.DNSServers, ","))
	if err != nil {
		return err
	}

	var customRules []map[string]interface{}
	for _, rule := range profile.CustomRules {
		customRule := make(map[string]interface{})
		customRule["sourcesubnet"] = rule.SrcSubnet
		customRule["srcport"] = rule.SrcPort
		customRule["dstport"] = rule.DstPort
		customRule["description"] = rule.Description
		customRule["protocol"] = rule.Protocol
		customRules = append(customRules, customRule)
	}

	err = d.Set("customrule", customRules)
	if err != nil {
		return err
	}

	var fabricsettingsList []map[string]interface{}
	fabricsettings := make(map[string]interface{})
	fabricsettings["optimisebgpoverlay"] = profile.FabricProps.OptimiseBgpOverlay
	fabricsettings["optimisebgpoverlayhypervisor"] = profile.FabricProps.OptimiseBgpOverlayHypervisor
	fabricsettings["unnumberedbgpunderlay"] = profile.FabricProps.UnnumberedBgpUnderlay
	fabricsettings["automaticlinkaggregation"] = profile.FabricProps.AutomaticLinkAggregation
	fabricsettings["mclag"] = profile.FabricProps.MCLag
	fabricsettings["serverbasedesi"] = profile.FabricProps.ServerBasedESI
	fabricsettingsList = append(fabricsettingsList, fabricsettings)

	var gpuclustersettingsList []map[string]interface{}
	gpuclustersettings := make(map[string]interface{})
	gpuclustersettings["qosandroce"] = profile.GpuClusterProps.Roce
	gpuclustersettings["roceadaptiverouting"] = profile.GpuClusterProps.RoceAdaptiveRouting
	gpuclustersettings["congestioncontrol"] = profile.GpuClusterProps.CongestionControl
	gpuclustersettings["asicmonitoring"] = profile.GpuClusterProps.AsicMonitoring
	gpuclustersettings["hwmp"] = profile.GpuClusterProps.Hwmp
	gpuclustersettings["aggregatel3vpnprefix"] = aggregateL3VpnPrefixForRefArch(profile.GpuClusterProps.RefArch)
	gpuclustersettings["refarch"] = profile.GpuClusterProps.RefArch
	gpuclustersettingsList = append(gpuclustersettingsList, gpuclustersettings)

	err = d.Set("fabricsettings", fabricsettingsList)
	if err != nil {
		return err
	}
	err = d.Set("gpuclustersettings", gpuclustersettingsList)
	if err != nil {
		return err
	}

	var snmpv2List []map[string]interface{}
	snmpv2 := make(map[string]interface{})
	snmpv2["enabled"] = profile.SNMPv2Props.Enabled
	snmpv2["community"] = profile.SNMPv2Props.Community
	snmpv2["contact"] = profile.SNMPv2Props.Contact
	snmpv2["location"] = profile.SNMPv2Props.Location
	snmpv2["ipv4_list"] = profile.SNMPv2Props.Ipv4List
	snmpv2List = append(snmpv2List, snmpv2)

	var ztpsettingsList []map[string]interface{}
	ztpsettings := make(map[string]interface{})
	ztpsettings["nos_image"] = profile.ZTPProps.NOSImage
	ztpsettings["password"] = profile.ZTPProps.Password
	ztpsettingsList = append(ztpsettingsList, ztpsettings)

	var netqsettingsList []map[string]interface{}
	if profile.NetQProps.Enabled || len(profile.NetQProps.ServerAddrs) > 0 {
		netqsettings := make(map[string]interface{})
		netqsettings["enabled"] = profile.NetQProps.Enabled
		netqsettings["server_addrs"] = profile.NetQProps.ServerAddrs
		netqsettings["server_port"] = int(profile.NetQProps.ServerPort)
		netqsettingsList = append(netqsettingsList, netqsettings)
	}

	var syslogDestinationsList []map[string]interface{}
	if profile.SyslogDestinations.Enabled || len(profile.SyslogDestinations.Servers) > 0 {
		syslogDestinationsList = append(syslogDestinationsList, syslogDestinationsToMap(profile.SyslogDestinations))
	}

	err = d.Set("snmpv2", snmpv2List)
	if err != nil {
		return err
	}
	err = d.Set("ztpsettings", ztpsettingsList)
	if err != nil {
		return err
	}
	err = d.Set("netqsettings", netqsettingsList)
	if err != nil {
		return err
	}
	err = d.Set("syslog_destinations", syslogDestinationsList)
	if err != nil {
		return err
	}

	aaaList := []map[string]interface{}{aaaToMap(profile.AAAProps, existingRadiusServersByHostPort(d))}
	err = d.Set("aaa", aaaList)
	if err != nil {
		return err
	}

	return nil
}

func dataResourceExists(d *schema.ResourceData, m interface{}) (bool, error) {
	return true, nil
}

func dataRresourceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	clientset := m.(*api.Clientset)
	name := d.Id()
	var profile *inventoryprofile.Profile
	var ok bool
	profile, ok = findByName(name, clientset)
	if !ok {
		return []*schema.ResourceData{d}, fmt.Errorf("Coudn't find inventory profile '%s'", d.Get("name").(string))
	}
	d.SetId(strconv.Itoa(profile.ID))

	return []*schema.ResourceData{d}, nil
}
