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
	"fmt"
	"net"
	"strconv"

	"github.com/netrisai/netriswebapi/v2/types/ipam"
	"github.com/netrisai/netriswebapi/v2/types/vnet"
	"github.com/netrisai/terraform-provider-netris/netris/subnet"

	api "github.com/netrisai/netriswebapi/v2"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func DataResource() *schema.Resource {
	return &schema.Resource{
		Description: "Data Source: Vnets",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the vnet.",
			},
			"tenantid": {
				Computed:    true,
				Optional:    true,
				Type:        schema.TypeInt,
				Description: "ID of tenant. Users of this tenant will be permitted to edit this unit.",
			},
			"state": {
				Computed:    true,
				Optional:    true,
				Type:        schema.TypeString,
				Description: "V-Net state.",
			},
			"sites": {
				Computed:    true,
				Optional:    true,
				Type:        schema.TypeList,
				Description: "Block of per site vnet configuration.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "The site ID. Ports from these sites will be allowed to participate in the V-Net.",
						},
						"ports": {
							Optional:    true,
							Type:        schema.TypeList,
							Description: "Block of ports",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Switch port name.",
									},
									"vlanid": {
										Default:     "1",
										Type:        schema.TypeString,
										Optional:    true,
										Description: "VLAN tag for current port.",
									},
								},
							},
						},
						"gateways": {
							Optional:    true,
							Type:        schema.TypeList,
							Description: "Block of gateways.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"prefix": {
										ValidateFunc: validateGateway,
										Type:         schema.TypeString,
										Required:     true,
										Description:  "The address will be serving as anycast default gateway for selected subnet.",
									},
									"vlanid": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"vpcid": {
				ForceNew:    true,
				Optional:    true,
				Type:        schema.TypeInt,
				Description: "ID of VPC.",
			},
		},
		Read:   dataResourceRead,
		Exists: dataResourceExists,
	}
}

func dataResourceRead(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	name := d.Get("name").(string)

	var VNet *vnet.VNet

	vnets, err := clientset.VNet().Get()
	if err != nil {
		return err
	}

	for _, v := range vnets {
		if v.Name == name {
			VNet = v
			break
		}
	}

	if VNet == nil {
		return fmt.Errorf("coudn't find vnet %s", name)
	}

	vnet, err := clientset.VNet().GetByID(VNet.ID)
	if err != nil {
		return err
	}

	d.SetId(strconv.Itoa(vnet.ID))
	err = d.Set("name", vnet.Name)
	if err != nil {
		return err
	}
	err = d.Set("tenantid", vnet.Tenant.ID)
	if err != nil {
		return err
	}
	err = d.Set("state", vnet.State)
	if err != nil {
		return err
	}

	subnets, err := clientset.IPAM().GetSubnets()
	if err != nil {
		return err
	}

	hostsList := make(map[int][]*ipam.Host)

	var sites []map[string]interface{}
	for _, site := range vnet.Sites {
		s := make(map[string]interface{})
		portList := make([]interface{}, 0)
		for _, port := range vnet.Ports {
			if port.Site.ID == site.ID {
				m := make(map[string]interface{})
				m["name"] = fmt.Sprintf("%s@%s", port.Port, port.SwitchName)
				m["vlanid"] = port.Vlan
				portList = append(portList, m)
			}
		}
		gatewayList := make([]interface{}, 0)
		for _, gateway := range vnet.Gateways {
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
				m := make(map[string]interface{})
				m["prefix"] = gateway.Prefix
				m["vlanid"] = gateway.Vlan
				gatewayList = append(gatewayList, m)
			}
		}
		s["id"] = site.ID
		s["ports"] = portList
		s["gateways"] = gatewayList
		sites = append(sites, s)
	}

	err = d.Set("sites", sites)
	if err != nil {
		return err
	}

	err = d.Set("vpcid", vnet.Vpc.ID)
	if err != nil {
		return err
	}

	return nil
}

func dataResourceExists(d *schema.ResourceData, m interface{}) (bool, error) {
	return true, nil
}
