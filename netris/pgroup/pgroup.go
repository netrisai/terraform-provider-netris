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

package pgroup

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/netrisai/netriswebapi/http"
	"github.com/netrisai/netriswebapi/v1/types/permission"
	api "github.com/netrisai/netriswebapi/v2"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: "Creates and manages Permission Groups",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Permission Group",
			},
			"description": {
				Optional:    true,
				Default:     "",
				Type:        schema.TypeString,
				Description: "Permission Group description",
			},
			"groups": {
				Optional: true,
				Type:     schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "List of groups. Example: `[\"services.l4loadbalancer:view\"]`. Possible action value is `view` or `edit`. Addition action value `external-acl` only for key `services.acl`",
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

	groups := []string{}
	grList := d.Get("groups").([]interface{})
	for _, group := range grList {
		groups = append(groups, group.(string))
	}

	if len(groups) == 0 {
		return fmt.Errorf("Please specify the groups")
	}

	groupParameters := parseGroups(strings.Join(groups, ","))
	js, _ := json.Marshal(groupParameters)
	log.Println("[DEBUG] Group Parameters", string(js))

	exceptHidden, exceptReadOnly := makeExceptionList(groupParameters, mappings.getMap())
	js, _ = json.Marshal(exceptHidden)
	log.Println("[DEBUG] Except Hidden", string(js))
	js, _ = json.Marshal(exceptReadOnly)
	log.Println("[DEBUG] Except ReadOnly", string(js))

	hiddenList, readOnlyList := makePermLists(exceptHidden, exceptReadOnly, sectionNames)
	js, _ = json.Marshal(exceptHidden)
	log.Println("[DEBUG] Hidden List", string(js))
	js, _ = json.Marshal(exceptReadOnly)
	log.Println("[DEBUG] ReadOnly List", string(js))

	externalACl := false

	if key, ok := groupParameters["services"]; ok {
		if subkeys, ok := key["acl"]; ok {
			for _, subkey := range subkeys {
				if subkey == "external-acl" {
					externalACl = true
					break
				}
			}
		}
	}

	pAdd := &permission.PermissionGroupAdd{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		ExternalACL: externalACl,
		Hidden:      hiddenList,
		ReadOnly:    readOnlyList,
	}

	js, _ = json.Marshal(pAdd)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.Permission().Add(pAdd)
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
	var gr *permission.PermissionGroup = nil

	groups, err := clientset.Permission().Get()
	if err != nil {
		return err
	}

	for _, group := range groups {
		if group.ID == id {
			gr = group
			break
		}
	}

	if gr == nil {
		return fmt.Errorf("couldn't find permission group '%s'", d.Get("name").(string))
	}

	// Fill the data

	return nil
}

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	groups := []string{}
	grList := d.Get("groups").([]interface{})
	for _, group := range grList {
		groups = append(groups, group.(string))
	}

	if len(groups) == 0 {
		return fmt.Errorf("Please specify the groups")
	}

	groupParameters := parseGroups(strings.Join(groups, ","))
	js, _ := json.Marshal(groupParameters)
	log.Println("[DEBUG] Group Parameters", string(js))

	exceptHidden, exceptReadOnly := makeExceptionList(groupParameters, mappings.getMap())
	js, _ = json.Marshal(exceptHidden)
	log.Println("[DEBUG] Except Hidden", string(js))
	js, _ = json.Marshal(exceptReadOnly)
	log.Println("[DEBUG] Except ReadOnly", string(js))

	hiddenList, readOnlyList := makePermLists(exceptHidden, exceptReadOnly, sectionNames)
	js, _ = json.Marshal(exceptHidden)
	log.Println("[DEBUG] Hidden List", string(js))
	js, _ = json.Marshal(exceptReadOnly)
	log.Println("[DEBUG] ReadOnly List", string(js))

	externalACl := false

	if key, ok := groupParameters["services"]; ok {
		if subkeys, ok := key["acl"]; ok {
			for _, subkey := range subkeys {
				if subkey == "external-acl" {
					externalACl = true
					break
				}
			}
		}
	}

	id, _ := strconv.Atoi(d.Id())
	pAdd := &permission.PermissionGroupAdd{
		ID:          id,
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		ExternalACL: externalACl,
		Hidden:      hiddenList,
		ReadOnly:    readOnlyList,
	}

	js, _ = json.Marshal(pAdd)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.Permission().Update(pAdd)
	if err != nil {
		log.Println("[DEBUG]", err)
		return err
	}

	js, _ = json.Marshal(reply)
	log.Println("[DEBUG]", string(js))

	log.Println("[DEBUG]", string(reply.Data))

	return nil
}

func resourceDelete(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	id, _ := strconv.Atoi(d.Id())
	reply, err := clientset.Permission().Delete(id)
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
	groups, err := clientset.Permission().Get()
	if err != nil {
		return false, err
	}

	for _, group := range groups {
		if group.ID == id {
			return true, nil
		}
	}

	return false, nil
}

func resourceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	clientset := m.(*api.Clientset)

	name := d.Id()
	var gr *permission.PermissionGroup = nil

	groups, err := clientset.Permission().Get()
	if err != nil {
		return []*schema.ResourceData{d}, err
	}

	for _, group := range groups {
		if group.Name == name {
			gr = group
			d.SetId(strconv.Itoa(gr.ID))
			return []*schema.ResourceData{d}, nil
		}
	}

	if gr == nil {
		return []*schema.ResourceData{d}, fmt.Errorf("couldn't find permission group '%s'", name)
	}

	return []*schema.ResourceData{d}, nil
}
