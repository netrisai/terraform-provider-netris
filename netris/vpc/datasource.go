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

package vpc

import (
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	api "github.com/netrisai/netriswebapi/v2"
	"github.com/netrisai/netriswebapi/v2/types/vpc"
)

func DataResource() *schema.Resource {
	return &schema.Resource{
		Description: "Data Source: VPC",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the vpc",
			},
			"tenantid": {
				Optional:    true,
				Type:        schema.TypeInt,
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
				Optional: true,
				Type:     schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
		Read:   dataResourceRead,
		Exists: dataResourceExists,
	}
}

func dataResourceRead(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	name := d.Get("name").(string)

	var apiVPC *vpc.VPC

	list, err := clientset.VPC().Get()
	if err != nil {
		return err
	}

	for _, v := range list {
		if v.Name == name {
			apiVPC = v
			break
		}
	}

	if apiVPC == nil {
		return fmt.Errorf("coudn't find vpc '%s';", name)
	}

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

	var gTenantsList []map[string]interface{}
	for _, gTenant := range apiVPC.GuestTenant {
		gt := make(map[string]interface{})
		gt["id"] = gTenant.ID
		gTenantsList = append(gTenantsList, gt)
	}

	err = d.Set("guesttenantid", gTenantsList)
	if err != nil {
		return err
	}
	err = d.Set("tags", apiVPC.Tags)
	if err != nil {
		return err
	}

	return nil
}

func dataResourceExists(d *schema.ResourceData, m interface{}) (bool, error) {
	clientset := m.(*api.Clientset)

	id, _ := strconv.Atoi(d.Id())
	item, err := clientset.VPC().GetByID(id)
	if err != nil {
		log.Println("[DEBUG] vpcExist response err:", err)
	}

	if item == nil {
		return false, nil
	}
	if item.ID > 0 {
		return true, nil
	}

	return false, nil
}
