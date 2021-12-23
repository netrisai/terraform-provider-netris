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
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the resource, also acts as it's unique ID",
			},
			"siteid": {
				Required: true,
				Type:     schema.TypeInt,
			},
			"hardware": {
				Optional: true,
				Type:     schema.TypeString,
			},
			"neighboras": {
				Optional: true,
				Type:     schema.TypeInt,
			},
			"portid": {
				Computed: true,
				Optional: true,
				Type:     schema.TypeInt,
			},
			"vnetid": {
				Computed: true,
				Optional: true,
				Type:     schema.TypeInt,
			},
			"vlanid": {
				Computed: true,
				Optional: true,
				Type:     schema.TypeInt,
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
				Optional: true,
				Type:     schema.TypeString,
			},
			"state": {
				Optional:     true,
				Default:      "enabled",
				ValidateFunc: validateState,
				Type:         schema.TypeString,
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
				Optional: true,
				Type:     schema.TypeString,
			},
			"allowasin": {
				Optional: true,
				Type:     schema.TypeInt,
			},
			"defaultoriginate": {
				Computed: true,
				Optional: true,
				Type:     schema.TypeBool,
			},
			"prefixinboundmax": {
				Computed: true,
				Optional: true,
				Type:     schema.TypeString,
			},
			"inboundroutemap": {
				Computed: true,
				Optional: true,
				Type:     schema.TypeString,
			},
			"outboundroutemap": {
				Computed: true,
				Optional: true,
				Type:     schema.TypeString,
			},
			"localpreference": {
				Computed: true,
				Optional: true,
				Type:     schema.TypeInt,
			},
			"weight": {
				Computed: true,
				Optional: true,
				Type:     schema.TypeInt,
			},
			"prependinbound": {
				Computed: true,
				Optional: true,
				Type:     schema.TypeInt,
			},
			"prependoutbound": {
				Computed: true,
				Optional: true,
				Type:     schema.TypeInt,
			},
			"prefixlistinbound": {
				Computed:    true,
				Optional:    true,
				Type:        schema.TypeList,
				Description: "Switch Ports",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"prefixlistoutbound": {
				Computed:    true,
				Optional:    true,
				Type:        schema.TypeList,
				Description: "Switch Ports",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"sendbgpcommunity": {
				Computed:    true,
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
		vlanID    = 1
		state     = "enabled"
		ipVersion = "ipv6"
		hwID      = 0
		portID    = 0
		vnetID    = 0
	)

	originate := "disabled"
	localPreference := 100

	siteID := d.Get("siteid").(int)

	if d.Get("defaultoriginate").(bool) {
		originate = "enabled"
	}

	hardware := d.Get("hardware").(string)

	inventory, err := clientset.Inventory().Get()
	if err != nil {
		return err
	}

	for _, hw := range inventory {
		if hw.Name == hardware && hardware != "auto" {
			hwID = hw.ID
		}
	}

	transportVlanID := d.Get("vlanid").(int)

	localPreferenceTmp := d.Get("localpreference").(int)
	if localPreferenceTmp > 0 {
		localPreference = localPreferenceTmp
	}

	if d.Get("state").(string) != "" {
		state = d.Get("state").(string)
	}

	if v := d.Get("portid").(int); v > 0 {
		portID = v
	} else if v := d.Get("vnetid").(int); v > 0 {
		vnetID = v
	}

	if transportVlanID >= 1 {
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

	multihopNeighborAddress := ""
	multihopUpdateSource := ""
	multihopHop := 0

	multihopMap := d.Get("multihop").(map[string]interface{})
	if v, ok := multihopMap["neighboraddress"]; ok {
		multihopNeighborAddress = v.(string)
	}
	if v, ok := multihopMap["updatesource"]; ok {
		multihopUpdateSource = v.(string)
	}
	if v, ok := multihopMap["hops"]; ok {
		multihopHop, _ = strconv.Atoi(v.(string))
	}

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

	var vnetIDNone interface{} = vnetID
	if vnetID == 0 {
		vnetIDNone = "none"
	}

	var hwIDNone interface{} = hwID
	if hwID == 0 {
		hwIDNone = "auto"
	}

	bgpAdd := &bgp.EBGPAdd{
		Name:               d.Get("name").(string),
		Site:               bgp.IDName{ID: siteID},
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
		Hardware:           bgp.IDNone{ID: hwIDNone},
		Vnet:               bgp.IDNone{ID: vnetIDNone},
		Port:               bgp.IDName{ID: portID},
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

	d.SetId(strconv.Itoa(idStruct.ID))
	return nil
}

func resourceRead(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)
	var bgp *bgp.EBGP
	bgps, err := clientset.BGP().Get()
	if err != nil {
		return err
	}
	id, _ := strconv.Atoi(d.Id())
	for _, b := range bgps {
		if b.ID == id {
			bgp = b
			break
		}
	}

	if bgp == nil {
		return nil
	}

	d.SetId(strconv.Itoa(bgp.ID))
	err = d.Set("name", bgp.Name)
	if err != nil {
		return err
	}
	err = d.Set("siteid", bgp.SiteID)
	if err != nil {
		return err
	}
	err = d.Set("hardware", bgp.TermSwName)
	if err != nil {
		return err
	}
	err = d.Set("neighboras", bgp.NeighborAs)
	if err != nil {
		return err
	}

	if a, ok := bgp.Vnet.ID.(float64); ok {
		err := d.Set("vnetid", int(a))
		if err != nil {
			return err
		}
	} else {
		err := d.Set("portid", bgp.Port.ID)
		if err != nil {
			return err
		}
	}

	if d.Get("vlanid").(int) > 0 {
		err := d.Set("vlanid", bgp.Vlan)
		if err != nil {
			return err
		}
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

	multihop := make(map[string]interface{})
	if bgp.Multihop > 0 {
		multihop["neighboraddress"] = bgp.NeighborAddress
		multihop["updatesource"] = bgp.UpdateSource
		multihop["hops"] = strconv.Itoa(bgp.Multihop)
		err = d.Set("multihop", multihop)
		if err != nil {
			return err
		}
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
	if bgp.InboundRouteMap > 0 {
		err = d.Set("inboundroutemap", bgp.InboundRouteMapName)
		if err != nil {
			return err
		}
	}

	if bgp.OutboundRouteMap > 0 {
		err = d.Set("outboundroutemap", bgp.OutboundRouteMapName)
		if err != nil {
			return err
		}
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
		portID    = 0
		vnetID    = 0
	)

	originate := "disabled"
	localPreference := 100

	siteID := d.Get("siteid").(int)

	if d.Get("defaultoriginate").(bool) {
		originate = "enabled"
	}

	hardware := d.Get("hardware").(string)

	inventory, err := clientset.Inventory().Get()
	if err != nil {
		return err
	}

	for _, hw := range inventory {
		if hw.Name == hardware && hardware != "auto" {
			hwID = hw.ID
		}
	}

	transportVlanID := d.Get("vlanid").(int)

	localPreferenceTmp := d.Get("localpreference").(int)
	if localPreferenceTmp > 0 {
		localPreference = localPreferenceTmp
	}

	if d.Get("state").(string) != "" {
		state = d.Get("state").(string)
	}

	if v := d.Get("portid").(int); v > 0 {
		portID = v
	} else if v := d.Get("vnetid").(int); v > 0 {
		vnetID = v
	}

	if transportVlanID >= 1 {
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

	multihopNeighborAddress := ""
	multihopUpdateSource := ""
	multihopHop := 0

	multihopMap := d.Get("multihop").(map[string]interface{})
	if v, ok := multihopMap["neighboraddress"]; ok {
		multihopNeighborAddress = v.(string)
	}
	if v, ok := multihopMap["updatesource"]; ok {
		multihopUpdateSource = v.(string)
	}
	if v, ok := multihopMap["hops"]; ok {
		multihopHop, _ = strconv.Atoi(v.(string))
	}

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

	var vnetIDNone interface{} = vnetID
	if vnetID == 0 {
		vnetIDNone = "none"
	}

	var hwIDNone interface{} = hwID
	if hwID == 0 {
		hwIDNone = "auto"
	}

	bgpID, _ := strconv.Atoi(d.Id())

	bgpUpdate := &bgp.EBGPUpdate{
		Name:               d.Get("name").(string),
		Site:               bgp.IDName{ID: siteID},
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
		Hardware:           bgp.IDNone{ID: hwIDNone},
		Vnet:               bgp.IDNone{ID: vnetIDNone},
		Port:               bgp.IDName{ID: portID},
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
	clientset := m.(*api.Clientset)
	bgpID, _ := strconv.Atoi(d.Id())

	bgps, err := clientset.BGP().Get()
	if err != nil {
		return false, err
	}

	for _, bgp := range bgps {
		if bgpID == bgp.ID {
			return true, nil
		}
	}

	return false, nil
}

func resourceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	clientset := m.(*api.Clientset)

	bgps, _ := clientset.BGP().Get()
	name := d.Id()
	for _, bgp := range bgps {
		if bgp.Name == name {
			d.SetId(strconv.Itoa(bgp.ID))
			return []*schema.ResourceData{d}, nil
		}
	}

	return []*schema.ResourceData{d}, nil
}

func resourceDelete(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)
	id, _ := strconv.Atoi(d.Id())
	reply, err := clientset.BGP().Delete(id)
	if err != nil {
		return err
	}

	if reply.StatusCode != 200 {
		return fmt.Errorf(string(reply.Data))
	}

	d.SetId("")
	return nil
}
