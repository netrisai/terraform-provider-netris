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

package bgpobject

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/netrisai/netriswebapi/http"
	"github.com/netrisai/netriswebapi/v1/types/bgpobject"

	api "github.com/netrisai/netriswebapi/v2"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: "Creates and manages BGP Objects",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the BGP Object",
			},
			"type": {
				ValidateFunc: validateType,
				Required:     true,
				Type:         schema.TypeString,
				Description:  "BGP Objects type. Possible values: `ipv4`, `ipv6`, `aspath`, `community`, `extended`, `large`. Detailed documentation about objects types is available [here](https://www.netris.ai/docs/en/stable/network-policies.html#bgp-objects)",
			},
			"value": {
				Required:    true,
				Type:        schema.TypeString,
				Description: "Object value. For type `ipv4`, `ipv6` value can be multiline",
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

	name := d.Get("name").(string)
	typo := d.Get("type").(string)
	value := d.Get("value").(string)

	objectAdd := &bgpobject.BGPObjectW{
		Name:      name,
		Type:      typo,
		TypeValue: value,
	}

	js, _ := json.Marshal(objectAdd)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.BGPObject().Add(objectAdd)
	if err != nil {
		log.Println("[DEBUG]", err)
		return err
	}

	js, _ = json.Marshal(reply)
	log.Println("[DEBUG]", string(js))

	log.Println("[DEBUG]", string(reply.Data))

	id := 0

	data, err := reply.Parse()
	if err != nil {
		log.Println("[DEBUG]", err)
		return err
	}

	err = http.Decode(data.Data, &id)
	if err != nil {
		log.Println("[DEBUG]", err)
		return err
	}

	log.Println("[DEBUG] ID:", id)

	if reply.StatusCode != 200 {
		return fmt.Errorf(string(reply.Data))
	}

	d.SetId(strconv.Itoa(id))

	return nil
}

func resourceRead(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)
	id, _ := strconv.Atoi(d.Id())
	obj, ok := findByID(id, clientset)
	if !ok {
		return fmt.Errorf("Coudn't find bgp object '%s'", d.Get("name").(string))
	}

	d.SetId(strconv.Itoa(obj.ID))
	err := d.Set("name", obj.Name)
	if err != nil {
		return err
	}
	err = d.Set("type", obj.Type)
	if err != nil {
		return err
	}
	err = d.Set("value", obj.TypeValue)
	if err != nil {
		return err
	}
	return nil
}

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	name := d.Get("name").(string)
	typo := d.Get("type").(string)
	value := d.Get("value").(string)
	id, _ := strconv.Atoi(d.Id())
	objectUpdate := &bgpobject.BGPObjectW{
		ID:        id,
		Name:      name,
		Type:      typo,
		TypeValue: value,
	}

	js, _ := json.Marshal(objectUpdate)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.BGPObject().Update(objectUpdate)
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
	reply, err := clientset.BGPObject().Delete(id)
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
	_, ok := findByID(id, clientset)

	return ok, nil
}

func resourceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	clientset := m.(*api.Clientset)
	name := d.Id()
	var obj *bgpobject.BGPObject
	var ok bool
	obj, ok = findByName(name, clientset)
	if !ok {
		return []*schema.ResourceData{d}, fmt.Errorf("Coudn't find bgp object '%s'", d.Get("name").(string))
	}
	d.SetId(strconv.Itoa(obj.ID))

	return []*schema.ResourceData{d}, nil
}
