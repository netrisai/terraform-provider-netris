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

package roh

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/netrisai/netriswebapi/http"
	api "github.com/netrisai/netriswebapi/v2"
	"github.com/netrisai/netriswebapi/v2/types/roh"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: "Creates and manages Instances (ROH)",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Instance name. If type == `hypervisor` the name must be the same as the hypervisor's hostname",
			},
			"tenantid": {
				Required:    true,
				Type:        schema.TypeInt,
				ForceNew:    true,
				Description: "ID of tenant. Users of this tenant will be permitted to manage instance",
			},
			"siteid": {
				Required:    true,
				Type:        schema.TypeInt,
				Description: "The site ID where the current ROH instance belongs",
			},
			"type": {
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateType,
				Type:         schema.TypeString,
				Description:  "Possible values: `physical` or `hypervisor` Physical Server, for all servers forming a BGP adjacency directly with the switch fabric. Hypervisor, for using the hypervisor as an interim router. Proxmox is currently the only supported hypervisor.",
			},
			"routingprofile": {
				ValidateFunc: validateRProfile,
				Default:      "inherit",
				Optional:     true,
				Type:         schema.TypeString,
				Description:  "Possible values: `inherit`, `default`, `default_agg`, `full_table`. Default value is `inherit`. Detailed documentation about routing profiles is available [here](https://www.netris.ai/docs/en/stable/roh.html#adding-roh-hosts)",
			},
			"unicastips": {
				Required:    true,
				Type:        schema.TypeList,
				Description: "List of IPv4 addresses for the loopback interface.",
				Elem: &schema.Schema{
					ValidateFunc: validateIP,
					Type:         schema.TypeString,
				},
			},
			"anycastips": {
				Required:    true,
				Type:        schema.TypeList,
				Description: "List of anycast IP addresses",
				Elem: &schema.Schema{
					ValidateFunc: validateIP,
					Type:         schema.TypeString,
				},
			},
			"ports": {
				Required:    true,
				Type:        schema.TypeList,
				Description: "List of physical switch ports",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"inboundprefixlist": {
				Optional:    true,
				Type:        schema.TypeList,
				Description: "List of additional prefixes that the ROH server may advertise. Only when type == `hypervisor`",
				Elem: &schema.Schema{
					ValidateFunc: validatePrefixRule,
					Type:         schema.TypeString,
				},
			},
		},
		Create: resourceCreate,
		Read:   resourceRead,
		Update: resourceUpdate,
		Delete: resourceDelete,
		Exists: resourceExists,
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

	rohType := d.Get("type").(string)
	portList := d.Get("ports").([]interface{})
	ports := []roh.IDName{}
	for _, port := range portList {
		ports = append(ports, roh.IDName{Name: port.(string)})
	}

	prefixList := d.Get("inboundprefixlist").([]interface{})
	inboundList := []roh.InboundPrefixW{}

	rohAdd := &roh.ROHw{
		Name:            d.Get("name").(string),
		Tenant:          roh.IDName{ID: d.Get("tenantid").(int)},
		Site:            roh.IDName{ID: d.Get("siteid").(int)},
		Type:            d.Get("type").(string),
		Ports:           ports,
		InboundPrefixes: inboundList,
	}

	addresses := []roh.Address{}
	unicastList := d.Get("unicastips").([]interface{})
	anycastList := d.Get("anycastips").([]interface{})
	for _, anycast := range anycastList {
		addresses = append(addresses, roh.Address{Prefix: anycast.(string), Anycast: true})
	}
	rohAdd.Addresses = addresses
	if rohType == "physical" {
		rohAdd.RoutingProfile = d.Get("routingprofile").(string)
		for _, anycast := range unicastList {
			rohAdd.Addresses = append(rohAdd.Addresses, roh.Address{Prefix: anycast.(string), Anycast: false})
		}
	} else {
		if len(unicastList) != 1 {
			return fmt.Errorf("Hypervisor can only have one unicast ip address")
		}
		rohAdd.UnicastAddress = unicastList[0].(string)
		for _, pfx := range prefixList {
			inboundList = append(inboundList, parsePrefixList(pfx.(string)))
		}

		rohAdd.InboundPrefixes = inboundList
	}

	js, _ := json.Marshal(rohAdd)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.ROH().Add(rohAdd)
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

	roh, err := clientset.ROH().GetByID(id)
	if err != nil {
		return err
	}

	d.SetId(strconv.Itoa(roh.ID))
	err = d.Set("name", roh.Name)
	if err != nil {
		return err
	}
	err = d.Set("tenantid", roh.Tenant.ID)
	if err != nil {
		return err
	}
	err = d.Set("siteid", roh.Site.ID)
	if err != nil {
		return err
	}
	err = d.Set("type", roh.Type)
	if err != nil {
		return err
	}

	unicastIPs := []string{}
	anycastIPs := []string{}
	for _, address := range roh.Addresses {
		if address.Anycast {
			anycastIPs = append(anycastIPs, address.Prefix)
		} else {
			unicastIPs = append(unicastIPs, address.Prefix)
		}
	}
	if roh.Type == "hypervisor" {
		unicastIPs = []string{roh.UnicastAddress}
	} else {
		err = d.Set("routingprofile", roh.RoutingProfile.Tag)
		if err != nil {
			return err
		}
	}
	err = d.Set("unicastips", unicastIPs)
	if err != nil {
		return err
	}
	err = d.Set("anycastips", anycastIPs)
	if err != nil {
		return err
	}

	ports := []string{}
	for _, port := range roh.Ports {
		ports = append(ports, fmt.Sprintf("%s@%s", port.Port_, port.Switch.Name))
	}
	err = d.Set("ports", ports)
	if err != nil {
		return err
	}

	terInboundPrefixes := []string{}
	terInbounds := d.Get("inboundprefixlist").([]interface{})
	for _, pfx := range terInbounds {
		terInboundPrefixes = append(terInboundPrefixes, pfx.(string))
	}

	sort.Strings(terInboundPrefixes)

	inboundPrefixes := []string{}
	for _, inboundPrefix := range roh.InboundPrefixes {
		inboundPrefixes = append(inboundPrefixes, fmt.Sprintf("%s %s %s", inboundPrefix.Action, inboundPrefix.Subnet.Prefix, inboundPrefix.Condition))
	}

	sort.Strings(inboundPrefixes)
	updateInbounds := false
	if len(inboundPrefixes) != len(terInboundPrefixes) {
		updateInbounds = true
	} else {
		for i, terInboundPrefix := range terInboundPrefixes {
			terInbound := parsePrefixList(terInboundPrefix)
			if i < len(inboundPrefixes) {
				Inbound := parsePrefixList(inboundPrefixes[i])
				if !(terInbound.Action == Inbound.Action && terInbound.Condition == Inbound.Condition && terInbound.Subnet == Inbound.Subnet) {
					updateInbounds = true
				}
			} else {
				updateInbounds = true
			}
		}
	}

	if updateInbounds {
		err = d.Set("inboundprefixlist", inboundPrefixes)
		if err != nil {
			return err
		}
	}

	return nil
}

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	rohType := d.Get("type").(string)
	portList := d.Get("ports").([]interface{})
	ports := []roh.IDName{}
	for _, port := range portList {
		ports = append(ports, roh.IDName{Name: port.(string)})
	}

	prefixList := d.Get("inboundprefixlist").([]interface{})
	inboundList := []roh.InboundPrefixW{}

	rohAdd := &roh.ROHw{
		Name:            d.Get("name").(string),
		Tenant:          roh.IDName{ID: d.Get("tenantid").(int)},
		Site:            roh.IDName{ID: d.Get("siteid").(int)},
		Type:            d.Get("type").(string),
		Ports:           ports,
		InboundPrefixes: inboundList,
	}

	addresses := []roh.Address{}
	unicastList := d.Get("unicastips").([]interface{})
	anycastList := d.Get("anycastips").([]interface{})
	for _, anycast := range anycastList {
		addresses = append(addresses, roh.Address{Prefix: anycast.(string), Anycast: true})
	}
	rohAdd.Addresses = addresses
	if rohType == "physical" {
		rohAdd.RoutingProfile = d.Get("routingprofile").(string)
		for _, anycast := range unicastList {
			rohAdd.Addresses = append(rohAdd.Addresses, roh.Address{Prefix: anycast.(string), Anycast: false})
		}
	} else {
		if len(unicastList) != 1 {
			return fmt.Errorf("Hypervisor can only have one unicast ip address")
		}
		rohAdd.UnicastAddress = unicastList[0].(string)
		for _, pfx := range prefixList {
			inboundList = append(inboundList, parsePrefixList(pfx.(string)))
		}

		rohAdd.InboundPrefixes = inboundList
	}

	js, _ := json.Marshal(rohAdd)
	log.Println("[DEBUG]", string(js))

	id, _ := strconv.Atoi(d.Id())
	reply, err := clientset.ROH().Update(id, rohAdd)
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

	return nil
}

func resourceDelete(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	id, _ := strconv.Atoi(d.Id())
	reply, err := clientset.ROH().Delete(id)
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

	_, err := clientset.ROH().GetByID(id)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func resourceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	clientset := m.(*api.Clientset)

	rohs, _ := clientset.ROH().Get()
	name := d.Id()
	for _, roh := range rohs {
		if roh.Name == name {
			d.SetId(strconv.Itoa(roh.ID))
			return []*schema.ResourceData{d}, nil
		}
	}

	return []*schema.ResourceData{d}, nil
}
