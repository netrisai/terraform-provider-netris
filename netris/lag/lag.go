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
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/netrisai/netriswebapi/http"
	"github.com/netrisai/netriswebapi/v2/types/port"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	api "github.com/netrisai/netriswebapi/v2"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: "Manages LAGs",
		Schema: map[string]*schema.Schema{
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "LAG desired description",
			},
			"tenantid": {
				Type:        schema.TypeInt,
				Required:    true,
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
				Required:    true,
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
							Required:    true,
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
		Create: resourceCreate,
		Read:   resourceRead,
		Update: resourceUpdate,
		Delete: resourceDelete,
		Exists: resourceExists,
	}
}

func DiffSuppress(k, old, new string, d *schema.ResourceData) bool {
	return true
}

func resourceCreate(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	ports := []port.IDName{}
	portList := d.Get("members").(*schema.Set).List()
	for _, p := range portList {
		ports = append(ports, port.IDName{Name: p.(string)})
	}

	description := d.Get("description").(string)
	tenantID := d.Get("tenantid").(int)

	lagAdd := &port.PortLAG{}
	lagAdd.Description = description
	lagAdd.Tenant = port.IDName{ID: tenantID}
	lagAdd.Ports = ports

	mtu := d.Get("mtu").(int)
	lacp := d.Get("lacp").(string)
	autoneg := d.Get("autoneg").(string)
	if autoneg == "default" {
		autoneg = "none"
	}
	if autoneg == "default" {
		autoneg = "none"
	}

	extension := port.PortLAGExtension{}
	ext := d.Get("extension").(map[string]interface{})
	if n, ok := ext["extensionname"]; ok {
		extensionName := n.(string)
		if e, ok := findExtensionByName(extensionName, clientset); ok {
			extension.ID = e.ID
		} else if v, ok := ext["vlanrange"]; ok {
			vlanrange := strings.Split(v.(string), "-")
			from, _ := strconv.Atoi(vlanrange[0])
			to, _ := strconv.Atoi(vlanrange[1])
			extension.VlanFrom = from
			extension.VlanTo = to
			extension.Name = extensionName
		} else {
			return fmt.Errorf("please provide vlan range for extension \"%s\"", extensionName)
		}
	}

	lagAdd.Mtu = mtu
	lagAdd.Extension = extension
	lagAdd.LACP = lacp

	js, _ := json.Marshal(lagAdd)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.Port().AddToLAG(lagAdd)
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

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	id, _ := strconv.Atoi(d.Id())

	ports := []port.IDName{}
	portList := d.Get("members").(*schema.Set).List()
	for _, p := range portList {
		ports = append(ports, port.IDName{Name: p.(string)})
	}

	description := d.Get("description").(string)
	tenantID := d.Get("tenantid").(int)

	lagAdd := &port.PortLAG{}
	lagAdd.ID = id
	lagAdd.Description = description
	lagAdd.Tenant = port.IDName{ID: tenantID}
	lagAdd.Ports = ports

	mtu := d.Get("mtu").(int)
	lacp := d.Get("lacp").(string)
	autoneg := d.Get("autoneg").(string)
	if autoneg == "default" {
		autoneg = "none"
	}
	if autoneg == "default" {
		autoneg = "none"
	}

	extension := port.PortLAGExtension{}
	ext := d.Get("extension").(map[string]interface{})
	if n, ok := ext["extensionname"]; ok {
		extensionName := n.(string)
		if e, ok := findExtensionByName(extensionName, clientset); ok {
			extension.ID = e.ID
		} else if v, ok := ext["vlanrange"]; ok {
			vlanrange := strings.Split(v.(string), "-")
			from, _ := strconv.Atoi(vlanrange[0])
			to, _ := strconv.Atoi(vlanrange[1])
			extension.VlanFrom = from
			extension.VlanTo = to
			extension.Name = extensionName
		} else {
			return fmt.Errorf("please provide vlan range for extension \"%s\"", extensionName)
		}
	}

	lagAdd.Mtu = mtu
	lagAdd.Extension = extension
	lagAdd.LACP = lacp

	js, _ := json.Marshal(lagAdd)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.Port().AddToLAG(lagAdd)
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

	d.SetId(strconv.Itoa(id))

	return nil
}

func resourceDelete(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	id, _ := strconv.Atoi(d.Id())

	reply, err := clientset.Port().Delete(id)
	if err != nil {
		log.Println("[DEBUG]", err)
		return err
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
