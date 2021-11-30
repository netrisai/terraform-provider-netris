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

	api "github.com/netrisai/netriswebapi/v2"
	"github.com/netrisai/netriswebapi/v2/types/link"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"itemid": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: DiffSuppress,
			},
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

	ports, err := clientset.Port().Get()
	if err != nil {
		return err
	}

	portList := d.Get("ports").([]interface{})
	if o, ok := findPortByName(ports, portList[0].(string), clientset); ok {
		origin = o.ID
	}
	if d, ok := findPortByName(ports, portList[1].(string), clientset); ok {
		dest = d.ID
	}

	linkAdd := &link.Link{
		Origin: origin,
		Dest:   dest,
	}

	js, _ := json.Marshal(linkAdd)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.Link().Add(linkAdd)
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

	_ = d.Set("itemid", fmt.Sprintf("%d-%d", origin, dest))
	d.SetId(fmt.Sprintf("%d-%d", origin, dest))
	return nil
}

func resourceRead(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)
	itemid := strings.Split(d.Get("itemid").(string), "-")
	if len(itemid) != 2 {
		return fmt.Errorf("invalid link")
	}

	origin, _ := strconv.Atoi(itemid[0])
	dest, _ := strconv.Atoi(itemid[1])

	portList := []interface{}{}

	ports, err := clientset.Port().Get()
	if err != nil {
		return err
	}

	if port, ok := findPortByID(ports, origin, clientset); ok {
		portList = append(portList, fmt.Sprintf("%s@%s", port.Port, port.SwitchName))
	} else {
		return fmt.Errorf("invalid link")
	}
	if port, ok := findPortByID(ports, dest, clientset); ok {
		portList = append(portList, fmt.Sprintf("%s@%s", port.Port, port.SwitchName))
	} else {
		return fmt.Errorf("invalid link")
	}

	d.SetId(d.Get("itemid").(string))
	err = d.Set("ports", portList)
	if err != nil {
		return err
	}

	return nil
}

func resourceDelete(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)
	itemid := strings.Split(d.Get("itemid").(string), "-")
	if len(itemid) != 2 {
		return fmt.Errorf("invalid link")
	}

	origin, _ := strconv.Atoi(itemid[0])
	dest, _ := strconv.Atoi(itemid[1])

	linkW := &link.Link{
		Origin: origin,
		Dest:   dest,
	}
	reply, err := clientset.Link().Delete(linkW)
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

	itemid := strings.Split(d.Get("itemid").(string), "-")
	if len(itemid) != 2 {
		return false, nil
	}

	origin, _ := strconv.Atoi(itemid[0])
	dest, _ := strconv.Atoi(itemid[1])

	ports, err := clientset.Port().Get()
	if err != nil {
		return false, err
	}

	if _, ok := findPortByID(ports, origin, clientset); !ok {
		return false, nil
	}
	if _, ok := findPortByID(ports, dest, clientset); !ok {
		return false, nil
	}

	return true, nil
}
