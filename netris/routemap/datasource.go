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

package routemap

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	api "github.com/netrisai/netriswebapi/v2"
)

func DataResource() *schema.Resource {
	return &schema.Resource{
		Description: "Creates and manages BGP Route-maps",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of route-map",
			},
		},

		Read:   dataResourceRead,
		Exists: dataResourceExists,
		Importer: &schema.ResourceImporter{
			State: resourceImport,
		},
	}
}

func dataResourceRead(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	id, _ := strconv.Atoi(d.Id())
	obj, ok := findByID(id, clientset)
	if !ok {
		return fmt.Errorf("Coudn't find routemap '%s'", d.Get("name").(string))
	}

	d.SetId(strconv.Itoa(obj.ID))
	err := d.Set("name", obj.Name)
	if err != nil {
		return err
	}

	return nil
}

func dataResourceExists(d *schema.ResourceData, m interface{}) (bool, error) {
	clientset := m.(*api.Clientset)
	id, _ := strconv.Atoi(d.Id())
	_, ok := findByID(id, clientset)
	return ok, nil
}
