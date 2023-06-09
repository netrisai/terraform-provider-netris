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

package bgpobject

import (
	"fmt"
	"strconv"

	"github.com/netrisai/netriswebapi/v1/types/bgpobject"

	api "github.com/netrisai/netriswebapi/v2"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func DataResource() *schema.Resource {
	return &schema.Resource{
		Description: "Data Source: BGP Objects",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the BGP Object",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "BGP Objects type. Detailed documentation about objects types is available [here](https://www.netris.ai/docs/en/stable/network-policies.html#bgp-objects)",
			},
			"value": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Object value.",
			},
		},
		Read:   dataResourceRead,
		Exists: dataResourceExists,
		Importer: &schema.ResourceImporter{
			State: dataResourceImport,
		},
	}
}

func dataResourceRead(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	obj, ok := findByName(d.Get("name").(string), clientset)
	if !ok {
		return fmt.Errorf("Coudn't find bgp object '%s'", d.Get("name").(string))
	}

	d.SetId(strconv.Itoa(obj.ID))
	err := d.Set("name", obj.Name)
	if err != nil {
		return err
	}
	err = d.Set("type", obj.Type)
	if err != nil {
		return err
	}
	err = d.Set("value", obj.TypeValue)
	if err != nil {
		return err
	}
	return nil
}

func dataResourceExists(d *schema.ResourceData, m interface{}) (bool, error) {
	clientset := m.(*api.Clientset)
	var ok bool
	_, ok = findByName(d.Get("name").(string), clientset)
	if !ok {
		return false, fmt.Errorf("Coudn't find bgp object '%s'", d.Get("name").(string))
	}

	return true, nil
}

func dataResourceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	clientset := m.(*api.Clientset)
	name := d.Id()
	var obj *bgpobject.BGPObject
	var ok bool
	obj, ok = findByName(name, clientset)
	if !ok {
		return []*schema.ResourceData{d}, fmt.Errorf("Coudn't find bgp object '%s'", d.Get("name").(string))
	}
	d.SetId(strconv.Itoa(obj.ID))

	return []*schema.ResourceData{d}, nil
}
