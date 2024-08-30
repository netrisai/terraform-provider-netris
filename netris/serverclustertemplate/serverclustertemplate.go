/*
Copyright 2023. Netris, Inc.

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

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/netrisai/netriswebapi/http"
	api "github.com/netrisai/netriswebapi/v2"
	"github.com/netrisai/netriswebapi/v2/types/serverclustertemplate"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: "Creates and manages ServerClusterTemplate",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "User assigned name of ServerClusterTemplate.",
			},
			"vnets": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Server Cluster VNets Template",
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
	log.Println("[DEBUG] serverClusterTemplateCreate")
	clientset := m.(*api.Clientset)
	var vnetsUnmarshaled interface{}

	jsonString := d.Get("vnets").(string)
	// Convert the JSON string to the interface{}
	err := json.Unmarshal([]byte(jsonString), &vnetsUnmarshaled)
	if err != nil {
		log.Println("[DEBUG] Error converting JSON string to interface{}: ", err)
		return err
	}

	vnetsSlice, ok := vnetsUnmarshaled.([]interface{})
	if !ok {
		log.Println("[DEBUG] Expected []interface{} but got something else")
		return err
	}

	serverClusterTemplateAdd := &serverclustertemplate.ServerClusterTemplateW{
		Name:  d.Get("name").(string),
		Vnets: vnetsSlice,
	}

	js, _ := json.Marshal(serverClusterTemplateAdd)
	log.Println("[DEBUG] serverClusterTemplateAdd", string(js))

	reply, err := clientset.ServerClusterTemplate().Add(serverClusterTemplateAdd)
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
	log.Println("[DEBUG] serverClusterTemplateRead")
	clientset := m.(*api.Clientset)

	id, _ := strconv.Atoi(d.Id())
	apiServerClusterTemplate, err := clientset.ServerClusterTemplate().GetByID(id)

	if err != nil {
		return nil
	}

	d.SetId(strconv.Itoa(apiServerClusterTemplate.ID))
	err = d.Set("name", apiServerClusterTemplate.Name)
	if err != nil {
		return err
	}

	netrisVNETs := apiServerClusterTemplate.Vnets
	for _, item := range netrisVNETs {
		if itemMap, ok := item.(map[string]interface{}); ok {
			delete(itemMap, "id")
		}
	}

	jsonVNETs, err := json.Marshal(netrisVNETs)

	if err != nil {
		log.Fatalf("Error marshalling data to JSON: %v", err)
	}

	err = d.Set("vnets", string(jsonVNETs))
	if err != nil {
		return err
	}

	return nil
}

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	log.Println("[DEBUG] serverClusterTemplateUpdate")
	clientset := m.(*api.Clientset)

	serverClusterTemplateID, _ := strconv.Atoi(d.Id())

	var vnetsUnmarshaled interface{}

	jsonString := d.Get("vnets").(string)
	// Convert the JSON string to the interface{}
	err := json.Unmarshal([]byte(jsonString), &vnetsUnmarshaled)
	if err != nil {
		log.Println("[DEBUG] Error converting JSON string to interface{}: ", err)
		return err
	}

	// If you need to work with the data, you can use type assertions
	vnetsSlice, ok := vnetsUnmarshaled.([]interface{})
	if !ok {
		log.Println("[DEBUG] Expected []interface{} but got something else")
		return err
	}

	serverClusterTemplateUpdate := &serverclustertemplate.ServerClusterTemplateW{
		Name:  d.Get("name").(string),
		Vnets: vnetsSlice,
	}

	js, _ := json.Marshal(serverClusterTemplateUpdate)
	log.Println("[DEBUG] serverClusterTemplateUpdate", string(js))

	reply, err := clientset.ServerClusterTemplate().Update(serverClusterTemplateID, serverClusterTemplateUpdate)
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

	id, _ := strconv.Atoi(d.Id())
	item, err := clientset.ServerClusterTemplate().GetByID(id)
	if err != nil {
		log.Println("[DEBUG] serverClusterTemplateExist response err:", err)
	}

	if item == nil {
		return false, nil
	}

	if item.ID > 0 {
		return true, nil
	}

	return false, nil
}

func resourceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	clientset := m.(*api.Clientset)

	items, _ := clientset.ServerClusterTemplate().Get()
	name := d.Id()
	for _, item := range items {
		if item.Name == name {
			d.SetId(strconv.Itoa(item.ID))
			return []*schema.ResourceData{d}, nil
		}
	}

	return []*schema.ResourceData{d}, nil
}

func resourceDelete(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	id, _ := strconv.Atoi(d.Id())
	reply, err := clientset.ServerClusterTemplate().Delete(id)
	if err != nil {
		return err
	}

	if reply.StatusCode != 200 {
		return fmt.Errorf(string(reply.Data))
	}

	d.SetId("")
	return nil
}
