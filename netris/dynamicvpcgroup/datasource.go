/*
Copyright 2026. Netris, Inc.

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

package dynamicvpcgroup

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	api "github.com/netrisai/netriswebapi/v2"
)

func DataResource() *schema.Resource {
	return &schema.Resource{
		Description: "Data Source: Dynamic VPC Group. Credentials are not returned.",
		Schema: map[string]*schema.Schema{
			"groupid": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resourcegroupid": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"radiusserver": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"displayid": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"member": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vpcid": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"vpcname": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"vni": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"vrfname": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"user": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"username": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"server": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"authport": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"acctport": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"radiusclients": radiusClientSchema(),
		},
		Read: dataResourceRead,
	}
}

func dataResourceRead(d *schema.ResourceData, m interface{}) error {
	id := d.Get("groupid").(int)
	item, err := m.(*api.Clientset).DynamicVPCGroup().GetByID(id)
	if err != nil {
		return err
	}
	if item == nil {
		return fmt.Errorf("Dynamic VPC Group %d was not found", id)
	}
	return flatten(d, item, false, false)
}
