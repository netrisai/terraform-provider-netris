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
			"softgate": {
				Optional:    true,
				Type:        schema.TypeString,
				Description: "A description of an item",
			},
			"neighboras": {
				Optional:    true,
				Type:        schema.TypeString,
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
				Optional:    true,
				Type:        schema.TypeString,
				Description: "A description of an item",
			},
			"remoteip": {
				Optional:    true,
				Type:        schema.TypeString,
				Description: "A description of an item",
			},
			"description": {
				Optional:    true,
				Type:        schema.TypeString,
				Description: "A description of an item",
			},
			"state": {
				Optional:    true,
				Type:        schema.TypeString,
				Description: "A description of an item",
			},
			"terminateonswitch": {
				Optional:    true,
				Type:        schema.TypeMap,
				Description: "Switch Ports",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"multihop": {
				Optional:    true,
				Type:        schema.TypeMap,
				Description: "Multihop",
				Elem: &schema.Schema{
					Type: schema.TypeString,
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

	sites, err := clientset.Site().Get()
	if err != nil {
		return err
	}
	var (
		vlanID            = 1
		siteID            int
		nfvID             int
		nfvPortID         int
		state             = "enabled"
		terminateOnSwitch = "no"
		termSwitchID      int
		portID            int
		vnetID            int
		ipVersion         = "ipv6"
	)

	originate := "disabled"
	localPreference := 100

	siteName := d.Get("site").(string)
	for _, site := range sites {
		if siteName == site.Name {
			siteID = site.ID
		}
	}
	if siteID == 0 {
		return fmt.Errorf("site '%s' not found", siteName)
	}

	if d.Get("defaultoriginate").(bool) {
		originate = "enabled"
	}

	softgate := d.Get("softgate").(string)

	transport := d.Get("transport").(map[string]interface{})
	transportName := transport["name"].(string)
	transportType := transport["type"].(string)
	transportVlanID, _ := strconv.Atoi(transport["vlanid"].(string))

	if transportVlanID > 1 {
		vlanID = transportVlanID
	}

	localPreferenceTmp := d.Get("localpreference").(int)
	if localPreferenceTmp > 0 {
		localPreference = localPreferenceTmp
	}

	if d.Get("state").(string) != "" {
		state = d.Get("state").(string)
	}

	terminateOnSwitchMap := d.Get("terminateonswitch").(map[string]interface{})
	terminateOnSwitchEnabled := terminateOnSwitchMap["enabled"].(string)
	terminateOnSwitchName := terminateOnSwitchMap["switchname"].(string)

	if terminateOnSwitchEnabled == "true" {
		terminateOnSwitch = "yes"
	} else {
		bpgOffloaders, err := clientset.BGP().GetOffloaders(siteID)
		if err != nil {
			return err
		}
		found := false
		for _, offloader := range bpgOffloaders {
			if softgate == offloader.Name {
				nfvID = offloader.ID
				termSwitchID = nfvID
				if len(offloader.Links) > 0 {
					nfvPortID = offloader.Links[0].Local.ID
				}
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("invalid softgate '%s'", softgate)
		}
	}

	if transportType == "" {
		transportType = "port"
	}

	if transportType == "port" {
		if port, ok := findPort(clientset, siteID, transportName); ok {
			portID = port.PortID
			if terminateOnSwitchEnabled == "true" {
				termSwitchID = port.SwitchID
			}
		} else {
			return fmt.Errorf("invalid port '%s'", transportName)
		}
	} else {
		vlanID = 1
		if vnet, ok := findVNetByName(clientset, transportName); ok {
			vnetID = vnet.ID
			if terminateOnSwitchEnabled == "true" {
				if sw, ok := findSwitchByName(clientset, siteID, transportName); ok {
					termSwitchID = sw.SwitchID
				} else {
					return fmt.Errorf("invalid TerminateOnSwitchName '%s'", transportName)
				}
			}
		} else {
			return fmt.Errorf("invalid vnet '%s'", transportName)
		}
	}

	localIP, cidr, _ := net.ParseCIDR(d.Get("localip").(string))
	remoteIP, _, _ := net.ParseCIDR(d.Get("remoteip").(string))
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
		SiteID:             siteID,
		Vlan:               vlanID,
		AllowasIn:          d.Get("allowasin").(int),
		BgpPassword:        d.Get("bgppassword").(string),
		Community:          strings.Join(communityArr, "\n"),
		Description:        d.Get("description").(string),
		IPVersion:          ipVersion,
		LocalIP:            localIP.String(),
		RemoteIP:           remoteIP.String(),
		LocalPreference:    localPreference,
		Multihop:           multihopHop,
		NeighborAddress:    &multihopNeighborAddress,
		UpdateSource:       multihopUpdateSource,
		NeighborAs:         d.Get("neighboras").(string),
		PrefixLength:       prefixLength,
		NfvID:              nfvID,
		NfvPortID:          nfvPortID,
		Originate:          originate,
		PrefixLimit:        d.Get("prefixinboundmax").(string),
		PrefixListInbound:  strings.Join(prefixListInboundArr, "\n"),
		PrefixListOutbound: strings.Join(prefixListOutbound, "\n"),
		PrependInbound:     d.Get("prependinbound").(int),
		PrependOutbound:    d.Get("prependoutbound").(int),
		RcircuitID:         vnetID,
		Status:             state,
		// SwitchID: ,
		// SwitchName: ,
		SwitchPortID:      portID,
		TermSwitchID:      termSwitchID,
		TermSwitchName:    terminateOnSwitchName,
		TerminateOnSwitch: terminateOnSwitch,
		Weight:            d.Get("weight").(int),
	}

	js, _ := json.Marshal(bgpAdd)
	log.Println("[DEBUG] vnetAdd", string(js))

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
		return fmt.Errorf("'%s' bgp not found", d.Get("name").(string))
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
	err = d.Set("softgate", bgp.OffloaderSwName)
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

	if bgp.CircuitInternal == 0 {
		transportType = "vnet"
		transportName = bgp.CircuitName
	}

	transport["type"] = transportType
	transport["name"] = transportName
	transport["vlanid"] = strconv.Itoa(bgp.Vlan)
	err = d.Set("transport", transport)
	if err != nil {
		return err
	}

	// portList := make([]interface{}, 0)
	// for _, port := range vnet.Ports {
	// 	m := make(map[string]interface{})
	// 	m["name"] = fmt.Sprintf("%s@%s", port.Port, port.SwitchName)
	// 	m["vlanid"] = port.Vlan
	// 	portList = append(portList, m)
	// }
	// err = d.Set("ports", portList)
	// if err != nil {
	// 	return err
	// }

	// gatewayList := make([]interface{}, 0)
	// for _, gateway := range vnet.Gateways {
	// 	m := make(map[string]interface{})
	// 	m["prefix"] = gateway.Prefix
	// 	m["vlanid"] = gateway.Vlan
	// 	gatewayList = append(gatewayList, m)
	// }
	// err = d.Set("gateways", gatewayList)
	// if err != nil {
	// 	return err
	// }
	return nil
}

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	// clientset := m.(*api.Clientset)

	// sites, err := clientset.Site().Get()
	// if err != nil {
	// 	return err
	// }
	// var (
	// 	vlanID            = 1
	// 	siteID            int
	// 	nfvID             int
	// 	nfvPortID         int
	// 	state             = "enabled"
	// 	terminateOnSwitch = "no"
	// 	termSwitchID      int
	// 	portID            int
	// 	vnetID            int
	// 	ipVersion         = "ipv6"
	// )

	// originate := "disabled"
	// localPreference := 100

	// siteName := d.Get("site").(string)
	// for _, site := range sites {
	// 	if siteName == site.Name {
	// 		siteID = site.ID
	// 	}
	// }
	// if siteID == 0 {
	// 	return fmt.Errorf("site '%s' not found", siteName)
	// }

	// if d.Get("defaultoriginate").(bool) {
	// 	originate = "enabled"
	// }

	// softgate := d.Get("softgate").(string)

	// transport := d.Get("transport").(map[string]interface{})
	// transportName := transport["name"].(string)
	// transportType := transport["type"].(string)
	// transportVlanID, _ := strconv.Atoi(transport["vlanid"].(string))

	// if transportVlanID > 1 {
	// 	vlanID = transportVlanID
	// }

	// localPreferenceTmp := d.Get("localpreference").(int)
	// if localPreferenceTmp > 0 {
	// 	localPreference = localPreferenceTmp
	// }

	// if d.Get("state").(string) != "" {
	// 	state = d.Get("state").(string)
	// }

	// terminateOnSwitchMap := d.Get("terminateonswitch").(map[string]interface{})
	// terminateOnSwitchEnabled := terminateOnSwitchMap["enabled"].(string)
	// terminateOnSwitchName := terminateOnSwitchMap["switchname"].(string)

	// if terminateOnSwitchEnabled == "true" {
	// 	terminateOnSwitch = "yes"
	// } else {
	// 	bpgOffloaders, err := clientset.BGP().GetOffloaders(siteID)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	found := false
	// 	for _, offloader := range bpgOffloaders {
	// 		if softgate == offloader.Name {
	// 			nfvID = offloader.ID
	// 			termSwitchID = nfvID
	// 			if len(offloader.Links) > 0 {
	// 				nfvPortID = offloader.Links[0].Local.ID
	// 			}
	// 			found = true
	// 			break
	// 		}
	// 	}
	// 	if !found {
	// 		return fmt.Errorf("invalid softgate '%s'", softgate)
	// 	}
	// }

	// if transportType == "" {
	// 	transportType = "port"
	// }

	// if transportType == "port" {
	// 	if port, ok := findPort(clientset, siteID, transportName); ok {
	// 		portID = port.PortID
	// 		if terminateOnSwitchEnabled == "true" {
	// 			termSwitchID = port.SwitchID
	// 		}
	// 	} else {
	// 		return fmt.Errorf("invalid port '%s'", transportName)
	// 	}
	// } else {
	// 	vlanID = 1
	// 	if vnet, ok := findVNetByName(clientset, transportName); ok {
	// 		vnetID = vnet.ID
	// 		if terminateOnSwitchEnabled == "true" {
	// 			if sw, ok := findSwitchByName(clientset, siteID, transportName); ok {
	// 				termSwitchID = sw.SwitchID
	// 			} else {
	// 				return fmt.Errorf("invalid TerminateOnSwitchName '%s'", transportName)
	// 			}
	// 		}
	// 	} else {
	// 		return fmt.Errorf("invalid vnet '%s'", transportName)
	// 	}
	// }

	// localIP, cidr, _ := net.ParseCIDR(d.Get("localip").(string))
	// remoteIP, _, _ := net.ParseCIDR(d.Get("remoteip").(string))
	// prefixLength, _ := cidr.Mask.Size()
	// if localIP.To4() != nil {
	// 	ipVersion = "ipv4"
	// }

	// multihopMap := d.Get("multihop").(map[string]interface{})
	// multihopNeighborAddress := multihopMap["neighboraddress"].(string)
	// multihopUpdateSource := multihopMap["updatesource"].(string)
	// multihopHop, _ := strconv.Atoi(multihopMap["hops"].(string))

	// prefixListInboundArr := []string{}
	// for _, pr := range d.Get("prefixlistinbound").([]interface{}) {
	// 	prefixListInboundArr = append(prefixListInboundArr, pr.(string))
	// }

	// prefixListOutbound := []string{}
	// for _, pr := range d.Get("prefixlistoutbound").([]interface{}) {
	// 	prefixListOutbound = append(prefixListOutbound, pr.(string))
	// }

	// communityArr := []string{}
	// for _, pr := range d.Get("sendbgpcommunity").([]interface{}) {
	// 	communityArr = append(communityArr, pr.(string))
	// }

	// bgpUpdate := &bgp.EBGPUpdate{
	// 	ID:                 d.Get("bgpid").(int),
	// 	Name:               d.Get("name").(string),
	// 	SiteID:             siteID,
	// 	Vlan:               vlanID,
	// 	AllowasIn:          d.Get("allowasin").(int),
	// 	BgpPassword:        d.Get("bgppassword").(string),
	// 	Community:          strings.Join(communityArr, "\n"),
	// 	Description:        d.Get("description").(string),
	// 	IPVersion:          ipVersion,
	// 	LocalIP:            localIP.String(),
	// 	RemoteIP:           remoteIP.String(),
	// 	LocalPreference:    localPreference,
	// 	Multihop:           multihopHop,
	// 	NeighborAddress:    &multihopNeighborAddress,
	// 	UpdateSource:       multihopUpdateSource,
	// 	NeighborAs:         d.Get("neighboras").(string),
	// 	PrefixLength:       prefixLength,
	// 	NfvID:              nfvID,
	// 	NfvPortID:          nfvPortID,
	// 	Originate:          originate,
	// 	PrefixLimit:        d.Get("prefixinboundmax").(string),
	// 	PrefixListInbound:  strings.Join(prefixListInboundArr, "\n"),
	// 	PrefixListOutbound: strings.Join(prefixListOutbound, "\n"),
	// 	PrependInbound:     d.Get("prependinbound").(int),
	// 	PrependOutbound:    d.Get("prependoutbound").(int),
	// 	RcircuitID:         vnetID,
	// 	Status:             state,
	// 	// SwitchID: ,
	// 	// SwitchName: ,
	// 	SwitchPortID:      portID,
	// 	TermSwitchID:      termSwitchID,
	// 	TermSwitchName:    terminateOnSwitchName,
	// 	TerminateOnSwitch: terminateOnSwitch,
	// 	Weight:            d.Get("weight").(int),
	// }

	// js, _ := json.Marshal(bgpUpdate)
	// log.Println("[DEBUG] bgpUpdate", string(js))

	// reply, err := clientset.BGP().Update(bgpUpdate)
	// if err != nil {
	// 	log.Println("[DEBUG]", err)
	// 	return err
	// }

	// js, _ = json.Marshal(reply)
	// log.Println("[DEBUG]", string(js))

	// log.Println("[DEBUG]", string(reply.Data))

	// idStruct := struct {
	// 	ID int `json:"id"`
	// }{}

	// data, err := reply.Parse()
	// if err != nil {
	// 	log.Println("[DEBUG]", err)
	// 	return err
	// }

	// err = http.Decode(data.Data, &idStruct)
	// if err != nil {
	// 	log.Println("[DEBUG]", err)
	// 	return err
	// }

	// log.Println("[DEBUG] ID:", idStruct.ID)

	// if reply.StatusCode != 200 {
	// 	return fmt.Errorf(string(reply.Data))
	// }

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
