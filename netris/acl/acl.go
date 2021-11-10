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

package acl

import (
	"encoding/json"
	"log"

	"github.com/netrisai/netriswebapi/v1/types/acl"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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
			"action": {
				Required: true,
				Type:     schema.TypeString,
			},
			"comment": {
				Default:  "",
				Optional: true,
				Type:     schema.TypeString,
			},
			"established": {
				ValidateFunc: validateEstablished,
				Default:      1,
				Optional:     true,
				Type:         schema.TypeInt,
			},
			"icmptype": {
				ValidateFunc: validateICMP,
				Default:      1,
				Optional:     true,
				Type:         schema.TypeInt,
			},
			"proto": {
				ValidateFunc: validateProto,
				Required:     true,
				Type:         schema.TypeString,
			},
			"reverse": {
				Default:  true,
				Optional: true,
				Type:     schema.TypeBool,
			},
			"srcprefix": {
				ValidateFunc: validateIPPrefix,
				Required:     true,
				Type:         schema.TypeString,
			},
			"srcportfrom": {
				ValidateFunc: validatePort,
				Optional:     true,
				Type:         schema.TypeInt,
			},
			"srcportto": {
				ValidateFunc: validatePort,
				Optional:     true,
				Type:         schema.TypeInt,
			},
			"srcportgroup": {
				Optional: true,
				Type:     schema.TypeString,
			},
			"dstprefix": {
				ValidateFunc: validateIPPrefix,
				Required:     true,
				Type:         schema.TypeString,
			},
			"dstportfrom": {
				ValidateFunc: validatePort,
				Optional:     true,
				Type:         schema.TypeInt,
			},
			"dstportto": {
				ValidateFunc: validatePort,
				Optional:     true,
				Type:         schema.TypeInt,
			},
			"dstportgroup": {
				Optional: true,
				Type:     schema.TypeString,
			},
			"validuntil": {
				ValidateFunc: validateDate,
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
	// clientset := m.(*api.Clientset)

	name := d.Get("name").(string)
	action := d.Get("action").(string)
	comment := d.Get("comment").(string)

	established := d.Get("established").(int)

	icmptype := d.Get("icmptype").(int)
	proto := d.Get("proto").(string)

	reverse := "yes"
	if r := d.Get("reverse").(bool); !r {
		reverse = "no"
	}

	srcprefix := d.Get("srcprefix").(string)
	// srcportfrom := d.Get("srcportfrom").(int)
	// srcportto := d.Get("srcportto").(int)
	// srcportgroup := d.Get("srcportgroup").(string)

	dstprefix := d.Get("dstprefix").(string)
	// dstportfrom := d.Get("dstportfrom").(string)
	// dstportto := d.Get("dstportto").(string)
	// dstportgroup := d.Get("dstportgroup").(int)

	// validuntil := d.Get("validuntil").(string)

	aclW := &acl.ACLw{
		Name:        name,
		Action:      action,
		Comment:     comment,
		Established: established,
		ICMPType:    icmptype,
		Proto:       proto,
		Reverse:     reverse,
		SrcPrefix:   srcprefix,
		DstPrefix:   dstprefix,
	}

	js, _ := json.Marshal(aclW)
	log.Println("[DEBUG]", string(js))

	// reply, err := clientset.ACL().Add(aclW)
	// if err != nil {
	// 	log.Println("[DEBUG]", err)
	// 	return err
	// }

	// js, _ = json.Marshal(reply)
	// log.Println("[DEBUG]", string(js))

	// log.Println("[DEBUG]", string(reply.Data))

	// idStruct := struct {
	// 	ID int `json:"id"`
	// }{}

	// data, err := reply.Parse()
	// if err != nil {
	// 	log.Println("[DEBUG]", err)
	// 	return err
	// }

	// err = http.Decode(data.Data, &idStruct)
	// if err != nil {
	// 	log.Println("[DEBUG]", err)
	// 	return err
	// }

	// log.Println("[DEBUG] ID:", idStruct.ID)

	// if reply.StatusCode != 200 {
	// 	return fmt.Errorf(string(reply.Data))
	// }

	// _ = d.Set("itemid", idStruct.ID)
	// d.SetId(aclW.Name)

	return nil
}

func resourceRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceExists(d *schema.ResourceData, m interface{}) (bool, error) {
	return true, nil
}

func resourceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	return []*schema.ResourceData{d}, nil
}
