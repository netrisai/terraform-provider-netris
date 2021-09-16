package vnet

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/netrisai/netriswebapi/http"
	"github.com/netrisai/netriswebapi/v2/types/vnet"

	api "github.com/netrisai/netriswebapi/v2"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"vnetid": {
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
			"owner": {
				Required:    true,
				Type:        schema.TypeString,
				ForceNew:    true,
				Description: "A description of an item",
			},
			"state": {
				Required:    true,
				Type:        schema.TypeString,
				Description: "A description of an item",
			},
			"ports": {
				Optional:    true,
				Type:        schema.TypeList,
				Description: "Switch Ports",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"vlanid": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"gateways": {
				Optional:    true,
				Type:        schema.TypeList,
				Description: "Gateways",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"prefix": {
							Type:     schema.TypeString,
							Required: true,
						},
						"vlanid": {
							Type:     schema.TypeString,
							Optional: true,
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

	gateways := d.Get("gateways").([]interface{})
	gatewayList := []vnet.VNetAddGateway{}
	for _, gw := range gateways {
		gateway := gw.(map[string]interface{})
		gatewayList = append(gatewayList, vnet.VNetAddGateway{
			Prefix: gateway["prefix"].(string),
			Vlan:   gateway["vlanid"].(string),
		})
	}

	members := []vnet.VNetAddPort{}
	ports := d.Get("ports").([]interface{})
	for _, p := range ports {
		port := p.(map[string]interface{})
		members = append(members, vnet.VNetAddPort{
			Name:  port["name"].(string),
			Vlan:  port["vlanid"].(string),
			Lacp:  "off",
			State: "active",
		})
	}

	vnetAdd := &vnet.VNetAdd{
		Name:         d.Get("name").(string),
		Tenant:       vnet.VNetAddTenant{Name: d.Get("owner").(string)},
		GuestTenants: []vnet.VNetAddTenant{},
		Sites:        []vnet.VNetAddSite{{Name: "Yerevan"}},
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

	portList := make([]interface{}, 0)
	for _, port := range vnet.Ports {
		m := make(map[string]interface{})
		m["name"] = fmt.Sprintf("%s@%s", port.Port, port.SwitchName)
		m["vlanid"] = port.Vlan
		portList = append(portList, m)
	}
	err = d.Set("ports", portList)
	if err != nil {
		return err
	}

	gatewayList := make([]interface{}, 0)
	for _, gateway := range vnet.Gateways {
		m := make(map[string]interface{})
		m["prefix"] = gateway.Prefix
		m["vlanid"] = gateway.Vlan
		gatewayList = append(gatewayList, m)
	}
	err = d.Set("gateways", gatewayList)
	if err != nil {
		return err
	}
	return nil
}

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	gateways := d.Get("gateways").([]interface{})
	gatewayList := []vnet.VNetUpdateGateway{}
	for _, gw := range gateways {
		gateway := gw.(map[string]interface{})
		gatewayList = append(gatewayList, vnet.VNetUpdateGateway{
			Prefix: gateway["prefix"].(string),
			Vlan:   gateway["vlanid"].(string),
		})
	}

	members := []vnet.VNetUpdatePort{}
	ports := d.Get("ports").([]interface{})
	for _, p := range ports {
		port := p.(map[string]interface{})
		members = append(members, vnet.VNetUpdatePort{
			Vlan:  port["vlanid"].(string),
			Name:  port["name"].(string),
			Lacp:  "off",
			State: "active",
		})
	}

	vnetUpdate := &vnet.VNetUpdate{
		Name:         d.Get("name").(string),
		GuestTenants: []vnet.VNetUpdateGuestTenant{},
		Sites:        []vnet.VNetUpdateSite{{Name: "Yerevan"}},
		State:        d.Get("state").(string),
		Gateways:     gatewayList,
		Ports:        members,
	}

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
