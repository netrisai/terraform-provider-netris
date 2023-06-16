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

	"github.com/netrisai/netriswebapi/http"
	"github.com/netrisai/netriswebapi/v2/types/ipam"
	"github.com/netrisai/netriswebapi/v2/types/vnet"
	"github.com/netrisai/terraform-provider-netris/netris/subnet"

	api "github.com/netrisai/netriswebapi/v2"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Only when global 'vlanid' is specified",
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
						"ports": {
							Optional:    true,
							Type:        schema.TypeSet,
							Description: "Block of ports",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
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
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Only when global 'vlanid' is specified",
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
		},
		Create: resourceCreate,
		Read:   resourceRead,
		Update: resourceUpdate,
		Delete: resourceDelete,
		Exists: resourceExists,
		Importer: &schema.ResourceImporter{
			State: resourceImport,
		},
	}
}

func DiffSuppress(k, old, new string, d *schema.ResourceData) bool {
	return true
}

func resourceCreate(d *schema.ResourceData, m interface{}) error {
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

	tagsList := d.Get("tags").(*schema.Set).List()
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
						Name:       port["name"].(string),
						Vlan:       vID,
						State:      "active",
					})
				}
			}

		}
	}

	var vlanidInterface any

	if vnetTypeOne {
		vlanidInterface = 0
	} else {
		vlanidInterface = vlanid
	}

	vnetAdd := &vnet.VNetAdd{
		Name:         d.Get("name").(string),
		Tenant:       vnet.VNetAddTenant{ID: d.Get("tenantid").(int)},
		GuestTenants: []vnet.VNetAddTenant{},
		Sites:        siteIDs,
		State:        d.Get("state").(string),
		Gateways:     gatewayList,
		Ports:        members,
		Vlan:         vlanidInterface,
		Tags:         tags,
	}

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
	return nil
}

func resourceRead(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	id, _ := strconv.Atoi(d.Id())
	vnetresp, err := clientset.VNet().GetByID(id)
	if err != nil {
		return nil
	}

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

	subnets, err := clientset.IPAM().GetSubnets()
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

	tPorts := make(map[string]map[string]interface{})
	interfaces := false
	gatewayMap := make(map[string]map[string]interface{})

	for _, site := range sitesList {
		if gws, ok := site["gateways"]; ok {
			gateways := gws.(*schema.Set).List()
			for _, g := range gateways {
				gw := g.(map[string]interface{})
				gatewayMap[gw["prefix"].(string)] = gw
			}
		}
		if p, ok := site["interface"]; ok {
			ports := p.(*schema.Set).List()
			if len(ports) > 0 {
				interfaces = true
				for _, p := range ports {
					port := p.(map[string]interface{})
					tportVlanIDMap[port["name"].(string)] = port["vlanid"].(string)
					tPorts[port["name"].(string)] = port
				}
			} else if p, ok := site["ports"]; ok {
				ports := p.(*schema.Set).List()
				for _, p := range ports {
					port := p.(map[string]interface{})
					tportVlanIDMap[port["name"].(string)] = port["vlanid"].(string)
					tPorts[port["name"].(string)] = port
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
					if vlanFromTf == "1" && vnetresp.Vlan != 0 && d.Get("vlanid").(string) != "" {
						port.Vlan = "1"
					}
				}
				m["name"] = name
				m["vlanid"] = port.Vlan
				m["lacp"] = "off"
				if tPortUntagged := tPorts[name]["untagged"]; tPortUntagged != "" {
					if port.Untagged {
						m["untagged"] = "yes"
					} else {
						m["untagged"] = "no"
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
				m := gatewayMap[gateway.Prefix]
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
		s["id"] = site.ID
		if interfaces {
			s["interface"] = portList
		} else {
			s["ports"] = portList
		}

		s["gateways"] = gatewayList
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

	return nil
}

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	sites := d.Get("sites").([]interface{})
	vlanid := d.Get("vlanid").(string)
	vnetTypeOne := false

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
							Vlan:       vID,
							State:      "active",
						})
					}
				}
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

	if vnetTypeOne {
		vlanidInterface = 0
	} else {
		vlanidInterface = vlanid
	}

	vnetUpdate := &vnet.VNetUpdate{
		Name:         d.Get("name").(string),
		GuestTenants: []vnet.VNetUpdateGuestTenant{},
		Sites:        siteIDs,
		State:        d.Get("state").(string),
		Gateways:     gatewayList,
		Ports:        members,
		Vlan:         vlanidInterface,
		Tags:         tags,
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
