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
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"switchid": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"tenantid": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"breakout": {
				Default:      "off",
				ValidateFunc: validateBreakout,
				Type:         schema.TypeString,
				Optional:     true,
			},
			"mtu": {
				Default:  9000,
				Optional: true,
				Type:     schema.TypeInt,
			},
			"autoneg": {
				Default:      "default",
				ValidateFunc: validateAutoneg,
				Type:         schema.TypeString,
				Optional:     true,
			},
			"speed": {
				Default:      "auto",
				ValidateFunc: validateSpeed,
				Type:         schema.TypeString,
				Optional:     true,
			},
			"extension": {
				ValidateFunc: validateExtension,
				Optional:     true,
				Type:         schema.TypeMap,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"extensionname": {
							Type:     schema.TypeString,
							Required: true,
						},
						"vlanrange": {
							ValidateFunc: validatePort,
							Type:         schema.TypeString,
							Required:     true,
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
	ports, err := clientset.Port().Get()
	if err != nil {
		return err
	}

	name := d.Get("name").(string)
	description := d.Get("description").(string)
	breakout := d.Get("breakout").(string)
	switchID := d.Get("switchid").(int)
	tenantID := d.Get("tenantid").(int)

	for _, p := range ports {
		if p.Port == name && p.Tenant.ID == tenantID && p.Switch.ID == switchID {
			hwPort = p
			break
		}
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
		if autoneg == "default" {
			autoneg = "none"
		}
		speed := d.Get("speed").(string)
		extension := port.PortUpdateExtenstion{}
		ext := d.Get("extension").(map[string]interface{})
		if v, ok := ext["vlanrange"]; ok {
			vlanrange := strings.Split(v.(string), "-")
			from, _ := strconv.Atoi(vlanrange[0])
			to, _ := strconv.Atoi(vlanrange[1])
			extension.Name = ext["extensionname"].(string)
			extension.VLANFrom = from
			extension.VLANTo = to
			if e, ok := findExtensionByName(extension.Name, clientset); ok {
				extension.ID = e.ID
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
		return err
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

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	name := d.Get("name").(string)
	description := d.Get("description").(string)
	breakout := d.Get("breakout").(string)
	tenantID := d.Get("tenantid").(int)

	id, _ := strconv.Atoi(d.Id())
	hwPort, err := clientset.Port().GetByID(id)
	if err != nil {
		return err
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
		if v, ok := ext["vlanrange"]; ok {
			vlanrange := strings.Split(v.(string), "-")
			from, _ := strconv.Atoi(vlanrange[0])
			to, _ := strconv.Atoi(vlanrange[1])
			extension.Name = ext["extensionname"].(string)
			extension.VLANFrom = from
			extension.VLANTo = to
			if e, ok := findExtensionByName(extension.Name, clientset); ok {
				extension.ID = e.ID
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
		return false, err
	}

	if port == nil {
		return false, nil
	}

	return true, nil
}
