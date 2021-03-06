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

package inventoryprofile

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/netrisai/netriswebapi/http"
	"github.com/netrisai/netriswebapi/v1/types/inventoryprofile"

	api "github.com/netrisai/netriswebapi/v2"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: "Creates and manages inventory profiles",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of inventory profile",
			},
			"description": {
				Computed:    true,
				Optional:    true,
				Type:        schema.TypeString,
				Description: "Inventory profile description",
			},
			"ipv4ssh": {
				Required:    true,
				Type:        schema.TypeList,
				Description: "List of IPv4 subnets allowed to ssh. Example `[\"10.0.10.0/24\", \"172.16.16.16\"]`",
				Elem: &schema.Schema{
					ValidateFunc: validateIP,
					Type:         schema.TypeString,
				},
			},
			"ipv6ssh": {
				Optional:    true,
				Type:        schema.TypeList,
				Description: "List of IPv6 subnets allowed to ssh. Example `[\"2001:DB8::/32\"]`",
				Elem: &schema.Schema{
					ValidateFunc: validateIP,
					Type:         schema.TypeString,
				},
			},
			"timezone": {
				ValidateFunc: validateTimeZone,
				Optional:     true,
				Type:         schema.TypeString,
				Description:  "Devices using this inventory profile will adjust their system time to the selected timezone. Valid value is a name from the TZ [database](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones).",
			},
			"ntpservers": {
				Optional:    true,
				Type:        schema.TypeList,
				Description: "List of domain names or IP addresses of NTP servers. Example `[\"0.pool.ntp.org\", \"132.163.96.5\"]`",
				Elem: &schema.Schema{
					ValidateFunc: validateNTP,
					Type:         schema.TypeString,
				},
			},
			"dnsservers": {
				Optional:    true,
				Type:        schema.TypeList,
				Description: "List of IP addresses of DNS servers. Example `[\"1.1.1.1\", \"8.8.8.8\"]`",
				Elem: &schema.Schema{
					ValidateFunc: validateIP,
					Type:         schema.TypeString,
				},
			},
			"customrule": {
				Optional:    true,
				Type:        schema.TypeList,
				Description: "Custom Rules configuration block. User defined rules to allow certain traffic.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"sourcesubnet": {
							ValidateFunc: validateIPPrefix,
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Source Subnet. Example `10.0.0.0/8`",
						},
						"srcport": {
							ValidateFunc: validatePort,
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Source port. 1-65535, or empty for any.",
						},
						"dstport": {
							ValidateFunc: validatePort,
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Destination port. 1-65535, or empty for any.",
						},
						"protocol": {
							ValidateFunc: validateProtocol,
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Protocol. Valid value is `udp`, `tcp` or `any`.",
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
	description := d.Get("description").(string)
	ipv4List := []string{}
	ipv4tmp := d.Get("ipv4ssh").([]interface{})
	for _, s := range ipv4tmp {
		ipv4List = append(ipv4List, s.(string))
	}

	ipv6List := []string{}
	ipv6tmp := d.Get("ipv6ssh").([]interface{})
	for _, s := range ipv6tmp {
		ipv6List = append(ipv6List, s.(string))
	}

	timezone := d.Get("timezone").(string)

	ntpList := []string{}
	ntptmp := d.Get("ntpservers").([]interface{})
	for _, s := range ntptmp {
		ntpList = append(ntpList, s.(string))
	}

	dnsList := []string{}
	dnstmp := d.Get("dnsservers").([]interface{})
	for _, s := range dnstmp {
		dnsList = append(dnsList, s.(string))
	}

	customRulesTmp := d.Get("customrule").([]interface{})
	var customRulesList []map[string]interface{}
	for _, customRule := range customRulesTmp {
		customRulesList = append(customRulesList, customRule.(map[string]interface{}))
	}

	customRules := []inventoryprofile.CustomRule{}

	for _, customRule := range customRulesList {
		customRules = append(customRules, inventoryprofile.CustomRule{
			SrcSubnet: customRule["sourcesubnet"].(string),
			SrcPort:   customRule["srcport"].(string),
			DstPort:   customRule["dstport"].(string),
			Protocol:  customRule["protocol"].(string),
		})
	}

	profileAdd := &inventoryprofile.ProfileW{
		Name:        name,
		Description: description,
		Ipv4List:    strings.Join(ipv4List, ","),
		Ipv6List:    strings.Join(ipv6List, ","),
		Timezone:    inventoryprofile.Timezone{Label: timezone, TzCode: timezone},
		NTPServers:  strings.Join(ntpList, ","),
		DNSServers:  strings.Join(dnsList, ","),
		CustomRules: customRules,
	}

	js, _ := json.Marshal(profileAdd)
	log.Println("[DEBUG] inventoryProfileAdd", string(js))

	reply, err := clientset.InventoryProfile().Add(profileAdd)
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
	var profile *inventoryprofile.Profile
	var ok bool
	id, _ := strconv.Atoi(d.Id())
	profile, ok = findByID(id, clientset)
	if !ok {
		return fmt.Errorf("Coudn't find inventory profile '%s'", d.Get("name").(string))
	}

	d.SetId(strconv.Itoa(profile.ID))
	err := d.Set("name", profile.Name)
	if err != nil {
		return err
	}
	err = d.Set("description", profile.Description)
	if err != nil {
		return err
	}
	if (d.Get("ipv4ssh") != nil && len(d.Get("ipv4ssh").([]interface{})) > 0) || profile.Ipv4SSH != "" {
		err = d.Set("ipv4ssh", strings.Split(profile.Ipv4SSH, ","))
		if err != nil {
			return err
		}
	}
	if (d.Get("ipv6ssh") != nil && len(d.Get("ipv6ssh").([]interface{})) > 0) || profile.Ipv6SSH != "" {
		err = d.Set("ipv6ssh", strings.Split(profile.Ipv6SSH, ","))
		if err != nil {
			return err
		}
	}

	err = d.Set("timezone", unmarshalTimezone(profile.Timezone).TzCode)
	if err != nil {
		return err
	}
	err = d.Set("ntpservers", strings.Split(profile.NTPServers, ","))
	if err != nil {
		return err
	}
	err = d.Set("dnsservers", strings.Split(profile.DNSServers, ","))
	if err != nil {
		return err
	}

	var customRules []map[string]interface{}
	for _, rule := range profile.CustomRules {
		customRule := make(map[string]interface{})
		customRule["sourcesubnet"] = rule.SrcSubnet
		customRule["srcport"] = rule.SrcPort
		customRule["dstport"] = rule.DstPort
		customRule["protocol"] = rule.Protocol
		customRules = append(customRules, customRule)
	}

	err = d.Set("customrule", customRules)
	if err != nil {
		return err
	}

	return nil
}

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	name := d.Get("name").(string)
	description := d.Get("description").(string)
	ipv4List := []string{}
	ipv4tmp := d.Get("ipv4ssh").([]interface{})
	for _, s := range ipv4tmp {
		ipv4List = append(ipv4List, s.(string))
	}

	ipv6List := []string{}
	ipv6tmp := d.Get("ipv6ssh").([]interface{})
	for _, s := range ipv6tmp {
		ipv6List = append(ipv6List, s.(string))
	}

	timezone := d.Get("timezone").(string)

	ntpList := []string{}
	ntptmp := d.Get("ntpservers").([]interface{})
	for _, s := range ntptmp {
		ntpList = append(ntpList, s.(string))
	}

	dnsList := []string{}
	dnstmp := d.Get("dnsservers").([]interface{})
	for _, s := range dnstmp {
		dnsList = append(dnsList, s.(string))
	}

	customRulesTmp := d.Get("customrule").([]interface{})
	var customRulesList []map[string]interface{}
	for _, customRule := range customRulesTmp {
		customRulesList = append(customRulesList, customRule.(map[string]interface{}))
	}

	customRules := []inventoryprofile.CustomRule{}

	for _, customRule := range customRulesList {
		customRules = append(customRules, inventoryprofile.CustomRule{
			SrcSubnet: customRule["sourcesubnet"].(string),
			SrcPort:   customRule["srcport"].(string),
			DstPort:   customRule["dstport"].(string),
			Protocol:  customRule["protocol"].(string),
		})
	}
	id, _ := strconv.Atoi(d.Id())
	profileUpdate := &inventoryprofile.ProfileW{
		ID:          id,
		Name:        name,
		Description: description,
		Ipv4List:    strings.Join(ipv4List, ","),
		Ipv6List:    strings.Join(ipv6List, ","),
		Timezone:    inventoryprofile.Timezone{Label: timezone, TzCode: timezone},
		NTPServers:  strings.Join(ntpList, ","),
		DNSServers:  strings.Join(dnsList, ","),
		CustomRules: customRules,
	}

	js, _ := json.Marshal(profileUpdate)
	log.Println("[DEBUG] inventoryProfileUpdate", string(js))

	reply, err := clientset.InventoryProfile().Update(profileUpdate)
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

func resourceExists(d *schema.ResourceData, m interface{}) (bool, error) {
	clientset := m.(*api.Clientset)
	id, _ := strconv.Atoi(d.Id())
	_, ok := findByID(id, clientset)
	return ok, nil
}

func resourceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	clientset := m.(*api.Clientset)
	name := d.Id()
	var profile *inventoryprofile.Profile
	var ok bool
	profile, ok = findByName(name, clientset)
	if !ok {
		return []*schema.ResourceData{d}, fmt.Errorf("Coudn't find inventory profile '%s'", d.Get("name").(string))
	}
	d.SetId(strconv.Itoa(profile.ID))

	return []*schema.ResourceData{d}, nil
}

func resourceDelete(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	id, _ := strconv.Atoi(d.Id())
	reply, err := clientset.InventoryProfile().Delete(id)
	if err != nil {
		return err
	}

	if reply.StatusCode != 200 {
		return fmt.Errorf(string(reply.Data))
	}

	d.SetId("")
	return nil
}
