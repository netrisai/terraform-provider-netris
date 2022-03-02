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

package userrole

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/netrisai/netriswebapi/http"
	"github.com/netrisai/netriswebapi/v1/types/userrole"

	api "github.com/netrisai/netriswebapi/v2"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: "Creates and manages User Roles",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the user role.",
			},
			"description": {
				Optional:    true,
				Default:     "",
				Type:        schema.TypeString,
				Description: "User Role description",
			},
			"pgroup": {
				Required:    true,
				Type:        schema.TypeString,
				Description: "The name of existing permission group",
			},
			"tenantids": {
				Optional: true,
				Type:     schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Description: "List of tenant IDs",
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
	pgroupName := d.Get("pgroup").(string)

	pgrp, ok := findPgroupByName(pgroupName, clientset)
	if !ok {
		return fmt.Errorf("couldn't find permission group '%s'", pgroupName)
	}

	tenantIds := []int{}
	tenants := d.Get("tenantids").(*schema.Set).List()
	for _, name := range tenants {
		tenantIds = append(tenantIds, name.(int))
	}

	roleTenants := []userrole.Tenant{}
	for _, id := range tenantIds {
		roleTenants = append(roleTenants, userrole.Tenant{
			ID:          id,
			TenantRead:  true,
			TenantWrite: true,
		})
	}

	urAdd := &userrole.UserRoleAdd{
		Name:            name,
		Description:     d.Get("description").(string),
		PermissionGroup: *pgrp,
		Tenants:         roleTenants,
	}

	js, _ := json.Marshal(urAdd)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.UserRole().Add(urAdd)
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
	var ur *userrole.UserRole = nil

	uroles, err := clientset.UserRole().Get()
	if err != nil {
		return err
	}

	for _, urole := range uroles {
		if urole.ID == id {
			ur = urole
			break
		}
	}

	if ur == nil {
		return fmt.Errorf("couldn't find user role '%s'", d.Get("name").(string))
	}

	d.SetId(strconv.Itoa(ur.ID))
	err = d.Set("name", ur.Name)
	if err != nil {
		return err
	}
	err = d.Set("pgroup", ur.PermName)
	if err != nil {
		return err
	}
	err = d.Set("description", ur.Description)
	if err != nil {
		return err
	}

	tenantsList := []int{}
	for _, tenant := range ur.Tenants {
		if tenant.TenantID == 0 {
			tenantsList = append(tenantsList, -1)
		} else {
			tenantsList = append(tenantsList, tenant.TenantID)
		}
	}
	err = d.Set("tenantids", tenantsList)
	if err != nil {
		return err
	}

	return nil
}

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	name := d.Get("name").(string)
	pgroupName := d.Get("pgroup").(string)

	pgrp, ok := findPgroupByName(pgroupName, clientset)
	if !ok {
		return fmt.Errorf("couldn't find permission group '%s'", pgroupName)
	}

	tenantIds := []int{}
	tenants := d.Get("tenantids").(*schema.Set).List()
	for _, name := range tenants {
		tenantIds = append(tenantIds, name.(int))
	}

	roleTenants := []userrole.Tenant{}
	for _, id := range tenantIds {
		roleTenants = append(roleTenants, userrole.Tenant{
			ID:          id,
			TenantRead:  true,
			TenantWrite: true,
		})
	}

	id, _ := strconv.Atoi(d.Id())
	urAdd := &userrole.UserRoleAdd{
		ID:              id,
		Name:            name,
		Description:     d.Get("description").(string),
		PermissionGroup: *pgrp,
		Tenants:         roleTenants,
	}

	js, _ := json.Marshal(urAdd)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.UserRole().Update(urAdd)
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
	reply, err := clientset.UserRole().Delete(id)
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

	uroles, err := clientset.UserRole().Get()
	if err != nil {
		return false, err
	}

	for _, urole := range uroles {
		if urole.ID == id {
			return true, nil
		}
	}

	return false, nil
}

func resourceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	clientset := m.(*api.Clientset)

	name := d.Id()
	var ur *userrole.UserRole = nil

	uroles, err := clientset.UserRole().Get()
	if err != nil {
		return []*schema.ResourceData{d}, err
	}

	for _, urole := range uroles {
		if urole.Name == name {
			ur = urole
			d.SetId(strconv.Itoa(ur.ID))
			return []*schema.ResourceData{d}, nil
		}
	}

	if ur == nil {
		return []*schema.ResourceData{d}, fmt.Errorf("couldn't find user role '%s'", name)
	}

	return []*schema.ResourceData{d}, nil
}
