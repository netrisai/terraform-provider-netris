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

package route

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/netrisai/netriswebapi/http"
	"github.com/netrisai/netriswebapi/v1/types/route"
	api "github.com/netrisai/netriswebapi/v2"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: "Creates and manages static routes",
		Schema: map[string]*schema.Schema{
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of route",
			},
			"prefix": {
				Required:    true,
				Type:        schema.TypeString,
				Description: "Route destination to match",
			},
			"nexthop": {
				Required:    true,
				Type:        schema.TypeString,
				ForceNew:    true,
				Description: "Traffic destined to the Prefix will be routed towards the Next-Hop. Note that static routes will be injected only on units that have the Next-Hop as a connected network",
			},
			"siteid": {
				Required:    true,
				Type:        schema.TypeInt,
				Description: "The site ID where the current route belongs",
			},
			"vpcid": {
				ForceNew:    true,
				Optional:    true,
				Type:        schema.TypeInt,
				Description: "ID of VPC. If not specified, the route will be created in the VPC marked as a default.",
			},
			"state": {
				Default:     "enabled",
				Optional:    true,
				Type:        schema.TypeString,
				Description: "Administrative state of the route. Possible values: `enabled` or `disabled`. Default value is `enabled`",
			},
			"hwids": {
				Optional:    true,
				Type:        schema.TypeList,
				Description: "Hardware ID where to apply this route. It is typically used for Null routes. If not set, Netris will automatically decide where to apply",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
		},
		Create: resourceCreate,
		Read:   resourceRead,
		Update: resourceUpdate,
		Delete: resourceDelete,
		Exists: resourceExists,
	}
}

func DiffSuppress(k, old, new string, d *schema.ResourceData) bool {
	return true
}

func resourceCreate(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	hwIds := []int{}
	hws := d.Get("hwids").([]interface{})
	vpcid := d.Get("vpcid").(int)

	for _, v := range hws {
		hwIds = append(hwIds, v.(int))
	}

	routeAdd := &route.RouteAdd{
		Description: d.Get("description").(string),
		Prefix:      d.Get("prefix").(string),
		NextHop:     d.Get("nexthop").(string),
		SiteID:      d.Get("siteid").(int),
		StateStatus: d.Get("state").(string),
		Switches:    hwIds,
	}

	if vpcid > 0 {
		routeAdd.Vpc = &route.IDName{ID: vpcid}
	}

	js, _ := json.Marshal(routeAdd)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.Route().Add(routeAdd)
	if err != nil {
		log.Println("[DEBUG]", err)
		return err
	}

	js, _ = json.Marshal(reply)
	log.Println("[DEBUG]", string(js))

	log.Println("[DEBUG]", string(reply.Data))

	idStruct := struct {
		ID int `json:"staticRouteID"`
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
	var route *route.Route
	routes, err := clientset.Route().Get()
	if err != nil {
		return err
	}
	for _, r := range routes {
		if r.ID == id {
			route = r
			break
		}
	}
	currentVpcId := d.Get("vpcid").(int)

	if route == nil {
		return fmt.Errorf("Coudn't find route by id %d", id)
	}

	d.SetId(strconv.Itoa(route.ID))
	err = d.Set("description", route.Description)
	if err != nil {
		return err
	}
	err = d.Set("prefix", fmt.Sprintf("%s/%d", route.Prefix, route.PrefixLength))
	if err != nil {
		return err
	}
	err = d.Set("nexthop", route.NextHop)
	if err != nil {
		return err
	}
	err = d.Set("siteid", route.SiteID)
	if err != nil {
		return err
	}
	err = d.Set("state", route.State)
	if err != nil {
		return err
	}

	if currentVpcId > 0 {
		err = d.Set("vpcid", route.Vpc.ID)
		if err != nil {
			return err
		}
	}

	hwids := []int{}
	for _, s := range route.FilteredSwitches {
		hwids = append(hwids, s.ID)
	}
	err = d.Set("hwids", hwids)
	if err != nil {
		return err
	}

	return nil
}

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	hwIds := []int{}
	hws := d.Get("hwids").([]interface{})
	for _, v := range hws {
		hwIds = append(hwIds, v.(int))
	}

	id, _ := strconv.Atoi(d.Id())
	routeAdd := &route.RouteAdd{
		RouteID:     id,
		Description: d.Get("description").(string),
		Prefix:      d.Get("prefix").(string),
		NextHop:     d.Get("nexthop").(string),
		SiteID:      d.Get("siteid").(int),
		StateStatus: d.Get("state").(string),
		Switches:    hwIds,
	}

	js, _ := json.Marshal(routeAdd)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.Route().Update(routeAdd)
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

func resourceDelete(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	id, _ := strconv.Atoi(d.Id())
	reply, err := clientset.Route().Delete(id)
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
	id, _ := strconv.Atoi(d.Id())
	routes, err := clientset.Route().Get()
	if err != nil {
		return false, err
	}
	for _, r := range routes {
		if r.ID == id {
			return true, nil
		}
	}

	return false, nil
}
