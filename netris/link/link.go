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

	local := 0
	remote := 0

	ports, err := clientset.Port().Get()
	if err != nil {
		return err
	}

	portList := d.Get("ports").([]interface{})
	if o, ok := findPortByName(ports, portList[0].(string), clientset); ok {
		local = o.ID
	} else {
		return fmt.Errorf("Couldn't find port %s", portList[0])
	}
	if d, ok := findPortByName(ports, portList[1].(string), clientset); ok {
		remote = d.ID
	} else {
		return fmt.Errorf("Couldn't find port %s", portList[1])
	}

	linkAdd := &link.Link{
		Local:  link.LinkIDName{ID: local},
		Remote: link.LinkIDName{ID: remote},
	}

	js, _ := json.Marshal(linkAdd)
	log.Println("[DEBUG] linkAdd", string(js))

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

	d.SetId(fmt.Sprintf("%d-%d", local, remote))
	return nil
}

func resourceRead(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)
	itemid := strings.Split(d.Id(), "-")
	if len(itemid) != 2 {
		return fmt.Errorf("invalid link")
	}

	local, _ := strconv.Atoi(itemid[0])
	remote, _ := strconv.Atoi(itemid[1])
	localName := ""
	remoteName := ""

	portList := []interface{}{}

	links, err := clientset.Link().Get()
	if err != nil {
		return err
	}

	ports, err := clientset.Port().Get()
	if err != nil {
		return err
	}

	if o, ok := findPortByID(ports, local, clientset); ok {
		localName = fmt.Sprintf("%s@%s", o.Port_, o.SwitchName)
	}
	if o, ok := findPortByID(ports, remote, clientset); ok {
		remoteName = fmt.Sprintf("%s@%s", o.Port_, o.SwitchName)
	}

	found := false
	for _, link := range links {
		if link.Local.ID == local && link.Remote.ID == remote {
			portList = append(portList, localName)
			portList = append(portList, remoteName)
			found = true
			break
		} else if link.Local.ID == remote && link.Remote.ID == local {
			portList = append(portList, remoteName)
			portList = append(portList, localName)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("Link not found")
	}

	d.SetId(d.Id())
	err = d.Set("ports", portList)
	if err != nil {
		return err
	}

	return nil
}

func resourceDelete(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)
	itemid := strings.Split(d.Id(), "-")
	if len(itemid) != 2 {
		return fmt.Errorf("invalid link")
	}

	local, _ := strconv.Atoi(itemid[0])
	remote, _ := strconv.Atoi(itemid[1])

	linkDelete := &link.Link{
		Local:  link.LinkIDName{ID: local},
		Remote: link.LinkIDName{ID: remote},
	}

	reply, err := clientset.Link().Delete(linkDelete)
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
	itemid := strings.Split(d.Id(), "-")
	if len(itemid) != 2 {
		return false, fmt.Errorf("invalid link")
	}

	local, _ := strconv.Atoi(itemid[0])
	remote, _ := strconv.Atoi(itemid[1])

	links, err := clientset.Link().Get()
	if err != nil {
		return false, err
	}

	found := false
	for _, link := range links {
		if (link.Local.ID == local && link.Remote.ID == remote) || (link.Local.ID == remote && link.Remote.ID == local) {
			found = true
		}
	}

	if !found {
		return false, nil
	}
	return true, nil
}
