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

package l4lb

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/netrisai/netriswebapi/http"
	"github.com/netrisai/netriswebapi/v1/types/l4lb"

	api "github.com/netrisai/netriswebapi/v2"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the resource, also acts as it's unique ID",
			},
			"tenantid": {
				Optional: true,
				Type:     schema.TypeString,
				ForceNew: true,
			},
			"siteid": {
				Required: true,
				Type:     schema.TypeString,
				ForceNew: true,
			},
			"state": {
				Optional:     true,
				Default:      "active",
				ValidateFunc: validateState,
				Type:         schema.TypeString,
			},
			"protocol": {
				Optional: true,
				Type:     schema.TypeString,
			},
			"frontend": {
				Optional: true,
				Type:     schema.TypeString,
			},
			"port": {
				Optional: true,
				Type:     schema.TypeInt,
			},
			"backend": {
				Optional: true,
				Type:     schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"check": {
				Optional:    true,
				Type:        schema.TypeMap,
				Description: "Check Options",
				Elem: &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
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

	bReg := regexp.MustCompile(`^(?P<ip>(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])):(?P<port>([1-9]|[1-9][0-9]{1,3}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-4]))$`)

	tenantID := d.Get("tenantid").(int)
	siteID := d.Get("site").(int)

	var state string
	var timeout string
	proto := "TCP"

	lbBackends := []l4lb.LBAddBackend{}

	l4lbMetaBackends := d.Get("backend").([]interface{})
	ipForTenant := ""

	for _, b := range l4lbMetaBackends {
		backend := b.(string)
		valueMatch := bReg.FindStringSubmatch(string(backend))
		result := regParser(valueMatch, bReg.SubexpNames())
		port, err := strconv.Atoi(result["port"])
		if err != nil {
			return err
		}
		ipForTenant = result["ip"]
		lbBackends = append(lbBackends, l4lb.LBAddBackend{
			IP:   result["ip"],
			Port: port,
		})
	}

	if tenantID == 0 {
		tenantid, err := findTenantByIP(clientset, ipForTenant)
		if err != nil {
			return err
		}
		tenantID = tenantid
	}

	status := d.Get("state").(string)
	if status == "" || status == "active" {
		state = "enable"
	} else {
		state = status
	}

	protocol := strings.ToUpper(d.Get("protocol").(string))
	if protocol != "" {
		proto = protocol
	}

	check := d.Get("check").(map[string]interface{})
	checkType := check["type"].(string)
	checkTimeout, _ := strconv.Atoi(check["timeout"].(string))
	checkRequestPath, _ := check["requestPath"].(string)

	healthCheck := "None"

	if proto == "TCP" {
		if checkTimeout == 0 {
			timeout = "2000"
		} else {
			timeout = strconv.Itoa(checkTimeout)
		}

		if checkType == "tcp" || checkType == "" {
			healthCheck = "HTTP"
		} else {
			healthCheck = "TCP"
		}
	}

	automatic := false
	frontendIP := d.Get("frontend").(string)
	if frontendIP == "" {
		automatic = true
		frontendIP = ""
	}

	l4lbAdd := &l4lb.LoadBalancerAdd{
		Name:        d.Get("name").(string),
		Tenant:      tenantID,
		SiteID:      siteID,
		Automatic:   automatic,
		Protocol:    proto,
		IP:          frontendIP,
		Port:        d.Get("port").(int),
		Status:      state,
		RequestPath: checkRequestPath,
		Timeout:     timeout,
		Backend:     lbBackends,
	}

	if healthCheck != "" {
		l4lbAdd.HealthCheck = healthCheck
	}

	js, _ := json.Marshal(l4lbAdd)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.L4LB().Add(l4lbAdd)
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

	var id int
	if automatic {
		err = http.Decode(data.Data, &idStruct)
		if err != nil {
			log.Println("[DEBUG]", err)
			return err
		}
		id = idStruct.ID
	} else {
		err = http.Decode(data.Data, &id)
		if err != nil {
			log.Println("[DEBUG]", err)
			return err
		}
	}

	log.Println("[DEBUG] ID:", id)

	if reply.StatusCode != 200 {
		return fmt.Errorf(string(reply.Data))
	}

	d.SetId(strconv.Itoa(id))
	return nil
}

func resourceRead(d *schema.ResourceData, m interface{}) error {
	// clientset := m.(*api.Clientset)

	return nil
}

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	bReg := regexp.MustCompile(`^(?P<ip>(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])):(?P<port>([1-9]|[1-9][0-9]{1,3}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-4]))$`)

	var (
		state   string
		timeout string
		proto   string = "tcp"
	)

	lbBackends := []l4lb.LBBackend{}

	l4lbMetaBackends := d.Get("backend").([]interface{})

	for _, b := range l4lbMetaBackends {
		backend := b.(string)
		valueMatch := bReg.FindStringSubmatch(string(backend))
		result := regParser(valueMatch, bReg.SubexpNames())
		lbBackends = append(lbBackends, l4lb.LBBackend{
			IP:   result["ip"],
			Port: result["port"],
		})
	}

	status := d.Get("state").(string)
	if status == "" || status == "active" {
		state = "enable"
	} else {
		state = status
	}

	protocol := d.Get("protocol").(string)
	if protocol != "" {
		proto = protocol
	}

	check := d.Get("check").(map[string]interface{})
	checkType := check["type"].(string)
	checkTimeout := check["timeout"].(int)
	checkRequestPath, _ := check["requestPath"].(string)

	healthCheck := "None"

	if proto == "tcp" {
		if checkTimeout == 0 {
			timeout = "2000"
		} else {
			timeout = strconv.Itoa(checkTimeout)
		}

		if checkType == "tcp" || checkType == "" {
			healthCheck = "HTTP"
		} else {
			healthCheck = "TCP"
		}
	}

	automatic := false
	frontendIP := d.Get("frontend").(string)
	if frontendIP == "" {
		automatic = true
		frontendIP = ""
	}

	id, _ := strconv.Atoi(d.Id())
	l4lbUpdate := &l4lb.LoadBalancerUpdate{
		ID:          id,
		Name:        d.Get("name").(string),
		Automatic:   automatic,
		Protocol:    proto,
		IP:          frontendIP,
		Port:        d.Get("port").(int),
		Status:      state,
		RequestPath: checkRequestPath,
		Timeout:     timeout,
		BackendIPs:  lbBackends,
	}

	if healthCheck != "" {
		l4lbUpdate.HealthCheck = healthCheck
	}

	js, _ := json.Marshal(l4lbUpdate)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.L4LB().Update(l4lbUpdate)
	if err != nil {
		log.Println("[DEBUG]", err)
		return err
	}

	if reply.StatusCode != 200 {
		return fmt.Errorf(string(reply.Data))
	}

	return nil
}

func resourceDelete(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	id, _ := strconv.Atoi(d.Id())
	reply, err := clientset.L4LB().Delete(id)
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
	l4lbs, err := clientset.L4LB().Get()
	if err != nil {
		return false, err
	}
	for _, lb := range l4lbs {
		if lb.ID == id {
			return true, nil
		}
	}

	return true, nil
}

func resourceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	clientset := m.(*api.Clientset)

	vnets, _ := clientset.VNet().Get()
	name := d.Id()
	for _, vnet := range vnets {
		if vnet.Name == name {
			d.SetId(strconv.Itoa(vnet.ID))
			return []*schema.ResourceData{d}, nil
		}
	}

	return []*schema.ResourceData{d}, nil
}
