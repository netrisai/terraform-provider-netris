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

package softgate

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/netrisai/netriswebapi/http"
	api "github.com/netrisai/netriswebapi/v2"
	"github.com/netrisai/netriswebapi/v2/types/inventory"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"itemid": {
				Type:             schema.TypeInt,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: DiffSuppress,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the resource, also acts as it's unique ID",
			},
			"tenant": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the resource, also acts as it's unique ID",
			},
			"siteid": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The name of the resource, also acts as it's unique ID",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of the resource, also acts as it's unique ID",
			},
			"profile": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of the resource, also acts as it's unique ID",
			},
			"mainip": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the resource, also acts as it's unique ID",
			},
			"mgmtip": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the resource, also acts as it's unique ID",
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

	profileID := 0
	profiles, err := clientset.Inventory().GetProfiles()
	if err != nil {
		return err
	}
	for _, profile := range profiles {
		if profile.Name == d.Get("profile").(string) {
			profileID = profile.ID
			break
		}
	}

	softgateAdd := &inventory.HWSoftgate{
		Name:        d.Get("name").(string),
		Tenant:      inventory.IDName{Name: d.Get("tenant").(string)},
		Site:        inventory.IDName{ID: d.Get("siteid").(int)},
		Description: d.Get("description").(string),
		Profile:     inventory.IDName{ID: profileID},
		MainAddress: d.Get("mainip").(string),
		MgmtAddress: d.Get("mgmtip").(string),
		Links:       []inventory.HWLink{},
	}

	js, _ := json.Marshal(softgateAdd)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.Inventory().AddSoftgate(softgateAdd)
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

	_ = d.Set("itemid", idStruct.ID)
	d.SetId(softgateAdd.Name)
	return nil
}

func resourceRead(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	sw, err := clientset.Inventory().GetByID(d.Get("itemid").(int))
	if err != nil {
		return err
	}

	d.SetId(sw.Name)
	err = d.Set("name", sw.Name)
	if err != nil {
		return err
	}
	err = d.Set("tenant", sw.Tenant.Name)
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

	return nil
}

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	profileID := 0
	profiles, err := clientset.Inventory().GetProfiles()
	if err != nil {
		return err
	}
	for _, profile := range profiles {
		if profile.Name == d.Get("profile").(string) {
			profileID = profile.ID
			break
		}
	}

	softgateUpdate := &inventory.HWSoftgateUpdate{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Tenant:      inventory.IDName{Name: d.Get("tenant").(string)},
		Site:        inventory.IDName{ID: d.Get("siteid").(int)},
		Profile:     inventory.IDName{ID: profileID},
		MainAddress: d.Get("mainip").(string),
		MgmtAddress: d.Get("mgmtip").(string),
		Links:       []inventory.HWLink{},
	}

	js, _ := json.Marshal(softgateUpdate)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.Inventory().UpdateSoftgate(d.Get("itemid").(int), softgateUpdate)
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

	reply, err := clientset.Inventory().Delete("softgate", d.Get("itemid").(int))
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

	sw, err := clientset.Inventory().GetByID(d.Get("itemid").(int))
	if err != nil {
		return false, err
	}

	if sw == nil {
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
			err := d.Set("itemid", sw.ID)
			if err != nil {
				return []*schema.ResourceData{d}, err
			}
			return []*schema.ResourceData{d}, nil
		}
	}

	return []*schema.ResourceData{d}, nil
}
