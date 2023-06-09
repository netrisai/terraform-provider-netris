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

package tenant

import (
	"strconv"

	api "github.com/netrisai/netriswebapi/v2"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func DataResource() *schema.Resource {
	return &schema.Resource{
		Description: "Data Source: Tenants",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the tenant",
			},
			"description": {
				Optional:    true,
				Type:        schema.TypeString,
				ForceNew:    true,
				Description: "Tenant's description",
			},
		},
		Read:   dataResourceRead,
		Exists: dataResourceExists,
	}
}

func dataResourceRead(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	tenants, err := clientset.Tenant().Get()
	if err != nil {
		return err
	}

	for _, tenant := range tenants {
		if tenant.Name == d.Get("name").(string) {
			d.SetId(strconv.Itoa(tenant.ID))
			err = d.Set("name", tenant.Name)
			if err != nil {
				return err
			}
			if tenant.Description != "" || d.Get("description").(string) != "" {
				err = d.Set("description", tenant.Description)
				if err != nil {
					return err
				}
			}
			break
		}
	}
	return nil
}

func dataResourceExists(d *schema.ResourceData, m interface{}) (bool, error) {
	clientset := m.(*api.Clientset)

	tenantID := 0
	tenants, err := clientset.Tenant().Get()
	if err != nil {
		return false, err
	}

	id, _ := strconv.Atoi(d.Get("name").(string))
	for _, tenant := range tenants {
		if tenant.ID == id {
			tenantID = tenant.ID
			break
		}
	}

	if tenantID == 0 {
		return false, nil
	}

	return true, nil
}
