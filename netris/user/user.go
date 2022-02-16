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

package user

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/netrisai/netriswebapi/http"
	"github.com/netrisai/netriswebapi/v1/types/permission"
	"github.com/netrisai/netriswebapi/v1/types/user"
	"github.com/netrisai/netriswebapi/v1/types/userrole"

	api "github.com/netrisai/netriswebapi/v2"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: "Creates and manages Users",
		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Unique username.",
			},
			"fullname": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Full Name of the user.",
			},
			"email": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The email address of the user. Also used for system notifications and for password retrieval.",
			},
			"emailcc": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Send copies of email notifications to this address.",
			},
			"phone": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "User’s phone number.",
			},
			"company": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Company the user works for. Usually useful for multi-tenant systems where the company provides Netris Controller access to customers.",
			},
			"position": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Position within the company.",
			},
			"userrole": {
				Required:    true,
				Type:        schema.TypeString,
				Description: "Name of User Role. When using a User Role object to define RBAC (role-based access control), `pgroup` and `tenants` fields will be ignoring.",
			},
			"pgroup": {
				Required:    true,
				Type:        schema.TypeString,
				Description: "Name of Permission Group. User permissions for viewing and editing parts of the Netris Controller. (if User Role is not used).",
			},
			"tenants": {
				Optional: true,
				Type:     schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "List of tenants. (if User Role is not used).",
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

	var (
		username    = d.Get("username").(string)
		fullname    = d.Get("fullname").(string)
		email       = d.Get("email").(string)
		emailcc     = d.Get("emailcc").(string)
		phone       = d.Get("phone").(string)
		company     = d.Get("company").(string)
		position    = d.Get("position").(string)
		role        *userrole.UserRole
		pgrp        = &permission.PermissionGroup{}
		roleTenants = []userrole.Tenant{}
	)

	uroleName := d.Get("userrole").(string)
	pgroupName := d.Get("pgroup").(string)

	if uroleName != "" {
		var ok bool
		role, ok = findRoleByName(uroleName, clientset)
		if !ok {
			return fmt.Errorf("couldn't find user role'%s'", uroleName)
		}
	}

	if role == nil {
		role = &userrole.UserRole{}
		var ok bool
		pgrp, ok = findPgroupByName(pgroupName, clientset)
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

		for _, tenant := range netrisTenants {
			roleTenants = append(roleTenants, userrole.Tenant{
				ID:          tenant.ID,
				TenantName:  tenant.Name,
				TenantRead:  true,
				TenantWrite: true,
			})
		}
	}

	uAdd := &user.UserAdd{
		Name:            username,
		Fullname:        fullname,
		Email:           email,
		EmailCc:         emailcc,
		Phonenumber:     phone,
		Company:         company,
		Position:        position,
		UserRole:        *role,
		PermissionGroup: *pgrp,
		Tenants:         roleTenants,
	}

	js, _ := json.Marshal(uAdd)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.User().Add(uAdd)
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
	var u *user.User = nil

	users, err := clientset.User().Get()
	if err != nil {
		return err
	}

	for _, user := range users {
		if user.ID == id {
			u = user
			break
		}
	}

	if u == nil {
		return fmt.Errorf("couldn't find user '%s'", d.Get("username").(string))
	}

	d.SetId(strconv.Itoa(u.ID))
	err = d.Set("username", u.Name)
	if err != nil {
		return err
	}
	err = d.Set("fullname", u.Fullname)
	if err != nil {
		return err
	}
	err = d.Set("email", u.Email)
	if err != nil {
		return err
	}
	err = d.Set("emailcc", u.EmailCc)
	if err != nil {
		return err
	}
	err = d.Set("phone", u.Phone)
	if err != nil {
		return err
	}
	err = d.Set("company", u.Company)
	if err != nil {
		return err
	}
	err = d.Set("position", u.Position)
	if err != nil {
		return err
	}
	err = d.Set("userrole", u.Rolename)
	if err != nil {
		return err
	}
	err = d.Set("pgroup", u.PermName)
	if err != nil {
		return err
	}

	var tenantsList []interface{}
	for _, tenant := range u.Tenants {
		if tenant.ID > 0 {
			tenantsList = append(tenantsList, tenant.Name)
		}
	}
	err = d.Set("tenants", tenantsList)
	if err != nil {
		return err
	}

	return nil
}

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	var (
		username    = d.Get("username").(string)
		fullname    = d.Get("fullname").(string)
		email       = d.Get("email").(string)
		emailcc     = d.Get("emailcc").(string)
		phone       = d.Get("phone").(string)
		company     = d.Get("company").(string)
		position    = d.Get("position").(string)
		role        *userrole.UserRole
		pgrp        = &permission.PermissionGroup{}
		roleTenants = []userrole.Tenant{}
	)

	uroleName := d.Get("userrole").(string)
	pgroupName := d.Get("pgroup").(string)

	if uroleName != "" {
		var ok bool
		role, ok = findRoleByName(uroleName, clientset)
		if !ok {
			return fmt.Errorf("couldn't find user role'%s'", uroleName)
		}
	}

	if role == nil {
		role = &userrole.UserRole{}
		var ok bool
		pgrp, ok = findPgroupByName(pgroupName, clientset)
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

		for _, tenant := range netrisTenants {
			roleTenants = append(roleTenants, userrole.Tenant{
				ID:          tenant.ID,
				TenantName:  tenant.Name,
				TenantRead:  true,
				TenantWrite: true,
			})
		}
	}

	id, _ := strconv.Atoi(d.Id())
	uAdd := &user.UserAdd{
		ID:              id,
		Name:            username,
		Fullname:        fullname,
		Email:           email,
		EmailCc:         emailcc,
		Phonenumber:     phone,
		Company:         company,
		Position:        position,
		UserRole:        *role,
		PermissionGroup: *pgrp,
		Tenants:         roleTenants,
	}

	js, _ := json.Marshal(uAdd)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.User().Update(uAdd)
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
	reply, err := clientset.User().Delete(id)
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

	users, err := clientset.User().Get()
	if err != nil {
		return false, err
	}

	for _, user := range users {
		if user.ID == id {
			return true, nil
		}
	}

	return false, nil
}

func resourceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	clientset := m.(*api.Clientset)

	name := d.Id()
	var u *user.User = nil

	users, err := clientset.User().Get()
	if err != nil {
		return []*schema.ResourceData{d}, err
	}

	for _, user := range users {
		if user.Name == name {
			u = user
			d.SetId(strconv.Itoa(u.ID))
			return []*schema.ResourceData{d}, nil
		}
	}

	if u == nil {
		return []*schema.ResourceData{d}, fmt.Errorf("couldn't find user '%s'", name)
	}

	return []*schema.ResourceData{d}, nil
}
