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

package link

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/netrisai/netriswebapi/http"
	api "github.com/netrisai/netriswebapi/v2"
	"github.com/netrisai/netriswebapi/v2/types/inventory"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"ports": {
				ForceNew: true,
				Required: true,
				Type:     schema.TypeList,
				Elem: &schema.Schema{
					// ValidateFunc: validateIP,
					Type: schema.TypeString,
				},
			},
		},
		Create: resourceCreate,
		Delete: resourceDelete,
		Read:   resourceRead,
		Exists: resourceExists,
	}
}

func DiffSuppress(k, old, new string, d *schema.ResourceData) bool {
	return true
}

func resourceCreate(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	origin := 0
	dest := 0
	hwID := 0

	ports, err := clientset.Port().Get()
	if err != nil {
		return err
	}

	portList := d.Get("ports").([]interface{})
	if o, ok := findPortByName(ports, portList[0].(string), clientset); ok {
		hwID = o.Switch.ID
		origin = o.ID
	}
	if d, ok := findPortByName(ports, portList[1].(string), clientset); ok {
		dest = d.ID
	}

	hw, err := clientset.Inventory().GetByID(hwID)
	if err != nil {
		return err
	}

	hwLink := inventory.HWLink{
		Local:  inventory.IDName{ID: origin},
		Remote: inventory.IDName{ID: dest},
	}

	hw.Links = append(hw.Links, hwLink)

	var reply http.HTTPReply

	if hw.Type == "switch" {
		swUpdate := hwToSwitchUpdate(hw)
		js, _ := json.Marshal(swUpdate)
		log.Println("[DEBUG]", string(js))
		reply, err = clientset.Inventory().UpdateSwitch(hw.ID, swUpdate)
		if err != nil {
			return err
		}
	} else if hw.Type == "softgate" {
		sgUpdate := hwToSoftgateUpdate(hw)
		js, _ := json.Marshal(sgUpdate)
		log.Println("[DEBUG]", string(js))
		reply, err = clientset.Inventory().UpdateSoftgate(hw.ID, sgUpdate)
		if err != nil {
			return err
		}
	}

	js, _ := json.Marshal(reply)
	log.Println("[DEBUG]", string(js))

	log.Println("[DEBUG]", string(reply.Data))

	if reply.StatusCode != 200 {
		return fmt.Errorf(string(reply.Data))
	}

	d.SetId(fmt.Sprintf("%d-%d", origin, dest))
	return nil
}

func resourceRead(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)
	itemid := strings.Split(d.Id(), "-")
	if len(itemid) != 2 {
		return fmt.Errorf("invalid link")
	}

	origin, _ := strconv.Atoi(itemid[0])
	dest, _ := strconv.Atoi(itemid[1])
	hwID := 0

	portList := []interface{}{}

	ports, err := clientset.Port().Get()
	if err != nil {
		return err
	}

	if port, ok := findPortByID(ports, origin, clientset); ok {
		portList = append(portList, fmt.Sprintf("%s@%s", port.Port, port.SwitchName))
		hwID = port.Switch.ID
	} else {
		return fmt.Errorf("invalid link")
	}
	if port, ok := findPortByID(ports, dest, clientset); ok {
		portList = append(portList, fmt.Sprintf("%s@%s", port.Port, port.SwitchName))
	} else {
		return fmt.Errorf("invalid link")
	}

	hw, err := clientset.Inventory().GetByID(hwID)
	if err != nil {
		return err
	}

	linkExist := false

	for _, link := range hw.Links {
		if (link.Local.ID == origin && link.Remote.ID == dest) || (link.Local.ID == dest && link.Remote.ID == origin) {
			linkExist = true
			break
		}
	}

	if linkExist {
		d.SetId(d.Id())
		err = d.Set("ports", portList)
		if err != nil {
			return err
		}
	}

	return nil
}

func resourceDelete(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)
	itemid := strings.Split(d.Id(), "-")
	if len(itemid) != 2 {
		return fmt.Errorf("invalid link")
	}

	origin, _ := strconv.Atoi(itemid[0])
	dest, _ := strconv.Atoi(itemid[1])

	port, err := clientset.Port().GetByID(origin)
	if err != nil {
		return err
	}
	hwID := port.Switch.ID

	hw, err := clientset.Inventory().GetByID(hwID)
	if err != nil {
		return err
	}

	for i, link := range hw.Links {
		fmt.Println(i)
		if (link.Local.ID == origin && link.Remote.ID == dest) || (link.Local.ID == dest && link.Remote.ID == origin) {
			hw.Links[i] = hw.Links[len(hw.Links)-1]
			hw.Links[len(hw.Links)-1] = inventory.HWLink{}
			hw.Links = hw.Links[:len(hw.Links)-1]
			break
		}
	}

	var reply http.HTTPReply

	if hw.Type == "switch" {
		swUpdate := hwToSwitchUpdate(hw)
		js, _ := json.Marshal(swUpdate)
		log.Println("[DEBUG]", string(js))
		reply, err = clientset.Inventory().UpdateSwitch(hw.ID, swUpdate)
		if err != nil {
			return err
		}
	} else if hw.Type == "softgate" {
		sgUpdate := hwToSoftgateUpdate(hw)
		js, _ := json.Marshal(sgUpdate)
		log.Println("[DEBUG]", string(js))
		reply, err = clientset.Inventory().UpdateSoftgate(hw.ID, sgUpdate)
		if err != nil {
			return err
		}
	}

	js, _ := json.Marshal(reply)
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
	itemid := strings.Split(d.Id(), "-")
	if len(itemid) != 2 {
		return false, fmt.Errorf("invalid link")
	}

	origin, _ := strconv.Atoi(itemid[0])
	dest, _ := strconv.Atoi(itemid[1])
	hwID := 0

	ports, err := clientset.Port().Get()
	if err != nil {
		return false, err
	}

	if port, ok := findPortByID(ports, origin, clientset); ok {
		hwID = port.Switch.ID
	} else {
		return false, fmt.Errorf("invalid link")
	}

	hw, err := clientset.Inventory().GetByID(hwID)
	if err != nil {
		return false, err
	}

	for _, link := range hw.Links {
		if (link.Local.ID == origin && link.Remote.ID == dest) || (link.Local.ID == dest && link.Remote.ID == origin) {
			return true, nil
		}
	}

	return false, nil
}
