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

package bgp

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/netrisai/netriswebapi/http"
	"github.com/netrisai/netriswebapi/v2/types/bgp"

	api "github.com/netrisai/netriswebapi/v2"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"bgpid": {
				Type:             schema.TypeInt,
				Optional:         true,
				Description:      "The name of the resource, also acts as it's unique ID",
				DiffSuppressFunc: DiffSuppress,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the resource, also acts as it's unique ID",
			},
			"site": {
				Required:    true,
				Type:        schema.TypeString,
				Description: "A description of an item",
			},
			"hardware": {
				Optional:    true,
				Type:        schema.TypeString,
				Description: "A description of an item",
			},
			"neighboras": {
				Optional:    true,
				Type:        schema.TypeInt,
				Description: "A description of an item",
			},
			"transport": {
				Optional:    true,
				Type:        schema.TypeMap,
				Description: "Switch Ports",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"localip": {
				ValidateFunc: validateIPPrefix,
				Required:     true,
				Type:         schema.TypeString,
				Description:  "Local IP. Example 10.0.1.1/24",
			},
			"remoteip": {
				ValidateFunc: validateIPPrefix,
				Required:     true,
				Type:         schema.TypeString,
				Description:  "Remote IP. Example 10.0.1.2/24",
			},
			"description": {
				Optional:    true,
				Type:        schema.TypeString,
				Description: "A description of an item",
			},
			"state": {
				Optional:     true,
				Default:      "enabled",
				ValidateFunc: validateState,
				Type:         schema.TypeString,
				Description:  "A description of an item",
			},
			"multihop": {
				Optional:    true,
				Type:        schema.TypeMap,
				Description: "Multihop",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validateMultihop,
				},
			},
			"bgppassword": {
				Optional:    true,
				Type:        schema.TypeString,
				Description: "A description of an item",
			},
			"allowasin": {
				Optional:    true,
				Type:        schema.TypeInt,
				Description: "A description of an item",
			},
			"defaultoriginate": {
				Optional:    true,
				Type:        schema.TypeBool,
				Description: "A description of an item",
			},
			"prefixinboundmax": {
				Optional:    true,
				Type:        schema.TypeString,
				Description: "A description of an item",
			},
			"inboundroutemap": {
				Optional:    true,
				Type:        schema.TypeString,
				Description: "A description of an item",
			},
			"outboundroutemap": {
				Optional:    true,
				Type:        schema.TypeString,
				Description: "A description of an item",
			},
			"localpreference": {
				Optional:    true,
				Type:        schema.TypeInt,
				Description: "A description of an item",
			},
			"weight": {
				Optional:    true,
				Type:        schema.TypeInt,
				Description: "A description of an item",
			},
			"prependinbound": {
				Optional:    true,
				Type:        schema.TypeInt,
				Description: "A description of an item",
			},
			"prependoutbound": {
				Optional:    true,
				Type:        schema.TypeInt,
				Description: "A description of an item",
			},
			"prefixlistinbound": {
				Optional:    true,
				Type:        schema.TypeList,
				Description: "Switch Ports",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"prefixlistoutbound": {
				Optional:    true,
				Type:        schema.TypeList,
				Description: "Switch Ports",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"sendbgpcommunity": {
				Optional:    true,
				Type:        schema.TypeList,
				Description: "Switch Ports",
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

	var (
		vlanID    int
		state     = "enabled"
		ipVersion = "ipv6"
		hwID      = 0
		port      = ""
		vnetID    = 0
	)

	originate := "disabled"
	localPreference := 100

	siteName := d.Get("site").(string)

	if d.Get("defaultoriginate").(bool) {
		originate = "enabled"
	}

	hardware := d.Get("softgate").(string)

	inventory, err := clientset.Inventory().Get()
	if err != nil {
		return err
	}

	for _, hw := range inventory {
		if hw.Name == hardware && hardware != "auto" {
			hwID = hw.ID
		}
	}

	transport := d.Get("transport").(map[string]interface{})
	transportName := transport["name"].(string)
	transportType := transport["type"].(string)
	transportVlanID := 0
	if transport["vlanid"] != nil {
		transportVlanID, _ = strconv.Atoi(transport["vlanid"].(string))
	}

	localPreferenceTmp := d.Get("localpreference").(int)
	if localPreferenceTmp > 0 {
		localPreference = localPreferenceTmp
	}

	if d.Get("state").(string) != "" {
		state = d.Get("state").(string)
	}

	if transportType == "" {
		transportType = "port"
	}

	if transportType == "port" {
		port = transportName
		vlanID = 1
	} else {
		vlanID = 1
		if vnet, ok := findVNetByName(clientset, transportName); ok {
			vnetID = vnet.ID
		} else {
			return fmt.Errorf("invalid vnet '%s'", transportName)
		}
	}

	if transportVlanID > 1 && transportVlanID > 0 {
		vlanID = transportVlanID
	}

	localIPString := d.Get("localip").(string)

	localIP, cidr, err := net.ParseCIDR(localIPString)
	if err != nil {
		return err
	}
	remoteIP, _, err := net.ParseCIDR(d.Get("remoteip").(string))
	if err != nil {
		return err
	}
	prefixLength, _ := cidr.Mask.Size()
	if localIP.To4() != nil {
		ipVersion = "ipv4"
	}

	multihopMap := d.Get("multihop").(map[string]interface{})
	multihopNeighborAddress := multihopMap["neighboraddress"].(string)
	multihopUpdateSource := multihopMap["updatesource"].(string)
	multihopHop, _ := strconv.Atoi(multihopMap["hops"].(string))

	prefixListInboundArr := []string{}
	for _, pr := range d.Get("prefixlistinbound").([]interface{}) {
		prefixListInboundArr = append(prefixListInboundArr, pr.(string))
	}

	prefixListOutbound := []string{}
	for _, pr := range d.Get("prefixlistoutbound").([]interface{}) {
		prefixListOutbound = append(prefixListOutbound, pr.(string))
	}

	communityArr := []string{}
	for _, pr := range d.Get("sendbgpcommunity").([]interface{}) {
		communityArr = append(communityArr, pr.(string))
	}

	bgpAdd := &bgp.EBGPAdd{
		Name:               d.Get("name").(string),
		Site:               bgp.IDName{Name: siteName},
		Vlan:               vlanID,
		AllowAsIn:          d.Get("allowasin").(int),
		BgpPassword:        d.Get("bgppassword").(string),
		BgpCommunity:       strings.Join(communityArr, "\n"),
		Description:        d.Get("description").(string),
		IPFamily:           ipVersion,
		LocalIP:            localIP.String(),
		RemoteIP:           remoteIP.String(),
		LocalPreference:    localPreference,
		Multihop:           multihopHop,
		NeighborAddress:    &multihopNeighborAddress,
		UpdateSource:       multihopUpdateSource,
		NeighborAS:         d.Get("neighboras").(int),
		PrefixLength:       prefixLength,
		DefaultOriginate:   originate,
		PrefixInboundMax:   d.Get("prefixinboundmax").(string),
		PrefixListInbound:  strings.Join(prefixListInboundArr, "\n"),
		PrefixListOutbound: strings.Join(prefixListOutbound, "\n"),
		PrependInbound:     d.Get("prependinbound").(int),
		PrependOutbound:    d.Get("prependoutbound").(int),
		Hardware:           bgp.IDNone{ID: hwID},
		Vnet:               bgp.IDNone{ID: vnetID},
		Port:               bgp.IDName{Name: port},
		State:              state,
		Weight:             d.Get("weight").(int),
	}

	js, _ := json.Marshal(bgpAdd)
	log.Println("[DEBUG] bgpAdd", string(js))

	reply, err := clientset.BGP().Add(bgpAdd)
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

	_ = d.Set("bgpid", idStruct.ID)
	// d.SetId(vnetAdd.Name)
	d.SetId(d.Get("name").(string))
	return nil
}

func resourceRead(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)
	var bgp *bgp.EBGP
	bgps, err := clientset.BGP().Get()
	if err != nil {
		return err
	}

	for _, b := range bgps {
		if b.ID == d.Get("bgpid").(int) {
			bgp = b
			break
		}
	}

	if bgp == nil {
		return nil
	}

	d.SetId(bgp.Name)
	err = d.Set("name", bgp.Name)
	if err != nil {
		return err
	}
	err = d.Set("site", bgp.SiteName)
	if err != nil {
		return err
	}
	err = d.Set("softgate", bgp.TermSwName)
	if err != nil {
		return err
	}
	err = d.Set("neighboras", strconv.Itoa(bgp.NeighborAs))
	if err != nil {
		return err
	}

	transport := make(map[string]interface{})
	transportType := "port"
	transportName := bgp.PortName
	if port, ok := findPortByID(clientset, bgp.SiteID, bgp.SwitchPortID); ok {
		transportName = fmt.Sprintf("%s@%s", port.PortName, port.SwitchName)
	}

	if bgp.CircuitInternal == 0 {
		transportType = "vnet"
		transportName = bgp.CircuitName
	} else {
		tr := d.Get("transport").(map[string]interface{})
		if tr["vlanid"] != nil {
			transportVlanID, _ := strconv.Atoi(transport["vlanid"].(string))
			if !(transportVlanID >= 1 && bgp.Vlan == 1) {
				transport["vlanid"] = strconv.Itoa(bgp.Vlan)
			}
		} else if bgp.Vlan > 1 {
			transport["vlanid"] = strconv.Itoa(bgp.Vlan)
		}
	}

	transport["type"] = transportType
	transport["name"] = transportName

	err = d.Set("transport", transport)
	if err != nil {
		return err
	}
	err = d.Set("localip", fmt.Sprintf("%s/%d", bgp.LocalIP, bgp.PrefixLength))
	if err != nil {
		return err
	}
	err = d.Set("remoteip", fmt.Sprintf("%s/%d", bgp.RemoteIP, bgp.PrefixLength))
	if err != nil {
		return err
	}
	err = d.Set("description", bgp.Description)
	if err != nil {
		return err
	}
	err = d.Set("state", bgp.Status)
	if err != nil {
		return err
	}

	terminateOnSwitchMap := d.Get("terminateonswitch").(map[string]interface{})
	terminateOnSwitchEnabled := "false"
	terminateOnSwitchName := terminateOnSwitchMap["switchname"].(string)
	if bgp.TerminateOnSwitch == "yes" {
		terminateOnSwitchEnabled = "false"
		terminateOnSwitchName = bgp.TermSwName
	}
	terminateOnSwitchMap["enabled"] = terminateOnSwitchEnabled
	terminateOnSwitchMap["switchname"] = terminateOnSwitchName
	err = d.Set("terminateonswitch", terminateOnSwitchMap)
	if err != nil {
		return err
	}

	multihop := make(map[string]interface{})
	multihop["neighboraddress"] = bgp.NeighborAddress
	multihop["updatesource"] = bgp.UpdateSource
	multihop["hops"] = strconv.Itoa(bgp.Multihop)
	err = d.Set("multihop", multihop)
	if err != nil {
		return err
	}

	err = d.Set("bgppassword", bgp.BgpPassword)
	if err != nil {
		return err
	}
	err = d.Set("allowasin", bgp.AllowasIn)
	if err != nil {
		return err
	}
	var defaultOriginate bool
	if bgp.DefaultOriginate == "enabled" {
		defaultOriginate = true
	}
	err = d.Set("defaultoriginate", defaultOriginate)
	if err != nil {
		return err
	}
	err = d.Set("prefixinboundmax", strconv.Itoa(bgp.PrefixLimit))
	if err != nil {
		return err
	}
	err = d.Set("inboundroutemap", bgp.InboundRouteMap)
	if err != nil {
		return err
	}
	err = d.Set("outboundroutemap", bgp.OutboundRouteMap)
	if err != nil {
		return err
	}
	err = d.Set("localpreference", bgp.LocalPreference)
	if err != nil {
		return err
	}
	err = d.Set("weight", bgp.Weight)
	if err != nil {
		return err
	}
	err = d.Set("prependinbound", bgp.PrependInbound)
	if err != nil {
		return err
	}
	err = d.Set("prependoutbound", bgp.PrependOutbound)
	if err != nil {
		return err
	}

	err = d.Set("prefixlistinbound", strings.Split(bgp.PrefixListInbound, "\n"))
	if err != nil {
		return err
	}
	err = d.Set("prefixlistoutbound", strings.Split(bgp.PrefixListOutbound, "\n"))
	if err != nil {
		return err
	}
	err = d.Set("sendbgpcommunity", strings.Split(bgp.Community, ","))
	if err != nil {
		return err
	}

	return nil
}

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	var (
		vlanID    = 1
		state     = "enabled"
		ipVersion = "ipv6"
		hwID      = 0
		port      = ""
		vnetID    = 0
	)

	originate := "disabled"
	localPreference := 100

	siteName := d.Get("site").(string)

	if d.Get("defaultoriginate").(bool) {
		originate = "enabled"
	}

	hardware := d.Get("softgate").(string)

	inventory, err := clientset.Inventory().Get()
	if err != nil {
		return err
	}

	for _, hw := range inventory {
		if hw.Name == hardware && hardware != "auto" {
			hwID = hw.ID
		}
	}

	transport := d.Get("transport").(map[string]interface{})
	transportName := transport["name"].(string)
	transportType := transport["type"].(string)
	transportVlanID := 0
	if transport["vlanid"] != nil {
		transportVlanID, _ = strconv.Atoi(transport["vlanid"].(string))
	}

	if transportVlanID > 1 && transportVlanID > 0 {
		vlanID = transportVlanID
	}

	localPreferenceTmp := d.Get("localpreference").(int)
	if localPreferenceTmp > 0 {
		localPreference = localPreferenceTmp
	}

	if d.Get("state").(string) != "" {
		state = d.Get("state").(string)
	}

	if transportType == "" {
		transportType = "port"
	}

	if transportType == "port" {
		port = transportName
	} else {
		vlanID = 1
		if vnet, ok := findVNetByName(clientset, transportName); ok {
			vnetID = vnet.ID
		} else {
			return fmt.Errorf("invalid vnet '%s'", transportName)
		}
	}

	localIPString := d.Get("localip").(string)

	localIP, cidr, err := net.ParseCIDR(localIPString)
	if err != nil {
		return err
	}
	remoteIP, _, err := net.ParseCIDR(d.Get("remoteip").(string))
	if err != nil {
		return err
	}
	prefixLength, _ := cidr.Mask.Size()
	if localIP.To4() != nil {
		ipVersion = "ipv4"
	}

	multihopMap := d.Get("multihop").(map[string]interface{})
	multihopNeighborAddress := multihopMap["neighboraddress"].(string)
	multihopUpdateSource := multihopMap["updatesource"].(string)
	multihopHop, _ := strconv.Atoi(multihopMap["hops"].(string))

	prefixListInboundArr := []string{}
	for _, pr := range d.Get("prefixlistinbound").([]interface{}) {
		prefixListInboundArr = append(prefixListInboundArr, pr.(string))
	}

	prefixListOutbound := []string{}
	for _, pr := range d.Get("prefixlistoutbound").([]interface{}) {
		prefixListOutbound = append(prefixListOutbound, pr.(string))
	}

	communityArr := []string{}
	for _, pr := range d.Get("sendbgpcommunity").([]interface{}) {
		communityArr = append(communityArr, pr.(string))
	}

	bgpID := d.Get("bgpid").(int)

	bgpUpdate := &bgp.EBGPUpdate{
		Name:               d.Get("name").(string),
		Site:               bgp.IDName{Name: siteName},
		Vlan:               vlanID,
		AllowAsIn:          d.Get("allowasin").(int),
		BgpPassword:        d.Get("bgppassword").(string),
		BgpCommunity:       strings.Join(communityArr, "\n"),
		Description:        d.Get("description").(string),
		IPFamily:           ipVersion,
		LocalIP:            localIP.String(),
		RemoteIP:           remoteIP.String(),
		LocalPreference:    localPreference,
		Multihop:           multihopHop,
		NeighborAddress:    &multihopNeighborAddress,
		UpdateSource:       multihopUpdateSource,
		NeighborAS:         d.Get("neighboras").(int),
		PrefixLength:       prefixLength,
		DefaultOriginate:   originate,
		PrefixInboundMax:   d.Get("prefixinboundmax").(string),
		PrefixListInbound:  strings.Join(prefixListInboundArr, "\n"),
		PrefixListOutbound: strings.Join(prefixListOutbound, "\n"),
		PrependInbound:     d.Get("prependinbound").(int),
		PrependOutbound:    d.Get("prependoutbound").(int),
		Hardware:           bgp.IDNone{ID: hwID},
		Vnet:               bgp.IDNone{ID: vnetID},
		Port:               bgp.IDName{Name: port},
		State:              state,
		Weight:             d.Get("weight").(int),
	}

	js, _ := json.Marshal(bgpUpdate)
	log.Println("[DEBUG] bgpUpdate", string(js))

	reply, err := clientset.BGP().Update(bgpID, bgpUpdate)
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

	return nil
}

func resourceExists(d *schema.ResourceData, m interface{}) (bool, error) {
	// clientset := m.(*api.Clientset)

	// vnet, _ := clientset.VNet().GetByID(d.Get("vnetid").(int))

	// if vnet == nil {
	// 	return false, nil
	// }
	// if vnet.ID > 0 {
	// 	return true, nil
	// }

	return true, nil
}

func resourceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	clientset := m.(*api.Clientset)

	bgps, _ := clientset.BGP().Get()
	name := d.Id()
	for _, bgp := range bgps {
		if bgp.Name == name {
			err := d.Set("bgpid", bgp.ID)
			if err != nil {
				return []*schema.ResourceData{d}, err
			}
			return []*schema.ResourceData{d}, nil
		}
	}

	return []*schema.ResourceData{d}, nil
}

func resourceDelete(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	reply, err := clientset.BGP().Delete(d.Get("bgpid").(int))
	if err != nil {
		return err
	}

	if reply.StatusCode != 200 {
		return fmt.Errorf(string(reply.Data))
	}

	d.SetId("")
	return nil
}
