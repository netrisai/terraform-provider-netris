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

package nat

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/netrisai/netriswebapi/http"
	"github.com/netrisai/netriswebapi/v2/types/nat"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	api "github.com/netrisai/netriswebapi/v2"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the resource, also acts as it's unique ID",
			},
			"state": {
				ValidateFunc: validateState,
				Default:      "enabled",
				Optional:     true,
				Type:         schema.TypeString,
			},
			"comment": {
				Computed: true,
				Optional: true,
				Type:     schema.TypeString,
			},
			"siteid": {
				Required: true,
				Type:     schema.TypeInt,
			},
			"action": {
				ValidateFunc: validateAction,
				Required:     true,
				Type:         schema.TypeString,
			},
			"protocol": {
				ValidateFunc: validateProto,
				Required:     true,
				Type:         schema.TypeString,
			},
			"srcaddress": {
				ValidateFunc: validateIPPrefix,
				Required:     true,
				Type:         schema.TypeString,
			},
			"srcport": {
				Optional: true,
				Type:     schema.TypeString,
			},
			"dstaddress": {
				ValidateFunc: validateIPPrefix,
				Required:     true,
				Type:         schema.TypeString,
			},
			"dstport": {
				Optional: true,
				Type:     schema.TypeString,
			},
			"dnattoip": {
				ValidateFunc: validateIPPrefix,
				Optional:     true,
				Type:         schema.TypeString,
			},
			"dnattoport": {
				Optional: true,
				Type:     schema.TypeString,
			},
			"snattoip": {
				ValidateFunc: validateIPPrefix,
				Optional:     true,
				Type:         schema.TypeString,
			},
			"snattopool": {
				ValidateFunc: validateIPPrefix,
				Optional:     true,
				Type:         schema.TypeString,
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

	name := d.Get("name").(string)
	state := d.Get("state").(string)
	comment := d.Get("comment").(string)
	siteID := d.Get("siteid").(int)
	action := d.Get("action").(string)
	protocol := d.Get("protocol").(string)
	srcaddress := d.Get("srcaddress").(string)
	srcport := d.Get("srcport").(string)
	dstaddress := d.Get("dstaddress").(string)
	dstport := d.Get("dstport").(string)
	dnattoip := d.Get("dnattoip").(string)
	dnattoport := d.Get("dnattoport").(string)
	snattoip := d.Get("snattoip").(string)
	snattopool := d.Get("snattopool").(string)

	natW := &nat.NATw{
		Name:               name,
		Action:             action,
		Comment:            comment,
		State:              state,
		Site:               nat.IDName{ID: siteID},
		Protocol:           protocol,
		SourceAddress:      srcaddress,
		SourcePort:         srcport,
		DestinationAddress: dstaddress,
		DestinationPort:    dstport,
		DnatToIP:           dnattoip,
		DnatToPort:         dnattoport,
		SnatToIP:           snattoip,
		SnatToPool:         snattopool,
	}

	js, _ := json.Marshal(natW)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.NAT().Add(natW)
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
	nat, err := clientset.NAT().GetByID(id)
	if err != nil {
		return err
	}

	d.SetId(strconv.Itoa(nat.ID))
	err = d.Set("name", nat.Name)
	if err != nil {
		return err
	}
	err = d.Set("state", nat.State.Value)
	if err != nil {
		return err
	}
	err = d.Set("comment", nat.Comment)
	if err != nil {
		return err
	}
	err = d.Set("siteid", nat.Site.ID)
	if err != nil {
		return err
	}
	err = d.Set("action", nat.Action.Value)
	if err != nil {
		return err
	}
	err = d.Set("protocol", nat.Protocol.Value)
	if err != nil {
		return err
	}
	err = d.Set("srcaddress", nat.SourceAddress)
	if err != nil {
		return err
	}
	if nat.SourcePort == "1-65535" && d.Get("srcport").(string) == "" {
		err = d.Set("srcport", nat.SourcePort)
		if err != nil {
			return err
		}
	}
	if dstAddr := strings.Split(d.Get("dstaddress").(string), "/")[0]; dstAddr != nat.DestinationAddress {
		err = d.Set("dstaddress", nat.DestinationAddress)
		if err != nil {
			return err
		}
	}
	err = d.Set("dstport", nat.DestinationPort)
	if err != nil {
		return err
	}
	err = d.Set("dnattoip", nat.DnatToIP)
	if err != nil {
		return err
	}
	err = d.Set("dnattoport", strconv.Itoa(nat.DnatToPort))
	if err != nil {
		return err
	}
	err = d.Set("snattoip", nat.SnatToIP)
	if err != nil {
		return err
	}
	err = d.Set("snattopool", nat.SnatToPool)
	if err != nil {
		return err
	}

	return nil
}

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	name := d.Get("name").(string)
	state := d.Get("state").(string)
	comment := d.Get("comment").(string)
	siteID := d.Get("siteid").(int)
	action := d.Get("action").(string)
	protocol := d.Get("protocol").(string)
	srcaddress := d.Get("srcaddress").(string)
	srcport := d.Get("srcport").(string)
	dstaddress := d.Get("dstaddress").(string)
	dstport := d.Get("dstport").(string)
	dnattoip := d.Get("dnattoip").(string)
	dnattoport := d.Get("dnattoport").(string)
	snattoip := d.Get("snattoip").(string)
	snattopool := d.Get("snattopool").(string)

	id, _ := strconv.Atoi(d.Id())
	natW := &nat.NATw{
		Name:               name,
		Action:             action,
		Comment:            comment,
		State:              state,
		Site:               nat.IDName{ID: siteID},
		Protocol:           protocol,
		SourceAddress:      srcaddress,
		SourcePort:         srcport,
		DestinationAddress: dstaddress,
		DestinationPort:    dstport,
		DnatToIP:           dnattoip,
		DnatToPort:         dnattoport,
		SnatToIP:           snattoip,
		SnatToPool:         snattopool,
	}

	js, _ := json.Marshal(natW)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.NAT().Update(id, natW)
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

	return nil
}

func resourceDelete(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)
	id, _ := strconv.Atoi(d.Id())
	reply, err := clientset.NAT().Delete(id)
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
	nat, err := clientset.NAT().GetByID(id)
	if err != nil {
		return false, err
	}

	if nat == nil {
		return false, nil
	}

	return true, nil
}

func resourceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	clientset := m.(*api.Clientset)

	nats, _ := clientset.NAT().Get()
	name := d.Id()
	for _, nat := range nats {
		if nat.Name == name {
			d.SetId(strconv.Itoa(nat.ID))
			return []*schema.ResourceData{d}, nil
		}
	}

	return []*schema.ResourceData{d}, nil
}
