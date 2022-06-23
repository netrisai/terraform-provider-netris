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
	"strconv"

	"github.com/netrisai/netriswebapi/http"
	"github.com/netrisai/netriswebapi/v1/types/site"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	api "github.com/netrisai/netriswebapi/v2"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: "Creates and manages Sites",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the site",
			},
			"publicasn": {
				Required:    true,
				Type:        schema.TypeInt,
				Description: "Site public ASN that should be used for external bgp peer configuration",
			},
			"rohasn": {
				Required:    true,
				Type:        schema.TypeInt,
				Description: "ASN for ROH (Routing on the Host) compute instances, should be unique within the scope of a site, can be same for different sites",
			},
			"vmasn": {
				Required:    true,
				Type:        schema.TypeInt,
				Description: "ASN for ROH (Routing on the Host) virtual compute instances, should be unique within the scope of a site, can be same for different sites",
			},
			"rohroutingprofile": {
				ValidateFunc: validateRoutingProfile,
				Required:     true,
				Type:         schema.TypeString,
				Description:  "ROH Routing profile defines set of routing prefixes to be advertised to ROH instances. Possible values: `default`, `default_agg`, `full`. Default route only - Will advertise 0.0.0.0/0 + loopback address of physically connected switch. Default + Aggregate - Will add prefixes of defined subnets + `Default` profile. Full - Will advertise all prefixes available in the routing table of the connected switch",
			},
			"sitemesh": {
				ValidateFunc: validateSiteMesh,
				Required:     true,
				Type:         schema.TypeString,
				Description:  "Site to site VPN mode. Site mesh available values are: `disabled`, `hub`, `spoke`, `dspoke`",
			},
			"acldefaultpolicy": {
				ValidateFunc: validateACLPolicy,
				Required:     true,
				Type:         schema.TypeString,
				Description:  "Possible values: `permit` or `deny`. Deny - Layer-3 packet forwarding is denied by default. ACLs are required to permit necessary traffic flows. Deny ACLs will be applied before Permit ACLs. Permit - Layer-3 packet forwarding is allowed by default. ACLs are required to deny unwanted traffic flows. Permit ACLs will be applied before Deny ACLs.",
			},
			"switchfabric": {
				ValidateFunc: validateSwitchFabric,
				Default:      "netris",
				Optional:     true,
				Type:         schema.TypeString,
				Description:  "Possible values: `equinix_metal`, `dot1q_trunk`, `netris`.",
			},
			"vlanrange": {
				ValidateFunc: validateVlanRange,
				Optional:     true,
				Type:         schema.TypeString,
				Description:  "VLAN range.",
			},
			"equinixprojectid": {
				Optional:    true,
				Type:        schema.TypeString,
				Description: "Equinix project ID.",
			},
			"equinixprojectapikey": {
				Optional:    true,
				Type:        schema.TypeString,
				Description: "Equinix project API Key.",
			},
			"equinixlocation": {
				ValidateFunc: validateEquinixLocation,
				Optional:     true,
				Type:         schema.TypeString,
				Description:  "Equinix project location. Possible values:`se`, `dc`, `at`, `hk`, `am`, `ny`, `ty`, `sl`, `md`, `sp`, `fr`, `sy`, `ld`, `sg`, `pa`, `tr`, `sv`, `la`, `ch`, `da`",
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
	fabric := d.Get("switchfabric").(string)
	vlanRange := d.Get("vlanrange").(string)

	siteW := &site.SiteAdd{
		Name:                name,
		PublicASN:           publicasn,
		PhysicalInstanceASN: rohasn,
		VirtualInstanceASN:  vmasn,
		RoutingProfileID:    routingProfiles[d.Get("rohroutingprofile").(string)],
		VPN:                 d.Get("sitemesh").(string),
		ACLPolicy:           d.Get("acldefaultpolicy").(string),
		SwitchFabric:        fabric,
	}

	if fabric == "dot1q_trunk" {
		siteW.VLANRange = vlanRange
	} else if fabric == "equinix_metal" {
		if err := valEquinixVlanRange(vlanRange); err != nil {
			return err
		}
		siteW.VLANRange = vlanRange
		siteW.EquinixProjectID = d.Get("equinixprojectid").(string)
		siteW.EquinixProjectAPIKey = d.Get("equinixprojectapikey").(string)
		siteW.EquinixLocation = d.Get("equinixlocation").(string)
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

	var id int

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
	d.SetId(strconv.Itoa(id))

	return nil
}

func resourceRead(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)
	var site *site.Site
	sites, err := clientset.Site().Get()
	if err != nil {
		return err
	}
	id, _ := strconv.Atoi(d.Id())

	for _, s := range sites {
		if s.ID == id {
			site = s
			break
		}
	}

	if site == nil {
		return nil
	}

	d.SetId(strconv.Itoa(site.ID))
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
	err = d.Set("switchfabric", site.SwitchFabric)
	if err != nil {
		return err
	}
	err = d.Set("vlanrange", site.VLANRange)
	if err != nil {
		return err
	}
	err = d.Set("equinixprojectid", site.EquinixProjectID)
	if err != nil {
		return err
	}
	err = d.Set("equinixprojectapikey", site.EquinixProjectAPIKey)
	if err != nil {
		return err
	}
	err = d.Set("equinixlocation", site.EquinixLocation)
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
	fabric := d.Get("switchfabric").(string)
	vlanRange := d.Get("vlanrange").(string)

	id, _ := strconv.Atoi(d.Id())
	siteW := &site.SiteAdd{
		ID:                  id,
		Name:                name,
		PublicASN:           publicasn,
		PhysicalInstanceASN: rohasn,
		VirtualInstanceASN:  vmasn,
		RoutingProfileID:    routingProfiles[d.Get("rohroutingprofile").(string)],
		VPN:                 d.Get("sitemesh").(string),
		ACLPolicy:           d.Get("acldefaultpolicy").(string),
		SwitchFabric:        fabric,
	}

	if fabric == "dot1q_trunk" {
		siteW.VLANRange = vlanRange
	} else if fabric == "equinix_metal" {
		if err := valEquinixVlanRange(vlanRange); err != nil {
			return err
		}
		siteW.VLANRange = vlanRange
		siteW.EquinixProjectID = d.Get("equinixprojectid").(string)
		siteW.EquinixProjectAPIKey = d.Get("equinixprojectapikey").(string)
		siteW.EquinixLocation = d.Get("equinixlocation").(string)
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
	id, _ := strconv.Atoi(d.Id())
	reply, err := clientset.Site().Delete(id)
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
	siteID, _ := strconv.Atoi(d.Id())

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
			d.SetId(strconv.Itoa(site.ID))
			return []*schema.ResourceData{d}, nil
		}
	}

	return []*schema.ResourceData{d}, nil
}
