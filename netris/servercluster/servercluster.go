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
	"sort"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
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
			"dynamicvpcgroupid": {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ValidateFunc:  validation.IntAtLeast(1),
				ConflictsWith: []string{"vpcid"},
				Description:   "Dynamic VPC Group for untagged L3VPN template objects. Conflicts with vpcid and configured vpcmapping.",
			},
			"vpcid": {
				ForceNew:    true,
				Computed:    true,
				Optional:    true,
				Type:        schema.TypeInt,
				Description: "ID of VPC. If not specified, a new VPC will be created.",
			},
			"vpcmapping": {
				ForceNew:    true,
				Computed:    true,
				Optional:    true,
				Type:        schema.TypeSet,
				Description: "VPC mappings for server cluster template objects. Each VPC-backed template object postfix must be mapped exactly once when this block is used. Set id to an existing VPC ID, or set id to 0 to create a new VPC. Use group to place multiple template objects into the same new VPC.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							ForceNew:    true,
							Computed:    true,
							Optional:    true,
							Type:        schema.TypeInt,
							Description: "VPC ID. Use an existing VPC ID, or 0 to create a new VPC.",
						},
						"dynamicvpcgroupid": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntAtLeast(1),
							Description:  "Dynamic VPC Group for this untagged L3VPN template object. Conflicts with an existing VPC ID, group or name.",
						},
						"group": {
							ForceNew:    true,
							Optional:    true,
							Type:        schema.TypeString,
							Description: "Grouping key for new VPC mappings. Template objects with the same group are created in the same new VPC.",
						},
						"name": {
							ForceNew:    true,
							Optional:    true,
							Type:        schema.TypeString,
							Description: "VPC name. For new VPC mappings, this can override the generated VPC name.",
						},
						"postfix": {
							ForceNew:    true,
							Required:    true,
							Type:        schema.TypeString,
							Description: "Server cluster template object postfix.",
						},
					},
				},
			},
			"templateid": {
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
			"servers": {
				Required: true,
				Type:     schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"vlan": {
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Type:        schema.TypeSet,
				Description: "Specifies the VLAN ID to use for a Server Cluster Template VNet whose VLAN mode is set to \"specify\". One block per VNet postfix.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"postfix": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Postfix of the VNet, as defined in the Server Cluster Template.",
						},
						"vlan_id": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "VLAN ID to assign to this VNet.",
						},
					},
				},
			},
		},
		CustomizeDiff: func(d *schema.ResourceDiff, m interface{}) error {
			// A read can derive the top-level group from per-object mappings.
			// Validate this conflict on creation, when computed values are absent.
			if d.Id() == "" && d.NewValueKnown("dynamicvpcgroupid") && d.Get("dynamicvpcgroupid").(int) > 0 && d.Get("vpcmapping").(*schema.Set).Len() > 0 {
				return fmt.Errorf("dynamicvpcgroupid conflicts with configured vpcmapping")
			}
			return validateVPCMappings(d.Get("vpcmapping").(*schema.Set).List())
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

func expandVLANs(d *schema.ResourceData) []servercluster.VNetVLAN {
	vlanList := d.Get("vlan").(*schema.Set).List()
	vlans := make([]servercluster.VNetVLAN, 0, len(vlanList))
	for _, v := range vlanList {
		vlanMap := v.(map[string]interface{})
		vlans = append(vlans, servercluster.VNetVLAN{
			Postfix: vlanMap["postfix"].(string),
			VLANID:  vlanMap["vlan_id"].(int),
		})
	}
	return vlans
}

func validateVPCMappings(mappings []interface{}) error {
	for _, item := range mappings {
		mapping := item.(map[string]interface{})
		if mapping["dynamicvpcgroupid"].(int) > 0 && (mapping["id"].(int) > 0 || mapping["group"].(string) != "" || mapping["name"].(string) != "") {
			return fmt.Errorf("vpcmapping %q: dynamicvpcgroupid conflicts with id, group and name", mapping["postfix"])
		}
	}
	return nil
}

func expandVPCMappings(mappingsList []interface{}) []servercluster.VPCMapping {
	mappings := []servercluster.VPCMapping{}
	for _, item := range mappingsList {
		mappingMap := item.(map[string]interface{})
		mapping := servercluster.VPCMapping{
			ID:      mappingMap["id"].(int),
			Postfix: mappingMap["postfix"].(string),
		}

		if group, ok := mappingMap["group"].(string); ok {
			mapping.Group = group
		}
		if name, ok := mappingMap["name"].(string); ok {
			mapping.Name = name
		}

		if id := mappingMap["dynamicvpcgroupid"].(int); id > 0 {
			mapping.DynamicVPCGroup = &servercluster.DynamicVPCGroupRef{ID: id}
		}
		mappings = append(mappings, mapping)
	}

	sort.Slice(mappings, func(i, j int) bool {
		return mappings[i].Postfix < mappings[j].Postfix
	})

	return mappings
}

func flattenVPCMappings(mappings []servercluster.VPCMapping) []map[string]interface{} {
	flattenedMappings := []map[string]interface{}{}
	for _, mapping := range mappings {
		flattenedMapping := map[string]interface{}{
			"id":      mapping.ID,
			"postfix": mapping.Postfix,
		}

		if mapping.Group != "" {
			flattenedMapping["group"] = mapping.Group
		}
		if mapping.Name != "" {
			flattenedMapping["name"] = mapping.Name
		}

		if mapping.DynamicVPCGroup != nil {
			flattenedMapping["dynamicvpcgroupid"] = mapping.DynamicVPCGroup.ID
		}
		flattenedMappings = append(flattenedMappings, flattenedMapping)
	}

	sort.Slice(flattenedMappings, func(i, j int) bool {
		return flattenedMappings[i]["postfix"].(string) < flattenedMappings[j]["postfix"].(string)
	})

	return flattenedMappings
}

func resourceCreate(d *schema.ResourceData, m interface{}) error {
	log.Println("[DEBUG] serverclusterCreate")
	clientset := m.(*api.Clientset)

	tagsList := d.Get("tags").(*schema.Set).List()
	tags := []string{}
	for _, tag := range tagsList {
		tags = append(tags, tag.(string))
	}

	serversList := d.Get("servers").(*schema.Set).List()
	servers := []int{}
	for _, server := range serversList {
		servers = append(servers, server.(int))
	}
	if err := validateVPCMappings(d.Get("vpcmapping").(*schema.Set).List()); err != nil {
		return err
	}
	if d.Get("dynamicvpcgroupid").(int) > 0 && (d.Get("vpcid").(int) > 0 || d.Get("vpcmapping").(*schema.Set).Len() > 0) {
		return fmt.Errorf("dynamicvpcgroupid conflicts with vpcid and configured vpcmapping")
	}
	vpcMappings := expandVPCMappings(d.Get("vpcmapping").(*schema.Set).List())

	sort.Ints(servers)
	sortedServers := []servercluster.Servers{}

	for _, sortedServer := range servers {
		sortedServers = append(sortedServers, servercluster.Servers{ID: sortedServer})
	}

	serverclusterAdd := &servercluster.ServerClusterW{
		Name:                    d.Get("name").(string),
		Admin:                   servercluster.IDName{ID: d.Get("adminid").(int)},
		Site:                    servercluster.IDName{ID: d.Get("siteid").(int)},
		VPC:                     servercluster.IDName{ID: d.Get("vpcid").(int)},
		VPCList:                 vpcMappings,
		SrvClusterTemplate:      servercluster.IDName{ID: d.Get("templateid").(int)},
		SrvClusterTemplateVLANs: expandVLANs(d),
		Tags:                    tags,
		Servers:                 sortedServers,
	}

	if id := d.Get("dynamicvpcgroupid").(int); id > 0 {
		serverclusterAdd.DynamicVPCGroup = &servercluster.DynamicVPCGroupRef{ID: id}
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
		return fmt.Errorf("%s", reply.Data)
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
	err = d.Set("adminid", apiServerCluster.Admin.ID)
	if err != nil {
		return err
	}
	err = d.Set("siteid", apiServerCluster.Site.ID)
	if err != nil {
		return err
	}
	groupID := 0
	if apiServerCluster.DynamicVPCGroup != nil {
		groupID = apiServerCluster.DynamicVPCGroup.ID
	}
	if err := d.Set("dynamicvpcgroupid", groupID); err != nil {
		return err
	}
	err = d.Set("vpcid", apiServerCluster.VPC.ID)
	if err != nil {
		return err
	}
	if d.Get("vpcmapping").(*schema.Set).Len() == 0 && len(apiServerCluster.VPCList) > 0 {
		err = d.Set("vpcmapping", flattenVPCMappings(apiServerCluster.VPCList))
		if err != nil {
			return err
		}
	}
	err = d.Set("templateid", apiServerCluster.SrvClusterTemplate.ID)
	if err != nil {
		return err
	}
	err = d.Set("tags", apiServerCluster.Tags)
	if err != nil {
		return err
	}

	serversList := apiServerCluster.Servers
	servers := []int{}
	for _, server := range serversList {
		servers = append(servers, server.ID)
	}
	sort.Ints(servers)

	err = d.Set("servers", servers)
	if err != nil {
		return err
	}

	if d.Get("vlan").(*schema.Set).Len() == 0 && len(apiServerCluster.SrvClusterTemplate.VLANs) > 0 {
		vlans := []interface{}{}
		for _, vlan := range apiServerCluster.SrvClusterTemplate.VLANs {
			vlans = append(vlans, map[string]interface{}{
				"postfix": vlan.Postfix,
				"vlan_id": vlan.VLANID,
			})
		}
		err = d.Set("vlan", vlans)
		if err != nil {
			return err
		}
	}

	return nil
}

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	log.Println("[DEBUG] serverclusterUpdate")
	clientset := m.(*api.Clientset)

	serverclusterID, _ := strconv.Atoi(d.Id())

	tagsList := d.Get("tags").(*schema.Set).List()
	tags := []string{}
	for _, tag := range tagsList {
		tags = append(tags, tag.(string))
	}

	serversList := d.Get("servers").(*schema.Set).List()
	servers := []int{}
	for _, server := range serversList {
		servers = append(servers, server.(int))
	}

	sort.Ints(servers)
	sortedServers := []servercluster.Servers{}

	for _, sortedServer := range servers {
		sortedServers = append(sortedServers, servercluster.Servers{ID: sortedServer})
	}

	serverclusterUpdate := &servercluster.ServerClusterU{
		Tags:    tags,
		Servers: sortedServers,
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
		return fmt.Errorf("%s", reply.Data)
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
		return fmt.Errorf("%s", reply.Data)
	}

	d.SetId("")
	return nil
}
