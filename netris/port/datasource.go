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

package port

import (
	"fmt"
	"strconv"

	"github.com/netrisai/netriswebapi/v2/types/port"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	api "github.com/netrisai/netriswebapi/v2"
)

func DataResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Computed: true,
				Type:     schema.TypeString,
				Optional: true,
			},
			"switchid": {
				Computed: true,
				Type:     schema.TypeInt,
				Optional: true,
			},
			"tenantid": {
				Computed: true,
				Type:     schema.TypeInt,
				Optional: true,
			},
			"breakout": {
				Computed: true,
				Type:     schema.TypeString,
				Optional: true,
			},
			"mtu": {
				Computed: true,
				Optional: true,
				Type:     schema.TypeInt,
			},
			"autoneg": {
				Computed: true,
				Type:     schema.TypeString,
				Optional: true,
			},
			"speed": {
				Computed: true,
				Type:     schema.TypeString,
				Optional: true,
			},
			"extension": {
				Computed: true,
				Optional: true,
				Type:     schema.TypeMap,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"extensionname": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"vlanrange": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
		Read:   dataResourceRead,
		Exists: dataResourceExists,
	}
}

func dataResourceRead(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	name := d.Get("name").(string)
	var hwPort *port.Port

	ports, err := clientset.Port().Get()
	if err != nil {
		return err
	}
	for _, p := range ports {
		if fmt.Sprintf("%s@%s", p.Port, p.Switch.Name) == name {
			hwPort = p
			break
		}
	}

	if hwPort == nil {
		return fmt.Errorf("Coudn't find port %s", name)
	}

	d.SetId(strconv.Itoa(hwPort.ID))
	err = d.Set("description", hwPort.Description)
	if err != nil {
		return err
	}
	err = d.Set("switchid", hwPort.Switch.ID)
	if err != nil {
		return err
	}
	err = d.Set("tenantid", hwPort.Tenant.ID)
	if err != nil {
		return err
	}
	err = d.Set("breakout", hwPort.Breakout)
	if err != nil {
		return err
	}
	if hwPort.Breakout == "off" || hwPort.Breakout == "manual" {
		err = d.Set("mtu", hwPort.Mtu)
		if err != nil {
			return err
		}
		autoneg := hwPort.AutoNeg
		if autoneg == "none" {
			autoneg = "default"
		}
		err = d.Set("autoneg", autoneg)
		if err != nil {
			return err
		}
		err = d.Set("speed", speedMapReversed[hwPort.DesiredSpeed])
		if err != nil {
			return err
		}

		var ext *port.PortExtension
		list, err := clientset.Port().GetExtenstion()
		if err != nil {
			return err
		}
		for _, e := range list {
			if e.ID == hwPort.Extension {
				ext = e
			}
		}

		extension := make(map[string]interface{})
		if ext != nil {
			extension["extensionname"] = ext.Name
			extension["vlanrange"] = fmt.Sprintf("%d-%d", ext.VlanFrom, ext.VlanTo)
		}

		err = d.Set("extension", extension)
		if err != nil {
			return err
		}
	}

	return nil
}

func dataResourceExists(d *schema.ResourceData, m interface{}) (bool, error) {
	return true, nil
}
