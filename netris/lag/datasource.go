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

package lag

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/netrisai/netriswebapi/v2/types/port"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	api "github.com/netrisai/netriswebapi/v2"
)

func DataResource() *schema.Resource {
	return &schema.Resource{
		Description: "Manages Switch Ports",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				ValidateFunc: validateName,
				Required:     true,
				Description:  "Aggregated port name (agg1@switch1)",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Lag desired description",
			},
			"tenantid": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "ID of tenant. Users of this tenant will be permitted to manage port",
			},
			"mtu": {
				Default:     9000,
				Optional:    true,
				Type:        schema.TypeInt,
				Description: "MTU must be integer between 68 and 9216. Default value is `9000`",
			},
			"lacp": {
				Default:     "off",
				Optional:    true,
				Type:        schema.TypeString,
				Description: "LACP option",
			},
			"autoneg": {
				Default:      "default",
				ValidateFunc: validateAutoneg,
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Toggle auto negotiation. Possible values: `default`, `on`, `off`. Default value is `default`",
			},
			"members": {
				Optional:    true,
				Type:        schema.TypeSet,
				Description: "Member ports",
				Elem: &schema.Schema{
					Type:    schema.TypeString,
					Default: "",
				},
			},
			"extension": {
				Optional:     true,
				Type:         schema.TypeMap,
				Description:  "Port extension configurations.",
				ValidateFunc: validateExtension,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"extensionname": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Name for new extension.",
						},
						"vlanrange": {
							ValidateFunc: validatePort,
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "VLAN ID range for new extension port. Example: `10-15`",
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

	ports, err := clientset.Port().Get()
	if err != nil {
		return err
	}

	var hwPort *port.Port
	splitedName := strings.Split(name, "@")
	aggPort := splitedName[0]
	swName := splitedName[1]

	for _, p := range ports {
		if p.Info.Port == aggPort && p.SwitchName == swName {
			hwPort, err = clientset.Port().GetByID(p.ID)
			if err != nil {
				return err
			}
			break
		}
	}

	d.SetId(strconv.Itoa(hwPort.ID))
	err = d.Set("description", hwPort.Description)
	if err != nil {
		return err
	}

	err = d.Set("lacp", hwPort.Lacp)
	if err != nil {
		return err
	}

	err = d.Set("tenantid", hwPort.Tenant.ID)
	if err != nil {
		return err
	}

	var ext *port.PortLAGExtension
	list, err := clientset.Port().GetExtenstion()
	if err != nil {
		return err
	}
	for _, e := range list {
		if e.ID == hwPort.Extension {
			ext = &port.PortLAGExtension{
				ID:       e.ID,
				Name:     e.Name,
				VlanFrom: e.VlanFrom,
				VlanTo:   e.VlanTo,
			}
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

	members := []string{}

	for _, p := range hwPort.SlavePorts {
		members = append(members, p.Info.Port+"@"+p.SwitchName)
	}

	err = d.Set("members", members)
	if err != nil {
		return err
	}

	return nil
}

func dataResourceExists(d *schema.ResourceData, m interface{}) (bool, error) {
	return true, nil
}
