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

package site

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/netrisai/netriswebapi/http"
	"github.com/netrisai/netriswebapi/v1/types/site"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	api "github.com/netrisai/netriswebapi/v2"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"itemid": {
				Type:             schema.TypeInt,
				Optional:         true,
				Computed: true,
				DiffSuppressFunc: DiffSuppress,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the resource, also acts as it's unique ID",
			},
			"publicasn": {
				Required: true,
				Type:     schema.TypeInt,
			},
			"rohasn": {
				Required: true,
				Type:     schema.TypeInt,
			},
			"vmasn": {
				Required: true,
				Type:     schema.TypeInt,
			},
			"rohroutingprofile": {
				ValidateFunc: validateRoutingProfile,
				Required:     true,
				Type:         schema.TypeString,
				Description:  "Routing profile available values are (default, default_agg, full_table)",
			},
			"sitemesh": {
				ValidateFunc: validateSiteMesh,
				Required:     true,
				Type:         schema.TypeString,
				Description:  "Site mesh available values are (disabled, hub, spoke, dspoke)",
			},
			"acldefaultpolicy": {
				ValidateFunc: validateACLPolicy,
				Required:     true,
				Type:         schema.TypeString,
				Description:  "ACL policy available values are (permit, deny)",
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
	publicasn := d.Get("publicasn").(int)
	rohasn := d.Get("rohasn").(int)
	vmasn := d.Get("vmasn").(int)

	siteW := &site.SiteAdd{
		Name:                name,
		PublicASN:           publicasn,
		PhysicalInstanceASN: rohasn,
		VirtualInstanceASN:  vmasn,
		RoutingProfileID:    routingProfiles[d.Get("rohroutingprofile").(string)],
		VPN:                 d.Get("sitemesh").(string),
		ACLPolicy:           d.Get("acldefaultpolicy").(string),
	}

	js, _ := json.Marshal(siteW)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.Site().Add(siteW)
	if err != nil {
		log.Println("[DEBUG]", err)
		return err
	}

	js, _ = json.Marshal(reply)
	log.Println("[DEBUG]", string(js))

	log.Println("[DEBUG]", string(reply.Data))

	id := 0

	data, err := reply.Parse()
	if err != nil {
		log.Println("[DEBUG]", err)
		return err
	}

	err = http.Decode(data.Data, &id)
	if err != nil {
		log.Println("[DEBUG]", err)
		return err
	}

	log.Println("[DEBUG] ID:", id)

	if reply.StatusCode != 200 {
		return fmt.Errorf(string(reply.Data))
	}

	_ = d.Set("itemid", id)
	d.SetId(siteW.Name)

	return nil
}

func resourceRead(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)
	var site *site.Site
	sites, err := clientset.Site().Get()
	if err != nil {
		return err
	}

	for _, s := range sites {
		if s.ID == d.Get("itemid").(int) {
			site = s
			break
		}
	}

	if site == nil {
		return nil
	}

	d.SetId(site.Name)
	err = d.Set("name", site.Name)
	if err != nil {
		return err
	}
	err = d.Set("publicasn", site.PublicASN)
	if err != nil {
		return err
	}
	err = d.Set("rohasn", site.PhysicalInstanceAsn)
	if err != nil {
		return err
	}
	err = d.Set("vmasn", site.VirtualInstanceASN)
	if err != nil {
		return err
	}
	err = d.Set("rohroutingprofile", site.RoutingProfilTag)
	if err != nil {
		return err
	}
	err = d.Set("sitemesh", site.VPN)
	if err != nil {
		return err
	}
	err = d.Set("acldefaultpolicy", site.ACLPolicy)
	if err != nil {
		return err
	}

	return nil
}

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	name := d.Get("name").(string)
	publicasn := d.Get("publicasn").(int)
	rohasn := d.Get("rohasn").(int)
	vmasn := d.Get("vmasn").(int)

	siteW := &site.SiteAdd{
		ID:                  d.Get("itemid").(int),
		Name:                name,
		PublicASN:           publicasn,
		PhysicalInstanceASN: rohasn,
		VirtualInstanceASN:  vmasn,
		RoutingProfileID:    routingProfiles[d.Get("rohroutingprofile").(string)],
		VPN:                 d.Get("sitemesh").(string),
		ACLPolicy:           d.Get("acldefaultpolicy").(string),
	}

	js, _ := json.Marshal(siteW)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.Site().Update(siteW)
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

	reply, err := clientset.Site().Delete(d.Get("itemid").(int))
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
	siteID := d.Get("itemid").(int)

	sites, err := clientset.Site().Get()
	if err != nil {
		return false, err
	}

	for _, site := range sites {
		if siteID == site.ID {
			return true, nil
		}
	}

	return false, nil
}

func resourceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	clientset := m.(*api.Clientset)

	sites, _ := clientset.Site().Get()
	name := d.Id()
	for _, site := range sites {
		if site.Name == name {
			err := d.Set("itemid", site.ID)
			if err != nil {
				return []*schema.ResourceData{d}, err
			}
			return []*schema.ResourceData{d}, nil
		}
	}

	return []*schema.ResourceData{d}, nil
}
