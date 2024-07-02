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

package serverclustertemplate

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
		Description: "Creates and manages server cluster template",
		Schema: map[string]*schema.Schema{
			"template": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The template object.",
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

	resourceAdd := d.Get("template").(string)

	js, _ := json.Marshal(resourceAdd)
	log.Println("[DEBUG]", resourceAdd)

	reply, err := clientset.ServerClusterTemplate().Add(resourceAdd)
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
	err = d.Set("customdata", sw.CustomData)
	if err != nil {
		return err
	}
	err = d.Set("portcount", sw.PortCount)
	if err != nil {
		return err
	}

	if asnumber := d.Get("asnumber"); asnumber.(string) != "auto" {
		err = d.Set("asnumber", strconv.Itoa(sw.Asn))
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

	return nil
}

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

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

	serverUpdate := &inventory.HWServer{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Tenant:      inventory.IDName{ID: d.Get("tenantid").(int)},
		Site:        inventory.IDName{ID: d.Get("siteid").(int)},
		MainAddress: d.Get("mainip").(string),
		MgmtAddress: d.Get("mgmtip").(string),
		Links:       sw.Links,
		PortCount:   d.Get("portcount").(int),
		Asn:         asnAny,
		CustomData:  d.Get("customdata").(string),
	}

	js, _ := json.Marshal(serverUpdate)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.Inventory().UpdateServer(id, serverUpdate)
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
	reply, err := clientset.Inventory().Delete("server", id)
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
