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

package servercluster

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/netrisai/netriswebapi/http"
	api "github.com/netrisai/netriswebapi/v2"
	"github.com/netrisai/netriswebapi/v2/types/servercluster"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: "Creates and manages ServerCluster",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "User assigned name of ServerCluster.",
			},
			"adminid": {
				Required:    true,
				Type:        schema.TypeInt,
				ForceNew:    true,
				Description: "ID of Admin tenant. Users of this tenant will be permitted to edit this unit.",
			},
			"siteid": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "The site ID where this ServerCluster belongs.",
			},
			"serverclusterid": {
				ForceNew:    true,
				Optional:    true,
				Type:        schema.TypeInt,
				Default: 0,
				Description: "ID of VPC. If not specified, a new VPC will be created.",
			},
			"template": {
				ForceNew:    true,
				Required:    true,
				Type:        schema.TypeInt,
				Description: "ID of Server Cluster Template.",
			},
			"tags": {
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
	log.Println("[DEBUG] serverclusterCreate")
	clientset := m.(*api.Clientset)

	guestTenantIdsList := d.Get("guesttenantid").(*schema.Set).List()
	log.Println("[DEBUG] guestTenantIdsList", guestTenantIdsList)
	guestTenants := []servercluster.GuestTenant{}

	for _, eachGuestTenant := range guestTenantIdsList {
		gTenant := eachGuestTenant.(map[string]interface{})
		guestTenants = append(guestTenants, servercluster.GuestTenant{
			ID: gTenant["id"].(int),
		})
	}

	tagsList := d.Get("tags").(*schema.Set).List()
	tags := []string{}
	for _, tag := range tagsList {
		tags = append(tags, tag.(string))
	}

	serverclusterAdd := &servercluster.ServerClusterw{
		Name:        d.Get("name").(string),
		AdminTenant: servercluster.AdminTenant{ID: d.Get("tenantid").(int)},
		GuestTenant: guestTenants,
		Tags:        tags,
	}

	js, _ := json.Marshal(serverclusterAdd)
	log.Println("[DEBUG] serverclusterAdd", string(js))

	reply, err := clientset.ServerCluster().Add(serverclusterAdd)
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
	log.Println("[DEBUG] serverclusterRead")
	clientset := m.(*api.Clientset)

	id, _ := strconv.Atoi(d.Id())
	apiServerCluster, err := clientset.ServerCluster().GetByID(id)

	if err != nil {
		return nil
	}

	d.SetId(strconv.Itoa(apiServerCluster.ID))
	err = d.Set("name", apiServerCluster.Name)
	if err != nil {
		return err
	}
	err = d.Set("tenantid", apiServerCluster.AdminTenant.ID)
	if err != nil {
		return err
	}

	log.Println("[DEBUG] serverclusterRead apiServerCluster.GuestTenant", apiServerCluster.GuestTenant)

	var gTenantsList []map[string]interface{}
	for _, gTenant := range apiServerCluster.GuestTenant {
		gt := make(map[string]interface{})
		gt["id"] = gTenant.ID
		gTenantsList = append(gTenantsList, gt)
	}

	err = d.Set("guesttenantid", gTenantsList)
	if err != nil {
		return err
	}
	err = d.Set("tags", apiServerCluster.Tags)
	if err != nil {
		return err
	}

	return nil
}

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	log.Println("[DEBUG] serverclusterUpdate")
	clientset := m.(*api.Clientset)

	serverclusterID, _ := strconv.Atoi(d.Id())

	guestTenantIdsList := d.Get("guesttenantid").(*schema.Set).List()
	log.Println("[DEBUG] guestTenantIdsList", guestTenantIdsList)
	guestTenants := []servercluster.GuestTenant{}

	for _, eachGuestTenant := range guestTenantIdsList {
		gTenant := eachGuestTenant.(map[string]interface{})
		guestTenants = append(guestTenants, servercluster.GuestTenant{
			ID: gTenant["id"].(int),
		})
	}

	tagsList := d.Get("tags").(*schema.Set).List()
	tags := []string{}
	for _, tag := range tagsList {
		tags = append(tags, tag.(string))
	}

	serverclusterUpdate := &servercluster.ServerClusterw{
		Name:        d.Get("name").(string),
		AdminTenant: servercluster.AdminTenant{ID: d.Get("tenantid").(int)},
		GuestTenant: guestTenants,
		Tags:        tags,
	}

	js, _ := json.Marshal(serverclusterUpdate)
	log.Println("[DEBUG] serverclusterUpdate", string(js))

	reply, err := clientset.ServerCluster().Update(serverclusterID, serverclusterUpdate)
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
	item, err := clientset.ServerCluster().GetByID(id)
	if err != nil {
		log.Println("[DEBUG] serverclusterExist response err:", err)
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

	items, _ := clientset.ServerCluster().Get()
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
	reply, err := clientset.ServerCluster().Delete(id)
	if err != nil {
		return err
	}

	if reply.StatusCode != 200 {
		return fmt.Errorf(string(reply.Data))
	}

	d.SetId("")
	return nil
}
