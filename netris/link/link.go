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

	"github.com/netrisai/netriswebapi/http"
	api "github.com/netrisai/netriswebapi/v2"
	"github.com/netrisai/netriswebapi/v2/types/link"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: "Creates and manages Links",
		Schema: map[string]*schema.Schema{
			"ports": {
				ForceNew: true,
				Required: true,
				Type:     schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "List of two ports.",
			},
			"ipv4": {
				ForceNew: true,
				Optional: true,
				Type:     schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "List of two IPv4 addresses.",
			},
			"ipv6": {
				ForceNew: true,
				Optional: true,
				Type:     schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "List of two IPv6 addresses",
			},
		},
		Create: resourceCreate,
		Delete: resourceDelete,
		Read:   resourceRead,
		Exists: resourceExists,
		// Update: resourceUpdate,
		Importer: &schema.ResourceImporter{
			State: resourceImport,
		},
	}
}

func DiffSuppress(k, old, new string, d *schema.ResourceData) bool {
	return true
}

func resourceCreate(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)
	portList := d.Get("ports").([]interface{})
	if len(portList) != 2 {
		return fmt.Errorf("`ports` should be the list of TWO ports")
	}
	local := portList[0].(string)
	remote := portList[1].(string)

	localIpv4 := ""
	remoteIpv4 := ""
	localIpv6 := ""
	remoteIpv6 := ""

	ipv4List := d.Get("ipv4").([]interface{})
	ipv6List := d.Get("ipv6").([]interface{})

	if len(ipv4List) == 2 {
		localIpv4 = ipv4List[0].(string)
		remoteIpv4 = ipv4List[1].(string)
	}

	if len(ipv6List) == 2 {
		localIpv6 = ipv6List[0].(string)
		remoteIpv6 = ipv6List[1].(string)
	}

	linkAdd := &link.Linkw{
		Local:  link.LinkIDName{Name: local, Ipv4: localIpv4, Ipv6: localIpv6},
		Remote: link.LinkIDName{Name: remote, Ipv4: remoteIpv4, Ipv6: remoteIpv6},
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

	idStruct := struct {
		ID int `json:"id"`
	}{}

	data, err := reply.Parse()
	if err != nil {
		log.Println("[DEBUG]", err)
		return err
	}

	err = http.Decode(data.Data, &idStruct)
	if err != nil {
		log.Println("[DEBUG]", err)
		return err
	}

	log.Println("[DEBUG] ID:", idStruct.ID)

	if reply.StatusCode != 200 {
		return fmt.Errorf(string(reply.Data))
	}

	d.SetId(strconv.Itoa(idStruct.ID))
	return nil
}

func resourceRead(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)
	portList := []interface{}{}

	id, _ := strconv.Atoi(d.Id())
	link, err := clientset.Link().GetByID(id)
	if err != nil {
		return err
	}

	portList = append(portList, link.Local.Name)
	portList = append(portList, link.Remote.Name)

	d.SetId(strconv.Itoa(link.ID))

	err = d.Set("ports", portList)
	if err != nil {
		return err
	}

	return nil
}

func resourceDelete(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)
	id, _ := strconv.Atoi(d.Id())

	reply, err := clientset.Link().DeletByID(id)
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
	id, _ := strconv.Atoi(d.Id())
	link, err := clientset.Link().GetByID(id)
	if err != nil {
		return false, nil
	}

	if link == nil {
		return false, nil
	}
	if link.ID > 0 {
		return true, nil
	}

	return false, nil
}

func resourceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	clientset := m.(*api.Clientset)
	id, _ := strconv.Atoi(d.Id())
	link, err := clientset.Link().GetByID(id)
	if err != nil {
		return nil, err
	}
	d.SetId(strconv.Itoa(link.ID))
	return []*schema.ResourceData{d}, nil
}
