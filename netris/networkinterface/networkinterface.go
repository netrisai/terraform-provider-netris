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

package networkinterface

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/netrisai/netriswebapi/v2/types/port"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	api "github.com/netrisai/netriswebapi/v2"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: "Manages Network Interfaces",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Network Interface's exact name",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Network Interface desired description",
			},
			"nodeid": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The node ID to whom this network interface belongs",
			},
			"tenantid": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "ID of tenant. Users of this tenant will be permitted to manage network interface",
			},
			"breakout": {
				Default:      "off",
				ValidateFunc: validateBreakout,
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Toggle breakout. Possible values: `off`, `4x10`, `4x25`, `4x100`, `manual`. Default value is `off`",
			},
			"mtu": {
				Default:     9000,
				Optional:    true,
				Type:        schema.TypeInt,
				Description: "MTU must be integer between 68 and 9216. Default value is `9000`",
			},
			"autoneg": {
				Default:      "default",
				ValidateFunc: validateAutoneg,
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Toggle auto negotiation. Possible values: `default`, `on`, `off`. Default value is `default`",
			},
			"speed": {
				Default:      "auto",
				ValidateFunc: validateSpeed,
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Toggle interface speed, make sure that current node supports the configured speed. Possibe values: `auto`, `1g`, `10g`, `25g`, `40g`, `50g`, `100g`, `200g`, `400g`. Default value is `auto`",
			},
			"extension": {
				Optional:     true,
				Type:         schema.TypeMap,
				Description:  "Network Interface extension configurations.",
				ValidateFunc: validateExtension,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"extensionname": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name for new extension.",
						},
						"vlanrange": {
							ValidateFunc: validatePort,
							Type:         schema.TypeString,
							Required:     true,
							Description:  "VLAN ID range for new extension network interface. Example: `10-15`",
						},
					},
				},
			},
		},
		Create: resourceCreate,
		Read:   resourceRead,
		Update: resourceUpdate,
		Delete: resourceDelete,
		Exists: resourceExists,
		// Importer: &schema.ResourceImporter{
		// 	State: resourceImport,
		// },
	}
}

func DiffSuppress(k, old, new string, d *schema.ResourceData) bool {
	return true
}

func resourceCreate(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	var hwPort *port.Port
	switchID := d.Get("nodeid").(int)

	ports, err := clientset.Port().GetBySwId(switchID)
	if err != nil {
		return err
	}

	name := d.Get("name").(string)
	description := d.Get("description").(string)
	breakout := d.Get("breakout").(string)
	tenantID := d.Get("tenantid").(int)

	for _, p := range ports {
		if p.Port == name && p.Switch.ID == switchID {
			hwPort = p
			break
		}
	}

	if hwPort == nil {
		return fmt.Errorf("Coudn't find network interface '%s'", name)
	}

	portUpdate := portDefault()
	portUpdate.Description = description
	portUpdate.Tenant = port.IDName{ID: tenantID}
	portUpdate.Breakout = breakout
	if breakout == "off" || breakout == "manual" {
		mtu := d.Get("mtu").(int)
		autoneg := d.Get("autoneg").(string)
		if autoneg == "default" {
			autoneg = "none"
		}
		if autoneg == "default" {
			autoneg = "none"
		}
		speed := d.Get("speed").(string)

		extension := port.PortUpdateExtenstion{}
		ext := d.Get("extension").(map[string]interface{})
		if n, ok := ext["extensionname"]; ok {
			extensionName := n.(string)
			if e, ok := findExtensionByName(extensionName, clientset); ok {
				extension.ID = e.ID
			} else if v, ok := ext["vlanrange"]; ok {
				vlanrange := strings.Split(v.(string), "-")
				from, _ := strconv.Atoi(vlanrange[0])
				to, _ := strconv.Atoi(vlanrange[1])
				extension.VLANFrom = from
				extension.VLANTo = to
				extension.Name = extensionName
			} else {
				return fmt.Errorf("Please provide vlan range for extension \"%s\"", extensionName)
			}
		}

		portUpdate.Mtu = mtu
		portUpdate.AutoNeg = autoneg
		portUpdate.Speed = speedMap[speed]
		portUpdate.Extension = extension
	}

	js, _ := json.Marshal(portUpdate)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.Port().Update(hwPort.ID, portUpdate)
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

	d.SetId(strconv.Itoa(hwPort.ID))
	return nil
}

func resourceRead(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	id, _ := strconv.Atoi(d.Id())
	hwPort, err := clientset.Port().GetByID(id)
	if err != nil {
		return nil
	}

	d.SetId(strconv.Itoa(hwPort.ID))
	err = d.Set("description", hwPort.Description)
	if err != nil {
		return err
	}
	err = d.Set("nodeid", hwPort.Switch.ID)
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

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	name := d.Get("name").(string)
	description := d.Get("description").(string)
	breakout := d.Get("breakout").(string)
	tenantID := d.Get("tenantid").(int)

	id, _ := strconv.Atoi(d.Id())
	hwPort, err := clientset.Port().GetByID(id)
	if err != nil {
		return nil
	}

	if hwPort == nil {
		return fmt.Errorf("Coudn't find port '%s'", name)
	}

	portUpdate := portDefault()
	portUpdate.Description = description
	portUpdate.Tenant = port.IDName{ID: tenantID}
	portUpdate.Breakout = breakout
	if breakout == "off" || breakout == "manual" {
		mtu := d.Get("mtu").(int)
		autoneg := d.Get("autoneg").(string)
		if autoneg == "default" {
			autoneg = "none"
		}
		speed := d.Get("speed").(string)

		extension := port.PortUpdateExtenstion{}
		ext := d.Get("extension").(map[string]interface{})
		if n, ok := ext["extensionname"]; ok {
			extensionName := n.(string)
			if e, ok := findExtensionByName(extensionName, clientset); ok {
				extension.ID = e.ID
			} else if v, ok := ext["vlanrange"]; ok {
				vlanrange := strings.Split(v.(string), "-")
				from, _ := strconv.Atoi(vlanrange[0])
				to, _ := strconv.Atoi(vlanrange[1])
				extension.VLANFrom = from
				extension.VLANTo = to
				extension.Name = extensionName
			} else {
				return fmt.Errorf("Please provide vlan range for extension \"%s\"", extensionName)
			}
		}

		portUpdate.Mtu = mtu
		portUpdate.AutoNeg = autoneg
		portUpdate.Speed = speedMap[speed]
		portUpdate.Extension = extension
	}

	js, _ := json.Marshal(portUpdate)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.Port().Update(hwPort.ID, portUpdate)
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

	d.SetId(strconv.Itoa(hwPort.ID))
	return nil
}

func resourceDelete(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	name := d.Get("name").(string)
	tenantID := d.Get("tenantid").(int)

	id, _ := strconv.Atoi(d.Id())

	portUpdate := portDefault()
	portUpdate.Description = name
	portUpdate.Tenant = port.IDName{ID: tenantID}

	js, _ := json.Marshal(portUpdate)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.Port().Update(id, portUpdate)
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

	d.SetId("")
	return nil
}

func resourceExists(d *schema.ResourceData, m interface{}) (bool, error) {
	clientset := m.(*api.Clientset)
	id, _ := strconv.Atoi(d.Id())
	port, err := clientset.Port().GetByID(id)
	if err != nil {
		return false, nil
	}

	if port == nil {
		return false, nil
	}

	return true, nil
}
