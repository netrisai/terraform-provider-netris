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
		Description: "Data Source: Switch Ports",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Port's exact name",
			},
			"description": {
				Computed:    true,
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Port desired description",
			},
			"switchid": {
				Computed:    true,
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The switch ID to whom this port belongs",
			},
			"tenantid": {
				Computed:    true,
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "ID of tenant. Users of this tenant will be permitted to manage port",
			},
			"breakout": {
				Computed:    true,
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Toggle breakout.",
			},
			"mtu": {
				Computed:    true,
				Optional:    true,
				Type:        schema.TypeInt,
				Description: "MTU must be integer between 68 and 9216.",
			},
			"autoneg": {
				Computed:    true,
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Toggle auto negotiation.",
			},
			"speed": {
				Computed:    true,
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Toggle interface speed, make sure that current switch supports the configured speed.",
			},
			"extension": {
				Computed:    true,
				Optional:    true,
				Type:        schema.TypeMap,
				Description: "Port extension configurations.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"extensionname": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Name for new extension.",
						},
						"vlanrange": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "VLAN ID range for new extension port.",
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
