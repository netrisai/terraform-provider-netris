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
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/netrisai/netriswebapi/http"
	api "github.com/netrisai/netriswebapi/v2"
	"github.com/netrisai/netriswebapi/v2/types/dynamicvpcgroup"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: "Creates and manages a Dynamic VPC Group.",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateName,
			},
			"resourcegroupid": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"radiusserver": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"netris-hosted", "third-party"}, false),
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"displayid": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"member": {
				Type:        schema.TypeSet,
				Required:    true,
				MinItems:    1,
				Description: "Full desired VPC members and their hosted users.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vpcid": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntAtLeast(1),
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
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"username": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validateUsername,
									},
									"password": {
										Type:         schema.TypeString,
										Optional:     true,
										Sensitive:    true,
										ValidateFunc: validatePassword,
										Description:  "Required for a new username; empty preserves an existing password. Stored in Terraform state.",
									},
								},
							},
						},
					},
				},
			},
			"server": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    3,
				Description: "Ordered RADIUS servers; order determines priority.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.IsIPAddress,
						},
						"authport": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      1812,
							ValidateFunc: validation.IntBetween(1, 65535),
						},
						"acctport": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      1813,
							ValidateFunc: validation.IntBetween(1, 65535),
						},
						"secret": {
							ValidateFunc: validateSecret,
							Type:         schema.TypeString,
							Optional:     true,
							Sensitive:    true,
							Description:  "Required on create; empty preserves an existing secret when the server can be matched. Stored in Terraform state.",
						},
					},
				},
			},
			"radiusclients": radiusClientSchema(),
		},
		CustomizeDiff: func(d *schema.ResourceDiff, m interface{}) error {
			if !d.NewValueKnown("radiusserver") || !d.NewValueKnown("member") || !d.NewValueKnown("server") {
				return nil
			}
			// Set hashes and nested values can remain unknown until VPCs are created.
			keys := append(d.GetChangedKeysPrefix("member."), d.GetChangedKeysPrefix("server.")...)
			for _, key := range keys {
				if strings.HasSuffix(key, ".#") || strings.HasSuffix(key, ".vpcid") || strings.HasSuffix(key, ".username") || strings.HasSuffix(key, ".host") || strings.HasSuffix(key, ".authport") {
					if !d.NewValueKnown(key) {
						return nil
					}
				}
			}
			return validateCollections(d.Get("radiusserver").(string), d.Get("member").(*schema.Set).List(), d.Get("server").([]interface{}))
		},
		Create: resourceCreate,
		Read:   resourceRead,
		Update: resourceUpdate,
		Delete: resourceDelete,
		Importer: &schema.ResourceImporter{
			State: resourceImport,
		},
	}
}

func radiusClientSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"nasidentifier": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"nasip": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"switchid": {
					Type:     schema.TypeInt,
					Computed: true,
				},
				"switchname": {
					Type:     schema.TypeString,
					Computed: true,
				},
			},
		},
	}
}

func resourceCreate(d *schema.ResourceData, m interface{}) error {
	members := d.Get("member").(*schema.Set).List()
	servers := d.Get("server").([]interface{})
	if err := validateCollections(d.Get("radiusserver").(string), members, servers); err != nil {
		return err
	}
	memberWrites, err := expandMembers(members, nil)
	if err != nil {
		return err
	}
	serverWrites, err := expandServers(servers, nil, true)
	if err != nil {
		return err
	}
	group := &dynamicvpcgroup.DynamicVPCGroupAdd{
		Name:          d.Get("name").(string),
		ResourceGroup: dynamicvpcgroup.IDName{ID: d.Get("resourcegroupid").(int)},
		RadiusServer:  d.Get("radiusserver").(string),
		Tags:          expandTags(d),
		Members:       memberWrites,
	}
	if group.RadiusServer == "third-party" {
		group.Servers = serverWrites
	}
	reply, err := m.(*api.Clientset).DynamicVPCGroup().Add(group)
	if err != nil {
		return err
	}
	if err := checkReply(reply); err != nil {
		return err
	}
	var result struct {
		Data struct {
			ID int `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(reply.Data, &result); err != nil {
		return fmt.Errorf("Dynamic VPC Group create returned an invalid response: %s", err)
	}
	if result.Data.ID <= 0 {
		return fmt.Errorf("Dynamic VPC Group create returned an invalid ID")
	}
	d.SetId(strconv.Itoa(result.Data.ID))
	return readGroup(d, m, true)
}

func resourceRead(d *schema.ResourceData, m interface{}) error {
	return readGroup(d, m, false)
}

func readGroup(d *schema.ResourceData, m interface{}, newUsers bool) error {
	id, err := strconv.Atoi(d.Id())
	if err != nil || id <= 0 {
		return fmt.Errorf("Dynamic VPC Group ID must be a positive integer")
	}
	item, err := m.(*api.Clientset).DynamicVPCGroup().GetByID(id)
	if err != nil {
		return err
	}
	if item == nil {
		d.SetId("")
		return nil
	}
	return flatten(d, item, true, newUsers)
}

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	d.Partial(true)
	id, err := strconv.Atoi(d.Id())
	if err != nil || id <= 0 {
		return fmt.Errorf("Dynamic VPC Group ID must be a positive integer")
	}
	members := d.Get("member").(*schema.Set).List()
	servers := d.Get("server").([]interface{})
	if err := validateCollections(d.Get("radiusserver").(string), members, servers); err != nil {
		return err
	}
	old, _ := d.GetChange("member")
	writes, err := expandMembers(members, old.(*schema.Set).List())
	if err != nil {
		return err
	}
	group := &dynamicvpcgroup.DynamicVPCGroupUpdate{
		Name:    d.Get("name").(string),
		Tags:    expandTags(d),
		Members: writes,
	}
	if d.HasChange("server") {
		old, _ := d.GetChange("server")
		writes, err := expandServers(servers, old.([]interface{}), false)
		if err != nil {
			return err
		}
		group.Servers = writes
	}
	reply, err := m.(*api.Clientset).DynamicVPCGroup().Update(id, group)
	if err != nil {
		return err
	}
	if err := checkReply(reply); err != nil {
		return err
	}
	if err := readGroup(d, m, true); err != nil {
		return err
	}
	d.Partial(false)
	return nil
}

func resourceDelete(d *schema.ResourceData, m interface{}) error {
	id, err := strconv.Atoi(d.Id())
	if err != nil || id <= 0 {
		return fmt.Errorf("Dynamic VPC Group ID must be a positive integer")
	}
	reply, err := m.(*api.Clientset).DynamicVPCGroup().Delete(id)
	if err != nil {
		return err
	}
	if reply.StatusCode != 404 && reply.StatusCode != 204 {
		if err := checkReply(reply); err != nil {
			return err
		}
	}
	d.SetId("")
	return nil
}

func resourceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	for _, digit := range d.Id() {
		if digit < '0' || digit > '9' {
			return nil, fmt.Errorf("import Dynamic VPC Group using its positive numeric ID")
		}
	}
	id, err := strconv.Atoi(d.Id())
	if err != nil || id <= 0 {
		return nil, fmt.Errorf("import Dynamic VPC Group using its positive numeric ID")
	}
	d.SetId(strconv.Itoa(id))
	return []*schema.ResourceData{d}, nil
}

func checkReply(reply http.HTTPReply) error {
	result, err := reply.Parse()
	if err != nil {
		return fmt.Errorf("Dynamic VPC Group request failed (HTTP %d): %s", reply.StatusCode, err)
	}
	if result == nil {
		return fmt.Errorf("Dynamic VPC Group request returned an invalid response (HTTP %d)", reply.StatusCode)
	}
	if reply.StatusCode < 200 || reply.StatusCode >= 300 || !result.IsSuccess {
		return fmt.Errorf("Dynamic VPC Group request failed (HTTP %d): %s", reply.StatusCode, result.Message)
	}
	return nil
}
