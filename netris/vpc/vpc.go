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
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	api "github.com/netrisai/netriswebapi/v2"
	"github.com/netrisai/netriswebapi/v2/types/vpc"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: "Creates and manages VPC",
		Schema: map[string]*schema.Schema{
			"vni": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateVNI,
				Description:  "L3VNI: auto or a number from 1 to 16777215. Requires Controller 4.21 or later.",
			},
			"importrt": routeTargetSchema(false),
			"exportrt": routeTargetSchema(false),
			"vrfname": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "VRF name supplied by the controller.",
			},
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
				Description: "Tenant allowed to add/remove services to the VPC but not allowed to manage other parameters of it.",
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
	log.Println("[DEBUG] vpcCreate")
	clientset := m.(*api.Clientset)

	guestTenantIdsList := d.Get("guesttenantid").(*schema.Set).List()
	log.Println("[DEBUG] guestTenantIdsList", guestTenantIdsList)
	guestTenants := []vpc.GuestTenant{}

	for _, eachGuestTenant := range guestTenantIdsList {
		gTenant := eachGuestTenant.(map[string]interface{})
		guestTenants = append(guestTenants, vpc.GuestTenant{
			ID: gTenant["id"].(int),
		})
	}

	tagsList := d.Get("tags").(*schema.Set).List()
	tags := []string{}
	for _, tag := range tagsList {
		tags = append(tags, tag.(string))
	}

	vpcAdd := &vpc.VPCw{
		VNI:         expandVNI(d),
		Name:        d.Get("name").(string),
		AdminTenant: vpc.AdminTenant{ID: d.Get("tenantid").(int)},
		GuestTenant: guestTenants,
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

	apiVPC, err := parseWriteReply(reply)
	if err != nil {
		return err
	}
	if apiVPC.ID <= 0 {
		return fmt.Errorf("VPC response is missing a valid ID")
	}

	d.SetId(strconv.Itoa(apiVPC.ID))
	if routes := expandRouteTargets(d); routes != nil {
		vpcAdd.VNI = nil
		vpcAdd.RouteTargets = routes
		reply, err := clientset.VPC().Update(apiVPC.ID, vpcAdd)
		if err != nil {
			return err
		}
		if _, err := parseWriteReply(reply); err != nil {
			return err
		}
	}
	return resourceRead(d, m)
}

func resourceRead(d *schema.ResourceData, m interface{}) error {
	log.Println("[DEBUG] vpcRead")
	clientset := m.(*api.Clientset)

	id, _ := strconv.Atoi(d.Id())
	apiVPC, err := clientset.VPC().GetByID(id)

	if err != nil {
		return err
	}
	if apiVPC == nil || apiVPC.ID <= 0 {
		return fmt.Errorf("VPC response is missing a valid ID")
	}

	d.SetId(strconv.Itoa(apiVPC.ID))
	err = d.Set("name", apiVPC.Name)
	if err != nil {
		return err
	}
	if err := flattenVNIAndRouteTargets(d, apiVPC, false); err != nil {
		return err
	}
	if err := d.Set("vrfname", apiVPC.VrfName); err != nil {
		return err
	}
	err = d.Set("tenantid", apiVPC.AdminTenant.ID)
	if err != nil {
		return err
	}

	log.Println("[DEBUG] vpcRead apiVPC.GuestTenant", apiVPC.GuestTenant)

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

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	d.Partial(true)
	log.Println("[DEBUG] vpcUpdate")
	clientset := m.(*api.Clientset)

	vpcID, _ := strconv.Atoi(d.Id())

	guestTenantIdsList := d.Get("guesttenantid").(*schema.Set).List()
	log.Println("[DEBUG] guestTenantIdsList", guestTenantIdsList)
	guestTenants := []vpc.GuestTenant{}

	for _, eachGuestTenant := range guestTenantIdsList {
		gTenant := eachGuestTenant.(map[string]interface{})
		guestTenants = append(guestTenants, vpc.GuestTenant{
			ID: gTenant["id"].(int),
		})
	}

	tagsList := d.Get("tags").(*schema.Set).List()
	tags := []string{}
	for _, tag := range tagsList {
		tags = append(tags, tag.(string))
	}

	vpcUpdate := &vpc.VPCw{
		RouteTargets: expandRouteTargets(d),
		Name:         d.Get("name").(string),
		AdminTenant:  vpc.AdminTenant{ID: d.Get("tenantid").(int)},
		GuestTenant:  guestTenants,
		Tags:         tags,
	}

	routes := vpcUpdate.RouteTargets
	var before *vpc.VPC
	if d.HasChange("vni") {
		var err error
		before, err = clientset.VPC().GetByID(vpcID)
		if err != nil {
			return err
		}
		vpcUpdate.VNI = expandVNI(d)
		// The controller may replace default RT IDs when it changes the VNI.
		vpcUpdate.RouteTargets = nil
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

	if _, err := parseWriteReply(reply); err != nil {
		return err
	}
	if d.HasChange("vni") && routes != nil {
		current, err := clientset.VPC().GetByID(vpcID)
		if err != nil {
			return err
		}
		remapRouteTargetIDs(routes, before, current)
		vpcUpdate.VNI = nil
		vpcUpdate.RouteTargets = routes
		reply, err = clientset.VPC().Update(vpcID, vpcUpdate)
		if err != nil {
			return err
		}
		if _, err := parseWriteReply(reply); err != nil {
			return err
		}
	}
	if err := resourceRead(d, m); err != nil {
		return err
	}
	d.Partial(false)
	return nil
}

func resourceExists(d *schema.ResourceData, m interface{}) (bool, error) {
	clientset := m.(*api.Clientset)

	id, _ := strconv.Atoi(d.Id())
	item, err := clientset.VPC().GetByID(id)
	if err != nil {
		log.Println("[DEBUG] vpcExist response err:", err)
		return false, err
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
