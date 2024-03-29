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

package allocation

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/netrisai/netriswebapi/http"
	"github.com/netrisai/netriswebapi/v2/types/ipam"

	api "github.com/netrisai/netriswebapi/v2"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: "Creates and manages Allocations",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Unique name for current allocation.",
			},
			"prefix": {
				Required:    true,
				Type:        schema.TypeString,
				Description: "Unique prefix for allocation, must not overlap with other allocations.",
			},
			"tenantid": {
				ForceNew:    true,
				Required:    true,
				Type:        schema.TypeInt,
				Description: "ID of tenant. Users of this tenant will be permitted to manage subnets under this allocation.",
			},
			"vpcid": {
				ForceNew:    true,
				Optional:    true,
				Type:        schema.TypeInt,
				Description: "ID of VPC. If not specified, the allocation will be created in the VPC marked as a default.",
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
	log.Println("[DEBUG] allocation resourceCreate")
	clientset := m.(*api.Clientset)

	name := d.Get("name").(string)
	prefix := d.Get("prefix").(string)
	tenant := d.Get("tenantid").(int)
	vpcid := d.Get("vpcid").(int)

	allAdd := &ipam.Allocation{
		Name:   name,
		Prefix: prefix,
		Tenant: ipam.IDName{ID: tenant},
	}

	if vpcid > 0 {
		allAdd.Vpc = &ipam.IDName{ID: vpcid}
	}

	js, _ := json.Marshal(allAdd)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.IPAM().AddAllocation(allAdd)
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
	log.Println("[DEBUG] allocation resourceRead")
	clientset := m.(*api.Clientset)
	currentVpcId := d.Get("vpcid").(int)
	var ipams []*ipam.IPAM
	var err error
	if currentVpcId > 0 {
		ipams, err = clientset.IPAM().GetByVPC(currentVpcId)
	} else {
		ipams, err = clientset.IPAM().Get()
	}
	if err != nil {
		return err
	}
	id, _ := strconv.Atoi(d.Id())
	ipam := getByID(ipams, id)
	if ipam == nil {
		return nil
	}

	d.SetId(strconv.Itoa(ipam.ID))
	err = d.Set("name", ipam.Name)
	if err != nil {
		return err
	}
	err = d.Set("prefix", ipam.Prefix)
	if err != nil {
		return err
	}
	err = d.Set("tenantid", ipam.Tenant.ID)
	if err != nil {
		return err
	}
	if currentVpcId > 0 {
		err = d.Set("vpcid", ipam.Vpc.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	log.Println("[DEBUG] allocation resourceUpdate")
	clientset := m.(*api.Clientset)

	name := d.Get("name").(string)
	prefix := d.Get("prefix").(string)
	tenant := d.Get("tenantid").(int)

	allUpdate := &ipam.Allocation{
		Name:   name,
		Prefix: prefix,
		Tenant: ipam.IDName{ID: tenant},
	}

	js, _ := json.Marshal(allUpdate)
	log.Println("[DEBUG]", string(js))
	id, _ := strconv.Atoi(d.Id())
	reply, err := clientset.IPAM().UpdateAllocation(id, allUpdate)
	if err != nil {
		log.Println("[DEBUG]", err)
		return err
	}

	js, _ = json.Marshal(reply)
	log.Println("[DEBUG]", string(js))

	log.Println("[DEBUG]", string(reply.Data))

	if reply.StatusCode != 200 {
		return fmt.Errorf(string(reply.Data))
	}

	return nil
}

func resourceDelete(d *schema.ResourceData, m interface{}) error {
	log.Println("[DEBUG] allocation resourceDelete")
	clientset := m.(*api.Clientset)
	id, _ := strconv.Atoi(d.Id())
	reply, err := clientset.IPAM().Delete("allocation", id)
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
	log.Println("[DEBUG] allocation resourceExists")
	clientset := m.(*api.Clientset)
	currentVpcId := d.Get("vpcid").(int)
	var ipams []*ipam.IPAM
	var err error
	if currentVpcId > 0 {
		ipams, err = clientset.IPAM().GetByVPC(currentVpcId)
	} else {
		ipams, err = clientset.IPAM().Get()
	}
	if err != nil {
		return false, err
	}
	id, _ := strconv.Atoi(d.Id())
	if ipam := getByID(ipams, id); ipam == nil {
		return false, nil
	}

	return true, nil
}

func resourceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	log.Println("[DEBUG] allocation resourceImport")
	clientset := m.(*api.Clientset)

	ipams, err := clientset.IPAM().Get()
	if err != nil {
		return []*schema.ResourceData{d}, err
	}
	prefix := d.Id()
	ipam := getByPrefix(ipams, prefix)
	if ipam == nil {
		return []*schema.ResourceData{d}, fmt.Errorf("allocation '%s' not found", prefix)
	}

	d.SetId(strconv.Itoa(ipam.ID))

	return []*schema.ResourceData{d}, nil
}

func getByPrefix(list []*ipam.IPAM, prefix string) *ipam.IPAM {
	for _, s := range list {
		if s.Prefix == prefix && s.Type == "allocation" {
			return s
		} else if len(s.Children) > 0 {
			if p := getByPrefix(s.Children, prefix); p != nil {
				return p
			}
		}
	}
	return nil
}

func getByID(list []*ipam.IPAM, id int) *ipam.IPAM {
	for _, s := range list {
		if s.ID == id && s.Type == "allocation" {
			return s
		} else if len(s.Children) > 0 {
			if p := getByID(s.Children, id); p != nil {
				return p
			}
		}
	}
	return nil
}
