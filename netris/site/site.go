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
	"github.com/netrisai/netriswebapi/v2/types/site"

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
				Description:  "Possible values: `equinix_metal`, `phoenixnap_bmc`, `dot1q_trunk`, `netris`.",
			},
			"vlanrange": {
				Computed:     true,
				ValidateFunc: validateVlanRange,
				Optional:     true,
				Type:         schema.TypeString,
				Description:  "VLAN range.",
			},
			"switchfabricproviders": {
				Optional:    true,
				Type:        schema.TypeSet,
				Description: "",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"equinixmetal": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"projectid": {
										Required: true,
										Type:     schema.TypeString,
									},
									"projectapikey": {
										Required: true,
										Type:     schema.TypeString,
									},
									"location": {
										ValidateFunc: validateEquinixLocation,
										Required:     true,
										Type:         schema.TypeString,
									},
								},
							},
						},
						"phoenixnapbmc": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"clientid": {
										Required: true,
										Type:     schema.TypeString,
									},
									"clientsecret": {
										Required: true,
										Type:     schema.TypeString,
									},
									"location": {
										ValidateFunc: validatephoenixLocation,
										Required:     true,
										Type:         schema.TypeString,
									},
								},
							},
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

	siteW := &site.Site{
		Name: name,

		PublicAsn:    publicasn,
		RohAsn:       rohasn,
		VMAsn:        vmasn,
		RohProfile:   &site.RohProfile{ID: routingProfiles[d.Get("rohroutingprofile").(string)]},
		SiteMesh:     site.IDName{Value: d.Get("sitemesh").(string)},
		AclPolicy:    d.Get("acldefaultpolicy").(string),
		SwitchFabric: fabric,
	}

	providersList := d.Get("switchfabricproviders").(*schema.Set).List()
	provider := map[string]interface{}{}
	if len(providersList) > 0 {
		if len(providersList) > 1 {
			return fmt.Errorf("please specify only one switchfabricproviders")
		}
		provider = providersList[0].(map[string]interface{})
	}

	if fabric == "dot1q_trunk" {
		if vlanRange == "" {
			vlanRange = "2-4094"
		}
	} else if fabric == "equinix_metal" {
		if vlanRange == "" {
			vlanRange = "2-3999"
		}
		if err := valEquinixVlanRange(vlanRange); err != nil {
			return err
		}

		detailsmissing := true

		if _, ok := provider["equinixmetal"]; ok {
			l := provider["equinixmetal"].(*schema.Set).List()
			if len(l) > 0 {
				detailsmissing = false
				equinixmetal := l[0].(map[string]interface{})
				siteW.SwitchFabricProviders = &site.SwitchFabricProviders{
					EquinixMetal: &site.EquinixMetal{
						ProjectID:     equinixmetal["projectid"].(string),
						ProjectAPIKey: equinixmetal["projectapikey"].(string),
						Location:      equinixmetal["location"].(string),
					},
				}
			}
		}
		if detailsmissing {
			return fmt.Errorf("please provide equinixmetal details")
		}
	} else if fabric == "phoenixnap_bmc" {
		if vlanRange == "" {
			vlanRange = "2-4094"
		}

		if err := valPhoenixVlanRange(vlanRange); err != nil {
			return err
		}

		detailsmissing := true

		if _, ok := provider["phoenixnapbmc"]; ok {
			l := provider["phoenixnapbmc"].(*schema.Set).List()
			if len(l) > 0 {
				detailsmissing = false
				equinixmetal := l[0].(map[string]interface{})
				siteW.SwitchFabricProviders = &site.SwitchFabricProviders{
					PhoenixNapBmc: &site.PhoenixNapBmc{
						ClientID:     equinixmetal["clientid"].(string),
						ClientSecret: equinixmetal["clientsecret"].(string),
						Location:     equinixmetal["location"].(string),
					},
				}
			}
		}
		if detailsmissing {
			return fmt.Errorf("please provide phoenixnapbmc details")
		}
	}

	siteW.VlanRange = vlanRange

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
	err = d.Set("publicasn", site.PublicAsn)
	if err != nil {
		return err
	}
	err = d.Set("rohasn", site.RohAsn)
	if err != nil {
		return err
	}
	err = d.Set("vmasn", site.VMAsn)
	if err != nil {
		return err
	}
	if site.RohProfile != nil {
		err = d.Set("rohroutingprofile", site.RohProfile.Value)
		if err != nil {
			return err
		}
	}
	err = d.Set("sitemesh", site.SiteMesh.Value)
	if err != nil {
		return err
	}
	err = d.Set("acldefaultpolicy", site.AclPolicy)
	if err != nil {
		return err
	}
	err = d.Set("switchfabric", site.SwitchFabric)
	if err != nil {
		return err
	}
	err = d.Set("vlanrange", site.VlanRange)
	if err != nil {
		return err
	}

	providers := []map[string]interface{}{}

	if site.SwitchFabric == "equinix_metal" && site.SwitchFabricProviders != nil && site.SwitchFabricProviders.EquinixMetal != nil {
		provider := make(map[string]interface{})
		equinixmetal := []map[string]interface{}{}
		p := make(map[string]interface{})
		p["projectid"] = site.SwitchFabricProviders.EquinixMetal.ProjectID
		p["projectapikey"] = site.SwitchFabricProviders.EquinixMetal.ProjectAPIKey
		p["location"] = site.SwitchFabricProviders.EquinixMetal.Location
		equinixmetal = append(equinixmetal, p)
		provider["equinixmetal"] = equinixmetal
		providers = append(providers, provider)
	} else if site.SwitchFabric == "phoenixnap_bmc" && site.SwitchFabricProviders != nil && site.SwitchFabricProviders.PhoenixNapBmc != nil {
		provider := make(map[string]interface{})
		phoenixnapbmc := []map[string]interface{}{}
		p := make(map[string]interface{})
		p["clientid"] = site.SwitchFabricProviders.PhoenixNapBmc.ClientID
		p["clientsecret"] = site.SwitchFabricProviders.PhoenixNapBmc.ClientSecret
		p["location"] = site.SwitchFabricProviders.PhoenixNapBmc.Location
		phoenixnapbmc = append(phoenixnapbmc, p)
		provider["phoenixnapbmc"] = phoenixnapbmc
		providers = append(providers, provider)
	}

	err = d.Set("switchfabricproviders", providers)
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

	siteW := &site.Site{
		Name: name,

		PublicAsn:    publicasn,
		RohAsn:       rohasn,
		VMAsn:        vmasn,
		RohProfile:   &site.RohProfile{ID: routingProfiles[d.Get("rohroutingprofile").(string)]},
		SiteMesh:     site.IDName{Value: d.Get("sitemesh").(string)},
		AclPolicy:    d.Get("acldefaultpolicy").(string),
		SwitchFabric: fabric,
	}

	providersList := d.Get("switchfabricproviders").(*schema.Set).List()
	provider := map[string]interface{}{}
	if len(providersList) > 0 {
		if len(providersList) > 1 {
			return fmt.Errorf("please specify only one switchfabricproviders")
		}
		provider = providersList[0].(map[string]interface{})
	}

	if fabric == "dot1q_trunk" {
		if vlanRange == "" {
			vlanRange = "2-4094"
		}
	} else if fabric == "equinix_metal" {
		if vlanRange == "" {
			vlanRange = "2-3999"
		}
		if err := valEquinixVlanRange(vlanRange); err != nil {
			return err
		}

		detailsmissing := true

		if _, ok := provider["equinixmetal"]; ok {
			l := provider["equinixmetal"].(*schema.Set).List()
			if len(l) > 0 {
				detailsmissing = false
				equinixmetal := l[0].(map[string]interface{})
				siteW.SwitchFabricProviders = &site.SwitchFabricProviders{
					EquinixMetal: &site.EquinixMetal{
						ProjectID:     equinixmetal["projectid"].(string),
						ProjectAPIKey: equinixmetal["projectapikey"].(string),
						Location:      equinixmetal["location"].(string),
					},
				}
			}
		}
		if detailsmissing {
			return fmt.Errorf("please provide equinixmetal details")
		}
	} else if fabric == "phoenixnap_bmc" {
		if vlanRange == "" {
			vlanRange = "2-4094"
		}

		if err := valPhoenixVlanRange(vlanRange); err != nil {
			return err
		}

		detailsmissing := true

		if _, ok := provider["phoenixnapbmc"]; ok {
			l := provider["phoenixnapbmc"].(*schema.Set).List()
			if len(l) > 0 {
				detailsmissing = false
				equinixmetal := l[0].(map[string]interface{})
				siteW.SwitchFabricProviders = &site.SwitchFabricProviders{
					PhoenixNapBmc: &site.PhoenixNapBmc{
						ClientID:     equinixmetal["clientid"].(string),
						ClientSecret: equinixmetal["clientsecret"].(string),
						Location:     equinixmetal["location"].(string),
					},
				}
			}
		}
		if detailsmissing {
			return fmt.Errorf("please provide phoenixnapbmc details")
		}
	}

	siteW.VlanRange = vlanRange

	js, _ := json.Marshal(siteW)
	log.Println("[DEBUG]", string(js))

	id, _ := strconv.Atoi(d.Id())

	reply, err := clientset.Site().Update(id, siteW)
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
