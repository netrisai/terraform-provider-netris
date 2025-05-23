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

package subnet

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
		Description: "Creates and manages Subnet",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Unique name for current subnet.",
			},
			"prefix": {
				ForceNew:    true,
				Required:    true,
				Type:        schema.TypeString,
				Description: "Unique prefix for subnet, must not overlap with other subnets.",
			},
			"tenantid": {
				Required:    true,
				Type:        schema.TypeInt,
				Description: "ID of tenant. Users of this tenant will be permitted to manage the subnet.",
			},
			"vpcid": {
				ForceNew:    true,
				Optional:    true,
				Type:        schema.TypeInt,
				Description: "ID of VPC. If not specified, the subnet will be created in the VPC marked as a default.",
			},
			"purpose": {
				Required:    true,
				Type:        schema.TypeString,
				Description: "Describes which kind of service will be able to use this subnet. Possible values: `common`, `loopback`, `management`, `load-balancer`, `nat`, `inactive`",
			},
			"defaultgateway": {
				Optional:    true,
				Type:        schema.TypeString,
				Description: "Use when purpose is set to `management`.",
			},
			"siteids": {
				Optional:    true,
				Type:        schema.TypeList,
				Description: "List of sites IDs where this subnet is available.",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"globalrouting": {
				Optional:    true,
				Type:        schema.TypeBool,
				Description: "Subnets with `Global Routing` enabled will be advertised from guest VPCs to the System VPC, and if the System VPC has upstream (Internet) connection such subnets will be advertised further upstream.",
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

	name := d.Get("name").(string)
	prefix := d.Get("prefix").(string)
	purpose := d.Get("purpose").(string)
	defaultgw := ""
	sitesList := d.Get("siteids").([]interface{})
	vpcid := d.Get("vpcid").(int)
	sites := []ipam.IDName{}
	for _, s := range sitesList {
		sites = append(sites, ipam.IDName{ID: s.(int)})
	}
	globalRouting := d.Get("globalrouting").(bool)

	if purpose == "management" {
		defaultgw = d.Get("defaultgateway").(string)
	}

	subnetAdd := &ipam.Subnet{
		Name:           name,
		Prefix:         prefix,
		Tenant:         ipam.IDName{ID: d.Get("tenantid").(int)},
		Purpose:        purpose,
		Sites:          sites,
		DefaultGateway: defaultgw,
		Tags:           []string{},
		GlobalRouting:  &globalRouting,
	}

	if vpcid > 0 {
		subnetAdd.Vpc = &ipam.IDName{ID: vpcid}
	}

	js, _ := json.Marshal(subnetAdd)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.IPAM().AddSubnet(subnetAdd)
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

	currentVpcId := d.Get("vpcid").(int)
	var ipams []*ipam.IPAM
	var err error
	if currentVpcId > 0 {
		ipams, err = clientset.IPAM().GetSubnetsByVPC(currentVpcId)
	} else {
		ipams, err = clientset.IPAM().GetSubnets()
	}

	if err != nil {
		return err
	}
	id, _ := strconv.Atoi(d.Id())
	ipam := GetByID(ipams, id)
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
	err = d.Set("purpose", ipam.Purpose)
	if err != nil {
		return err
	}
	err = d.Set("defaultgateway", ipam.DefaultGateway)
	if err != nil {
		return err
	}
	sites := []int{}
	for _, s := range ipam.Sites {
		sites = append(sites, s.ID)
	}
	err = d.Set("siteids", sites)
	if err != nil {
		return err
	}
	if currentVpcId > 0 {
		err = d.Set("vpcid", ipam.Vpc.ID)
		if err != nil {
			return err
		}

		err = d.Set("globalrouting", ipam.GlobalRouting)
		if err != nil {
			return err
		}

	}
	return nil
}

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	name := d.Get("name").(string)
	prefix := d.Get("prefix").(string)
	purpose := d.Get("purpose").(string)
	defaultgw := ""
	sitesList := d.Get("siteids").([]interface{})
	sites := []ipam.IDName{}
	for _, s := range sitesList {
		sites = append(sites, ipam.IDName{ID: s.(int)})
	}

	globalRouting := d.Get("globalrouting").(bool)

	if purpose == "management" {
		defaultgw = d.Get("defaultgateway").(string)
	}

	subnetUpdate := &ipam.Subnet{
		Name:           name,
		Prefix:         prefix,
		Tenant:         ipam.IDName{ID: d.Get("tenantid").(int)},
		Purpose:        purpose,
		Sites:          sites,
		DefaultGateway: defaultgw,
		Tags:           []string{},
		GlobalRouting:  &globalRouting,
	}

	js, _ := json.Marshal(subnetUpdate)
	log.Println("[DEBUG]", string(js))

	id, _ := strconv.Atoi(d.Id())
	reply, err := clientset.IPAM().UpdateSubnet(id, subnetUpdate)
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
	clientset := m.(*api.Clientset)

	id, _ := strconv.Atoi(d.Id())
	reply, err := clientset.IPAM().Delete("subnet", id)
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
	currentVpcId := d.Get("vpcid").(int)
	var ipams []*ipam.IPAM
	var err error
	if currentVpcId > 0 {
		ipams, err = clientset.IPAM().GetSubnetsByVPC(currentVpcId)
	} else {
		ipams, err = clientset.IPAM().GetSubnets()
	}

	if err != nil {
		return false, err
	}
	id, _ := strconv.Atoi(d.Id())
	if ipam := GetByID(ipams, id); ipam == nil {
		return false, nil
	}

	return true, nil
}

func resourceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	clientset := m.(*api.Clientset)

	ipams, err := clientset.IPAM().GetSubnets()
	if err != nil {
		return []*schema.ResourceData{d}, err
	}
	prefix := d.Id()
	ipam := GetByPrefix(ipams, prefix)
	if ipam == nil {
		return []*schema.ResourceData{d}, fmt.Errorf("allocation '%s' not found", prefix)
	}

	d.SetId(strconv.Itoa(ipam.ID))

	return []*schema.ResourceData{d}, nil
}

func GetByPrefix(list []*ipam.IPAM, prefix string) *ipam.IPAM {
	for _, s := range list {
		if s.Prefix == prefix && s.Type == "subnet" {
			return s
		} else if len(s.Children) > 0 {
			if p := GetByPrefix(s.Children, prefix); p != nil {
				return p
			}
		}
	}
	return nil
}

func GetByID(list []*ipam.IPAM, id int) *ipam.IPAM {
	for _, s := range list {
		if s.ID == id && s.Type == "subnet" {
			return s
		} else if len(s.Children) > 0 {
			if p := GetByID(s.Children, id); p != nil {
				return p
			}
		}
	}
	return nil
}
