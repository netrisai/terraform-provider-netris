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
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/netrisai/netriswebapi/http"
	api "github.com/netrisai/netriswebapi/v2"
	"github.com/netrisai/netriswebapi/v2/types/dhcp"
)

func Resource() *schema.Resource {
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

	dnsServers := []string{}
	dnsList := d.Get("dnsservers").(*schema.Set).List()
	for _, dns := range dnsList {
		dnsServers = append(dnsServers, dns.(string))
	}

	ntpServers := []string{}
	ntpList := d.Get("ntpservers").(*schema.Set).List()
	for _, dns := range ntpList {
		ntpServers = append(ntpServers, dns.(string))
	}

	options := []dhcp.AdditionalOption{}
	standardOptions := d.Get("standardtoption").(*schema.Set).List()
	customOptions := d.Get("customoption").(*schema.Set).List()

	for _, o := range standardOptions {
		opt := o.(map[string]interface{})
		options = append(options, dhcp.AdditionalOption{
			Code:  opt["code"].(int),
			Type:  dhcpStandardOptionTypes[opt["code"].(int)],
			Value: opt["value"].(string),
		})
	}

	for _, o := range customOptions {
		opt := o.(map[string]interface{})
		options = append(options, dhcp.AdditionalOption{
			Code:  opt["code"].(int),
			Type:  opt["type"].(string),
			Value: opt["value"].(string),
		})
	}

	dhcpAdd := &dhcp.DHCPw{
		Name:              d.Get("name").(string),
		Description:       d.Get("description").(string),
		DomainSearch:      d.Get("domainsearch").(string),
		DNSServers:        dnsServers,
		NTPServers:        ntpServers,
		LeaseTime:         d.Get("leasetime").(int),
		AdditionalOptions: options,
	}

	js, _ := json.Marshal(dhcpAdd)
	log.Println("[DEBUG] dhcpAdd", string(js))

	reply, err := clientset.DHCP().Add(dhcpAdd)
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
	apiDHCP, err := clientset.DHCP().GetByID(id)
	if err != nil {
		return nil
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

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	dhcpID, _ := strconv.Atoi(d.Id())

	dnsServers := []string{}
	dnsList := d.Get("dnsservers").(*schema.Set).List()
	for _, dns := range dnsList {
		dnsServers = append(dnsServers, dns.(string))
	}

	ntpServers := []string{}
	ntpList := d.Get("ntpservers").(*schema.Set).List()
	for _, dns := range ntpList {
		ntpServers = append(ntpServers, dns.(string))
	}

	options := []dhcp.AdditionalOption{}
	standardOptions := d.Get("standardtoption").(*schema.Set).List()
	customOptions := d.Get("customoption").(*schema.Set).List()

	for _, o := range standardOptions {
		opt := o.(map[string]interface{})
		options = append(options, dhcp.AdditionalOption{
			Code:  opt["code"].(int),
			Type:  dhcpStandardOptionTypes[opt["code"].(int)],
			Value: opt["value"].(string),
		})
	}

	for _, o := range customOptions {
		opt := o.(map[string]interface{})
		options = append(options, dhcp.AdditionalOption{
			Code:  opt["code"].(int),
			Type:  opt["type"].(string),
			Value: opt["value"].(string),
		})
	}

	dhcpUpdate := &dhcp.DHCPw{
		Name:              d.Get("name").(string),
		Description:       d.Get("description").(string),
		DomainSearch:      d.Get("domainsearch").(string),
		DNSServers:        dnsServers,
		NTPServers:        ntpServers,
		LeaseTime:         d.Get("leasetime").(int),
		AdditionalOptions: options,
	}

	js, _ := json.Marshal(dhcpUpdate)
	log.Println("[DEBUG] bgpUpdate", string(js))

	reply, err := clientset.DHCP().Update(dhcpID, dhcpUpdate)
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

func resourceExists(d *schema.ResourceData, m interface{}) (bool, error) {
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

func resourceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	clientset := m.(*api.Clientset)

	items, _ := clientset.DHCP().Get()
	name := d.Id()
	for _, item := range items {
		if item.Name == name {
			d.SetId(strconv.Itoa(item.ID))
			return []*schema.ResourceData{d}, nil
		}
	}

	return []*schema.ResourceData{d}, nil
}

func resourceDelete(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	id, _ := strconv.Atoi(d.Id())
	reply, err := clientset.DHCP().Delete(id)
	if err != nil {
		return err
	}

	if reply.StatusCode != 200 {
		return fmt.Errorf(string(reply.Data))
	}

	d.SetId("")
	return nil
}
