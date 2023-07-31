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
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/netrisai/netriswebapi/v1/types/inventoryprofile"

	api "github.com/netrisai/netriswebapi/v2"
)

func DataResource() *schema.Resource {
	return &schema.Resource{
		Description: "Data Source: inventory profiles",
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
				Optional:    true,
				Type:        schema.TypeList,
				Description: "List of IPv4 subnets allowed to ssh.",
				Elem: &schema.Schema{
					ValidateFunc: validateIP,
					Type:         schema.TypeString,
				},
			},
			"ipv6ssh": {
				Optional:    true,
				Type:        schema.TypeList,
				Description: "List of IPv6 subnets allowed to ssh.",
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
				Description: "List of domain names or IP addresses of NTP servers.",
				Elem: &schema.Schema{
					ValidateFunc: validateNTP,
					Type:         schema.TypeString,
				},
			},
			"dnsservers": {
				Optional:    true,
				Type:        schema.TypeList,
				Description: "List of IP addresses of DNS servers.",
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
							Description:  "Source Subnet.",
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
							Description:  "Protocol.",
						},
					},
				},
			},
		},
		Read:   dataResourceRead,
		Exists: dataResourceExists,
		Importer: &schema.ResourceImporter{
			State: dataRresourceImport,
		},
	}
}

func dataResourceRead(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)
	var profile *inventoryprofile.Profile
	var ok bool
	profile, ok = findByName(d.Get("name").(string), clientset)
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

func dataResourceExists(d *schema.ResourceData, m interface{}) (bool, error) {
	return true, nil
}

func dataRresourceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
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
