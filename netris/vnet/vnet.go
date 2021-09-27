package vnet

import (
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/netrisai/netriswebapi/http"
	"github.com/netrisai/netriswebapi/v2/types/ipam"
	"github.com/netrisai/netriswebapi/v2/types/vnet"
	"github.com/netrisai/terraform-provider-netris/netris/subnet"

	api "github.com/netrisai/netriswebapi/v2"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"vnetid": {
				Type:             schema.TypeInt,
				Optional:         true,
				DiffSuppressFunc: DiffSuppress,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the resource, also acts as it's unique ID",
			},
			"owner": {
				Required: true,
				Type:     schema.TypeString,
				ForceNew: true,
			},
			"state": {
				Optional:     true,
				Default:      "active",
				ValidateFunc: validateState,
				Type:         schema.TypeString,
			},
			"sites": {
				Required: true,
				Type:     schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"ports": {
							Optional: true,
							Type:     schema.TypeList,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"vlanid": {
										Default:  "1",
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"gateways": {
							Optional: true,
							Type:     schema.TypeList,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"prefix": {
										ValidateFunc: validateGateway,
										Type:         schema.TypeString,
										Required:     true,
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
	var sitesList []map[string]interface{}
	for _, site := range sites {
		sitesList = append(sitesList, site.(map[string]interface{}))
	}

	siteNames := []vnet.VNetAddSite{}
	members := []vnet.VNetAddPort{}
	gatewayList := []vnet.VNetAddGateway{}

	for _, site := range sitesList {
		if siteName, ok := site["name"]; ok {
			siteNames = append(siteNames, vnet.VNetAddSite{Name: siteName.(string)})
		}
		if gws, ok := site["gateways"]; ok {
			gateways := gws.([]interface{})

			for _, gw := range gateways {
				gateway := gw.(map[string]interface{})
				gatewayList = append(gatewayList, vnet.VNetAddGateway{
					Prefix: gateway["prefix"].(string),
					Vlan:   gateway["vlanid"].(string),
				})
			}
		}
		if p, ok := site["ports"]; ok {

			ports := p.([]interface{})
			for _, p := range ports {
				port := p.(map[string]interface{})
				members = append(members, vnet.VNetAddPort{
					Name:  port["name"].(string),
					Vlan:  port["vlanid"].(string),
					Lacp:  "off",
					State: "active",
				})
			}
		}
	}

	vnetAdd := &vnet.VNetAdd{
		Name:         d.Get("name").(string),
		Tenant:       vnet.VNetAddTenant{Name: d.Get("owner").(string)},
		GuestTenants: []vnet.VNetAddTenant{},
		Sites:        siteNames,
		State:        d.Get("state").(string),
		Gateways:     gatewayList,
		Ports:        members,
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

	_ = d.Set("vnetid", idStruct.ID)
	d.SetId(vnetAdd.Name)
	return nil
}

func resourceRead(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	vnet, err := clientset.VNet().GetByID(d.Get("vnetid").(int))
	if err != nil {
		return err
	}

	d.SetId(vnet.Name)
	err = d.Set("name", vnet.Name)
	if err != nil {
		return err
	}
	err = d.Set("owner", vnet.Tenant.Name)
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
		s["name"] = site.Name
		s["ports"] = portList
		s["gateways"] = gatewayList
		sites = append(sites, s)
	}

	err = d.Set("sites", sites)
	if err != nil {
		return err
	}
	return nil
}

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	sites := d.Get("sites").([]interface{})
	var sitesList []map[string]interface{}
	for _, site := range sites {
		sitesList = append(sitesList, site.(map[string]interface{}))
	}

	siteNames := []vnet.VNetUpdateSite{}
	members := []vnet.VNetUpdatePort{}
	gatewayList := []vnet.VNetUpdateGateway{}

	for _, site := range sitesList {
		if siteName, ok := site["name"]; ok {
			siteNames = append(siteNames, vnet.VNetUpdateSite{Name: siteName.(string)})
		}
		if gws, ok := site["gateways"]; ok {
			gateways := gws.([]interface{})

			for _, gw := range gateways {
				gateway := gw.(map[string]interface{})
				gatewayList = append(gatewayList, vnet.VNetUpdateGateway{
					Prefix: gateway["prefix"].(string),
					Vlan:   gateway["vlanid"].(string),
				})
			}
		}
		if p, ok := site["ports"]; ok {

			ports := p.([]interface{})
			for _, p := range ports {
				port := p.(map[string]interface{})
				members = append(members, vnet.VNetUpdatePort{
					Name:  port["name"].(string),
					Vlan:  port["vlanid"].(string),
					Lacp:  "off",
					State: "active",
				})
			}
		}
	}

	vnetUpdate := &vnet.VNetUpdate{
		Name:         d.Get("name").(string),
		GuestTenants: []vnet.VNetUpdateGuestTenant{},
		Sites:        siteNames,
		State:        d.Get("state").(string),
		Gateways:     gatewayList,
		Ports:        members,
	}

	js, _ := json.Marshal(vnetUpdate)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.VNet().Update(d.Get("vnetid").(int), vnetUpdate)
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

	reply, err := clientset.VNet().Delete(d.Get("vnetid").(int))
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

	vnet, _ := clientset.VNet().GetByID(d.Get("vnetid").(int))

	if vnet == nil {
		return false, nil
	}
	if vnet.ID > 0 {
		return true, nil
	}

	return true, nil
}

func resourceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	clientset := m.(*api.Clientset)

	vnets, _ := clientset.VNet().Get()
	name := d.Id()
	for _, vnet := range vnets {
		if vnet.Name == name {
			err := d.Set("vnetid", vnet.ID)
			if err != nil {
				return []*schema.ResourceData{d}, err
			}
			return []*schema.ResourceData{d}, nil
		}
	}

	return []*schema.ResourceData{d}, nil
}
