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

package portgroup

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/netrisai/netriswebapi/http"
	"github.com/netrisai/netriswebapi/v1/types/portgroup"

	api "github.com/netrisai/netriswebapi/v2"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: "Creates and manages ACL Port Groups",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the ACL Port Group",
			},
			"ports": {
				Required:    true,
				Type:        schema.TypeSet,
				Description: "List of ports. Valid values are: single port `22`, range of ports `1024-2048`",
				Elem: &schema.Schema{
					ValidateFunc: validatePort,
					Type:         schema.TypeString,
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

	name := d.Get("name").(string)
	portList := d.Get("ports").(*schema.Set).List()
	ports := []string{}
	for _, port := range portList {
		ports = append(ports, port.(string))
	}

	pAdd := &portgroup.PortGroupW{
		Name:  name,
		Ports: ports,
	}

	js, _ := json.Marshal(pAdd)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.PortGroup().Add(pAdd)
	if err != nil {
		log.Println("[DEBUG]", err)
		return err
	}

	js, _ = json.Marshal(reply)
	log.Println("[DEBUG]", string(js))

	log.Println("[DEBUG]", string(reply.Data))

	idStruct := struct {
		ID int `json:"portGroupId"`
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

	name := d.Get("name").(string)
	id, _ := strconv.Atoi(d.Id())
	var pGroup *portgroup.PortGroup
	var ok bool
	if pGroup, ok = findPortGroupByID(id, clientset); !ok {
		return fmt.Errorf("coudn't find portgroup '%s'", name)
	}

	d.SetId(strconv.Itoa(pGroup.ID))
	err := d.Set("name", pGroup.Name)
	if err != nil {
		return err
	}
	err = d.Set("ports", pGroup.Ports)
	if err != nil {
		return err
	}
	return nil
}

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	name := d.Get("name").(string)
	portList := d.Get("ports").(*schema.Set).List()
	id, _ := strconv.Atoi(d.Id())
	ports := []string{}
	for _, port := range portList {
		ports = append(ports, port.(string))
	}

	var pGroup *portgroup.PortGroup
	var ok bool
	if pGroup, ok = findPortGroupByID(id, clientset); !ok {
		return fmt.Errorf("coudn't find portgroup '%s'", name)
	}

	forAdd, forDelete := comparePorts(ports, pGroup.Ports)

	pUpdate := &portgroup.PortGroupW{
		ID:                 id,
		Name:               name,
		Ports:              ports,
		AddedArray:         forAdd,
		DeletedElementsArr: forDelete,
	}

	js, _ := json.Marshal(pUpdate)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.PortGroup().Update(pUpdate)
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
	reply, err := clientset.PortGroup().Delete(id)
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
	_, ok := findPortGroupByID(id, clientset)
	return ok, nil
}

func resourceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	clientset := m.(*api.Clientset)

	name := d.Id()
	var pGroup *portgroup.PortGroup
	var ok bool
	if pGroup, ok = findPortGroupByName(name, clientset); !ok {
		return []*schema.ResourceData{d}, fmt.Errorf("coudn't find portgroup '%s'", name)
	}

	d.SetId(strconv.Itoa(pGroup.ID))

	return []*schema.ResourceData{d}, nil
}
