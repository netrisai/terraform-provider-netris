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

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/netrisai/netriswebapi/http"
	api "github.com/netrisai/netriswebapi/v2"
	"github.com/netrisai/netriswebapi/v2/types/roh"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"itemid": {
				Type:             schema.TypeInt,
				Optional:         true,
				DiffSuppressFunc: DiffSuppress,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the resource, also acts as it's unique ID",
			},
			"tenant": {
				Required: true,
				Type:     schema.TypeString,
				ForceNew: true,
			},
			"site": {
				Required: true,
				Type:     schema.TypeString,
			},
			"type": {
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateType,
				Type:         schema.TypeString,
			},
			"routingprofile": {
				ValidateFunc: validateRProfile,
				Default:      "inherit",
				Optional:     true,
				Type:         schema.TypeString,
			},
			"unicastips": {
				Required:    true,
				Type:        schema.TypeList,
				Description: "Unicast IP addresses",
				Elem: &schema.Schema{
					ValidateFunc: validateIP,
					Type:         schema.TypeString,
				},
			},
			"anycastips": {
				Required:    true,
				Type:        schema.TypeList,
				Description: "Anycast IP addresses",
				Elem: &schema.Schema{
					ValidateFunc: validateIP,
					Type:         schema.TypeString,
				},
			},
			"ports": {
				Required:    true,
				Type:        schema.TypeList,
				Description: "Ports",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"inboundprefixlist": {
				Optional:    true,
				Type:        schema.TypeList,
				Description: "Inbound prefix list",
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
		Tenant:          roh.IDName{Name: d.Get("tenant").(string)},
		Site:            roh.IDName{Name: d.Get("site").(string)},
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

	_ = d.Set("itemid", idStruct.ID)
	d.SetId(rohAdd.Name)

	return nil
}

func resourceRead(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)
	id := d.Get("itemid").(int)

	roh, err := clientset.ROH().GetByID(id)
	if err != nil {
		return err
	}

	d.SetId(roh.Name)
	err = d.Set("name", roh.Name)
	if err != nil {
		return err
	}
	err = d.Set("tenant", roh.Tenant.Name)
	if err != nil {
		return err
	}
	err = d.Set("site", roh.Site.Name)
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
		Tenant:          roh.IDName{Name: d.Get("tenant").(string)},
		Site:            roh.IDName{Name: d.Get("site").(string)},
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

	reply, err := clientset.ROH().Update(d.Get("itemid").(int), rohAdd)
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

	reply, err := clientset.ROH().Delete(d.Get("itemid").(int))
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
	id := d.Get("itemid").(int)

	_, err := clientset.ROH().GetByID(id)
	if err != nil {
		return false, err
	}
	return true, nil
}

func resourceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	clientset := m.(*api.Clientset)

	rohs, _ := clientset.ROH().Get()
	name := d.Id()
	for _, roh := range rohs {
		if roh.Name == name {
			err := d.Set("itemid", roh.ID)
			if err != nil {
				return []*schema.ResourceData{d}, err
			}
			return []*schema.ResourceData{d}, nil
		}
	}

	return []*schema.ResourceData{d}, nil
}