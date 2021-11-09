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

	"github.com/netrisai/netriswebapi/http"
	"github.com/netrisai/netriswebapi/v1/types/userrole"

	api "github.com/netrisai/netriswebapi/v2"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"itemid": {
				Type:             schema.TypeInt,
				Optional:         true,
				DiffSuppressFunc: DiffSuppress,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the resource, also acts as it's unique ID",
			},
			"pgroup": {
				Required: true,
				Type:     schema.TypeString,
			},
			"tenants": {
				Optional: true,
				Type:     schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
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
	pgroupName := d.Get("pgroup").(string)

	pgrp, ok := findPgroupByName(pgroupName, clientset)
	if !ok {
		return fmt.Errorf("couldn't find permission group '%s'", pgroupName)
	}

	tenantNames := []string{}
	tenants := d.Get("tenants").([]interface{})
	for _, name := range tenants {
		tenantNames = append(tenantNames, name.(string))
	}

	netrisTenants, err := findTenatsByNames(tenantNames, clientset)
	if err != nil {
		log.Println("[DEBUG]", err)
		return err
	}

	roleTenants := []userrole.Tenant{}
	for _, tenant := range netrisTenants {
		roleTenants = append(roleTenants, userrole.Tenant{
			ID:          tenant.ID,
			TenantName:  tenant.Name,
			TenantRead:  true,
			TenantWrite: true,
		})
	}

	urAdd := &userrole.UserRoleAdd{
		Name:            name,
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

	_ = d.Set("itemid", idStruct.ID)
	d.SetId(urAdd.Name)

	return nil
}

func resourceRead(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	id := d.Get("itemid").(int)
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

	err = d.Set("name", ur.Name)
	if err != nil {
		return err
	}
	err = d.Set("pgroup", ur.PermName)
	if err != nil {
		return err
	}

	var tenantsList []interface{}
	for _, tenant := range ur.Tenants {
		tenantsList = append(tenantsList, tenant.TenantName)
	}
	err = d.Set("tenants", tenantsList)
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

	tenantNames := []string{}
	tenants := d.Get("tenants").([]interface{})
	for _, name := range tenants {
		tenantNames = append(tenantNames, name.(string))
	}

	netrisTenants, err := findTenatsByNames(tenantNames, clientset)
	if err != nil {
		log.Println("[DEBUG]", err)
		return err
	}

	roleTenants := []userrole.Tenant{}
	for _, tenant := range netrisTenants {
		roleTenants = append(roleTenants, userrole.Tenant{
			ID:          tenant.ID,
			TenantName:  tenant.Name,
			TenantRead:  true,
			TenantWrite: true,
		})
	}

	urAdd := &userrole.UserRoleAdd{
		ID:              d.Get("itemid").(int),
		Name:            name,
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

	reply, err := clientset.UserRole().Delete(d.Get("itemid").(int))
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

	id := d.Get("itemid").(int)
	var ur *userrole.UserRole = nil

	uroles, err := clientset.UserRole().Get()
	if err != nil {
		return false, err
	}

	for _, urole := range uroles {
		if urole.ID == id {
			ur = urole
			break
		}
	}

	if ur == nil {
		return false, fmt.Errorf("couldn't find user role '%s'", d.Get("name").(string))
	}

	return true, nil
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
			err := d.Set("itemid", ur.ID)
			if err != nil {
				return []*schema.ResourceData{d}, err
			}
			return []*schema.ResourceData{d}, nil
		}
	}

	if ur == nil {
		return []*schema.ResourceData{d}, fmt.Errorf("couldn't find user role '%s'", name)
	}

	return []*schema.ResourceData{d}, nil
}
