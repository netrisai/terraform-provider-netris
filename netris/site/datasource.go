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

package site

import (
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	api "github.com/netrisai/netriswebapi/v2"
	"github.com/netrisai/netriswebapi/v2/types/site"
)

func DataResource() *schema.Resource {
	return &schema.Resource{
		Description: "Data Source: Sites",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the site",
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
	var site *site.Site
	sites, err := clientset.Site().Get()
	if err != nil {
		return err
	}

	for _, s := range sites {
		if s.Name == d.Get("name").(string) {
			site = s
			break
		}
	}

	if site == nil {
		return nil
	}

	d.SetId(strconv.Itoa(site.ID))
	err = d.Set("name", site.Name)
	if err != nil {
		return err
	}

	return nil
}

func dataResourceExists(d *schema.ResourceData, m interface{}) (bool, error) {
	clientset := m.(*api.Clientset)
	name := d.Get("name").(string)

	sites, err := clientset.Site().Get()
	if err != nil {
		return false, err
	}

	for _, site := range sites {
		if name == site.Name {
			return true, nil
		}
	}

	return false, nil
}
