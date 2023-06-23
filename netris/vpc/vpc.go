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

package vpc

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/netrisai/netriswebapi/http"
	api "github.com/netrisai/netriswebapi/v2"
	"github.com/netrisai/netriswebapi/v2/types/vpc"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: "Creates and manages VPC",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "User assigned name of VPC.",
			},
			"tenantid": {
				Required:    true,
				Type:        schema.TypeInt,
				ForceNew:    true,
				Description: "ID of tenant. Users of this tenant will be permitted to edit this unit.",
			},
			"guesttenantid": {
				Optional:    true,
				Type:        schema.TypeSet,
				Description: "List of tenants allowed to add/remove services to the VPC but not allowed to manage other parameters of it.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "The ID of a guest tenant who is authorized to add/remove services to the VPC but not allowed to manage other parameters of it.",
						},
					},
				},
			},
			"tags": {
				Computed: true,
				Optional: true,
				Type:     schema.TypeSet,
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

	guestTenantIdsList := d.Get("guesttenantid").(*schema.Set).List()
	log.Println("[DEBUG] guestTenantIdsList", guestTenantIdsList)

	tagsList := d.Get("tags").(*schema.Set).List()
	tags := []string{}
	for _, tag := range tagsList {
		tags = append(tags, tag.(string))
	}

	vpcAdd := &vpc.VPCw{
		Name:        d.Get("name").(string),
		AdminTenant: vpc.AdminTenant{ID: d.Get("tenantid").(int)},
		GuestTenant: []vpc.GuestTenant{},
		Tags:        tags,
	}

	js, _ := json.Marshal(vpcAdd)
	log.Println("[DEBUG] vpcAdd", string(js))

	reply, err := clientset.VPC().Add(vpcAdd)
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
	apiVPC, err := clientset.VPC().GetByID(id)

	if err != nil {
		return nil
	}

	d.SetId(strconv.Itoa(apiVPC.ID))
	err = d.Set("name", apiVPC.Name)
	if err != nil {
		return err
	}
	err = d.Set("tenantid", apiVPC.AdminTenant.ID)
	if err != nil {
		return err
	}
	err = d.Set("guesttenantid", apiVPC.GuestTenant)
	if err != nil {
		return err
	}
	err = d.Set("tags", apiVPC.Tags)
	if err != nil {
		return err
	}

	return nil
}

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	vpcID, _ := strconv.Atoi(d.Id())

	guestTenantIdsList := d.Get("switchfabricproviders").(*schema.Set).List()
	log.Println("[DEBUG] guestTenantIdsList", guestTenantIdsList)

	tagsList := d.Get("tags").(*schema.Set).List()
	tags := []string{}
	for _, tag := range tagsList {
		tags = append(tags, tag.(string))
	}

	vpcUpdate := &vpc.VPCw{
		Name:        d.Get("name").(string),
		AdminTenant: vpc.AdminTenant{ID: d.Get("tenantid").(int)},
		GuestTenant: []vpc.GuestTenant{},
		Tags:        tags,
	}

	js, _ := json.Marshal(vpcUpdate)
	log.Println("[DEBUG] vpcUpdate", string(js))

	reply, err := clientset.VPC().Update(vpcID, vpcUpdate)
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
	item, _ := clientset.VPC().GetByID(id)

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

	items, _ := clientset.VPC().Get()
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
	reply, err := clientset.VPC().Delete(id)
	if err != nil {
		return err
	}

	if reply.StatusCode != 200 {
		return fmt.Errorf(string(reply.Data))
	}

	d.SetId("")
	return nil
}
