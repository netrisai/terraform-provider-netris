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

package sw

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/netrisai/netriswebapi/http"
	api "github.com/netrisai/netriswebapi/v2"
	"github.com/netrisai/netriswebapi/v2/types/inventory"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: "Creates and manages Switches",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "User assigned name of switch.",
			},
			"tenantid": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "ID of tenant. Users of this tenant will be permitted to edit this unit.",
			},
			"siteid": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The site ID where this switch belongs.",
			},
			"description": {
				Computed:    true,
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Switch description.",
			},
			"nos": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Switch OS. Possible values: `cumulus_linux`, `sonic`, `ubuntu_switch_dev`",
			},
			"asnumber": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Switch AS numbers. Valid value is ASN (example `420000002`) or `auto`. If set `auto` the controller will assign automatically from `System ASN range`",
			},
			"profileid": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "An inventory profile ID to define global configuration (NTP, DNS, timezone, etc...)",
			},
			"mainip": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A unique IP address which will be used as a loopback address of this unit. Valid value is ip address (example `198.51.100.21`) or `auto`. If set `auto` the controller will assign an ip address automatically from subnets with relevant purpose.",
			},
			"mgmtip": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A unique IP address to be used on out of band management interface. Valid value is ip address (example `192.0.2.21`) or `auto`. If set `auto` the controller will assign an ip address automatically from subnets with relevant purpose.",
			},
			"macaddress": {
				Computed: true,
				Type:     schema.TypeString,
				Optional: true,
			},
			"portcount": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Preliminary port count is used for definition of topology. Possible values: `16`, `32`, `48`, `54`, `56`, `64`",
			},
			"breakout": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Initial Break Out applies to all switch ports of this switch.",
				ForceNew:    true,
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

	nosList, err := clientset.Inventory().GetNOS()
	if err != nil {
		return err
	}

	nosMap := make(map[string]inventory.NOS)
	for _, nos := range nosList {
		nosMap[nos.Tag] = *nos
	}

	nos := nosMap[d.Get("nos").(string)]

	profileID := d.Get("profileid").(int)

	var asnAny interface{} = d.Get("asnumber").(string)
	asn := asnAny.(string)
	if !(asn == "auto" || asn == "") {
		asnInt, err := strconv.Atoi(asn)
		if err != nil {
			return err
		}
		asnAny = asnInt
	}

	swAdd := &inventory.HWSwitchAdd{
		Name:        d.Get("name").(string),
		Tenant:      inventory.IDName{ID: d.Get("tenantid").(int)},
		Site:        inventory.IDName{ID: d.Get("siteid").(int)},
		Description: d.Get("description").(string),
		Nos:         nos,
		Asn:         asnAny,
		Profile:     inventory.IDName{ID: profileID},
		MainAddress: d.Get("mainip").(string),
		MgmtAddress: d.Get("mgmtip").(string),
		MacAddress:  d.Get("macaddress").(string),
		PortCount:   d.Get("portcount").(int),
		Links:       []inventory.HWLink{},
		Breakout:    d.Get("breakout").(string),
	}

	js, _ := json.Marshal(swAdd)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.Inventory().AddSwitch(swAdd)
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
	sw, err := clientset.Inventory().GetByID(id)
	if err != nil {
		return nil
	}

	d.SetId(strconv.Itoa(sw.ID))
	err = d.Set("name", sw.Name)
	if err != nil {
		return err
	}
	err = d.Set("tenantid", sw.Tenant.ID)
	if err != nil {
		return err
	}
	err = d.Set("siteid", sw.Site.ID)
	if err != nil {
		return err
	}
	err = d.Set("description", sw.Description)
	if err != nil {
		return err
	}
	err = d.Set("nos", sw.Nos.Tag)
	if err != nil {
		return err
	}
	if asnumber := d.Get("asnumber"); asnumber.(string) != "auto" {
		err = d.Set("asnumber", strconv.Itoa(sw.Asn))
		if err != nil {
			return err
		}
	}

	if sw.Profile.Name != "None" {
		err = d.Set("profileid", sw.Profile.ID)
		if err != nil {
			return err
		}
	}

	if main := d.Get("mainip"); main.(string) != "auto" {
		err = d.Set("mainip", sw.MainAddress)
		if err != nil {
			return err
		}
	}
	if main := d.Get("mgmtip"); main.(string) != "auto" {
		err = d.Set("mgmtip", sw.MgmtAddress)
		if err != nil {
			return err
		}
	}
	err = d.Set("macaddress", sw.MacAddress)
	if err != nil {
		return err
	}
	err = d.Set("portcount", sw.PortCount)
	if err != nil {
		return err
	}

	return nil
}

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	nosList, err := clientset.Inventory().GetNOS()
	if err != nil {
		return err
	}

	nosMap := make(map[string]inventory.NOS)
	for _, nos := range nosList {
		nosMap[nos.Tag] = *nos
	}

	nos := nosMap[d.Get("nos").(string)]

	profileID := d.Get("profileid").(int)

	id, _ := strconv.Atoi(d.Id())
	sw, err := clientset.Inventory().GetByID(id)
	if err != nil {
		return nil
	}

	var asnAny interface{} = d.Get("asnumber").(string)
	asn := asnAny.(string)
	if !(asn == "auto" || asn == "") {
		asnInt, err := strconv.Atoi(asn)
		if err != nil {
			return err
		}
		asnAny = asnInt
	}

	swUpdate := &inventory.HWSwitchUpdate{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Tenant:      inventory.IDName{ID: d.Get("tenantid").(int)},
		Site:        inventory.IDName{ID: d.Get("siteid").(int)},
		Nos:         nos,
		Asn:         asnAny,
		Profile:     inventory.IDName{ID: profileID},
		MainAddress: d.Get("mainip").(string),
		MgmtAddress: d.Get("mgmtip").(string),
		MacAddress:  d.Get("macaddress").(string),
		PortCount:   d.Get("portcount").(int),
		Type:        "switch",
		Links:       sw.Links,
	}

	js, _ := json.Marshal(swUpdate)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.Inventory().UpdateSwitch(id, swUpdate)
	if err != nil {
		log.Println("[DEBUG]", err)
		return err
	}

	js, _ = json.Marshal(reply)
	log.Println("[DEBUG]", string(js))

	log.Println("[DEBUG]", string(reply.Data))

	if reply.StatusCode != 200 {
		return fmt.Errorf(string(reply.Data))
	}

	return nil
}

func resourceDelete(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	id, _ := strconv.Atoi(d.Id())
	reply, err := clientset.Inventory().Delete("switch", id)
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
	sw, err := clientset.Inventory().GetByID(id)
	if err != nil {
		return false, nil
	}

	if sw.ID == 0 {
		return false, nil
	}

	return true, nil
}

func resourceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	clientset := m.(*api.Clientset)

	sws, err := clientset.Inventory().Get()
	if err != nil {
		return []*schema.ResourceData{d}, err
	}
	name := d.Id()
	for _, sw := range sws {
		if sw.Name == name {
			d.SetId(strconv.Itoa(sw.ID))
			return []*schema.ResourceData{d}, nil
		}
	}

	return []*schema.ResourceData{d}, nil
}
