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

package dhcpoptionset

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	api "github.com/netrisai/netriswebapi/v2"
	"github.com/netrisai/netriswebapi/v2/types/dhcp"
)

func DataResource() *schema.Resource {
	return &schema.Resource{
		Description: "Creates and manages DHCP Option Set",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "User assigned name of DHCP Option Set.",
			},
			"description": {
				Optional:    true,
				Type:        schema.TypeString,
				Description: "Description.",
			},
			"domainsearch": {
				Optional:    true,
				Type:        schema.TypeString,
				Description: "The domain name that should be used as a suffix when resolving hostnames via the dns servers.",
			},
			"dnsservers": {
				Optional:    true,
				Type:        schema.TypeSet,
				Description: "List of IP addresses of dns servers.",
				Elem: &schema.Schema{
					Type:    schema.TypeString,
					Default: "",
				},
			},
			"ntpservers": {
				Optional:    true,
				Type:        schema.TypeSet,
				Description: "List of IP addresses of ntp servers.",
				Elem: &schema.Schema{
					Type:    schema.TypeString,
					Default: "",
				},
			},
			"leasetime": {
				Default:     86400,
				Optional:    true,
				Type:        schema.TypeInt,
				Description: "The amount of time in seconds a network device can use an IP address.",
			},
			"standardtoption": {
				Optional:    true,
				Type:        schema.TypeSet,
				Description: "User-defined additional DHCP Options",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"code": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Option code",
						},
						"value": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Value of option",
						},
					},
				},
			},
			"customoption": {
				Optional:    true,
				Type:        schema.TypeSet,
				Description: "User-defined custom DHCP Options",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"code": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Option code",
						},
						"type": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Value type",
						},
						"value": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Value of option",
						},
					},
				},
			},
		},
		Read:   dataResourceRead,
		Exists: dataResourceExists,
	}
}

func dataResourceRead(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	name := d.Get("name").(string)

	var apiDHCP *dhcp.DHCPOptionSet

	list, err := clientset.DHCP().Get()
	if err != nil {
		return err
	}

	for _, v := range list {
		if v.Name == name {
			apiDHCP = v
			break
		}
	}

	if apiDHCP == nil {
		return fmt.Errorf("Coudn't find dhcp %s", name)
	}

	d.SetId(strconv.Itoa(apiDHCP.ID))
	err = d.Set("name", apiDHCP.Name)
	if err != nil {
		return err
	}
	err = d.Set("description", apiDHCP.Description)
	if err != nil {
		return err
	}
	err = d.Set("domainsearch", apiDHCP.DomainSearch)
	if err != nil {
		return err
	}
	err = d.Set("dnsservers", apiDHCP.DNSServers)
	if err != nil {
		return err
	}
	err = d.Set("ntpservers", apiDHCP.NTPServers)
	if err != nil {
		return err
	}
	err = d.Set("leasetime", apiDHCP.LeaseTime)
	if err != nil {
		return err
	}

	var standardOptions []map[string]interface{}
	var customOptions []map[string]interface{}

	for _, option := range apiDHCP.AdditionalOptions {
		opt := make(map[string]interface{})
		opt["code"] = option.Code
		opt["value"] = option.Value
		if !option.IsCustom {
			standardOptions = append(standardOptions, opt)
		} else {
			opt["type"] = option.Type
			customOptions = append(customOptions, opt)
		}
	}

	err = d.Set("standardtoption", standardOptions)
	if err != nil {
		return err
	}
	err = d.Set("customoption", customOptions)
	if err != nil {
		return err
	}

	return nil
}


func dataResourceExists(d *schema.ResourceData, m interface{}) (bool, error) {
	clientset := m.(*api.Clientset)

	id, _ := strconv.Atoi(d.Id())
	item, _ := clientset.DHCP().GetByID(id)

	if item == nil {
		return false, nil
	}
	if item.ID > 0 {
		return true, nil
	}

	return false, nil
}
