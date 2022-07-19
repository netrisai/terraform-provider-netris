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
	"github.com/netrisai/netriswebapi/v2/types/l4lb"

	api "github.com/netrisai/netriswebapi/v2"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: "Creates and manages L4LBs",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the resource",
			},
			"tenantid": {
				Optional:    true,
				Type:        schema.TypeInt,
				ForceNew:    true,
				Description: "ID of tenant. Users of this tenant will be permitted to edit this unit.",
			},
			"siteid": {
				Optional:    true,
				Type:        schema.TypeInt,
				ForceNew:    true,
				Description: "The site ID. Resources defined in the selected site will be permitted to be used as backed entries for this L4 Load Balancer service.",
			},
			"state": {
				Optional:     true,
				Default:      "active",
				ValidateFunc: validateState,
				Type:         schema.TypeString,
				Description:  "Administrative status. Possible values: `active` or `disable`. Default value is `active`",
			},
			"protocol": {
				Optional:    true,
				Type:        schema.TypeString,
				Description: "Protocol. Possible values: `tcp` or `udp`",
			},
			"frontend": {
				Optional:    true,
				Type:        schema.TypeString,
				Description: "L4LB frontend IP. If not specified, will be assigned automatically from subnets with relevant purpose.",
			},
			"port": {
				Optional:    true,
				Type:        schema.TypeInt,
				Description: "L4LB frontend port to be exposed",
			},
			"backend": {
				Optional: true,
				Type:     schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "List of backends. Valid value is `ip`:`port` Example `[\"192.0.2.100:443\", \"192.0.2.101:443\"]`",
			},
			"check": {
				Optional:    true,
				Type:        schema.TypeMap,
				Description: "A health check determines whether instances in the target pool are healthy. If protocol == `udp` then check.type should be `none`",
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
	siteID := d.Get("siteid").(int)

	var state string
	var timeout string
	proto := "TCP"

	lbBackends := []l4lb.LBAddBackend{}

	l4lbMetaBackends := d.Get("backend").(*schema.Set).List()
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

	if siteID == 0 {
		siteid, err := findSiteByIP(clientset, ipForTenant)
		if err != nil {
			return err
		}
		siteID = siteid
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

	checkType := ""
	checkTimeout := 0
	checkRequestPath := ""

	check := d.Get("check").(map[string]interface{})
	if v, ok := check["type"]; ok {
		checkType = v.(string)
	}
	if v, ok := check["timeout"]; ok {
		checkTimeout, _ = strconv.Atoi(v.(string))
	}
	if v, ok := check["requestPath"]; ok {
		checkRequestPath = v.(string)
	}

	healthCheck := "None"

	if proto == "TCP" {
		if checkTimeout == 0 {
			timeout = "2000"
		} else {
			timeout = strconv.Itoa(checkTimeout)
		}

		if checkType == "tcp" || checkType == "" {
			healthCheck = "TCP"
		} else {
			healthCheck = "HTTP"
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
		Protocol:    strings.ToUpper(proto),
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
	clientset := m.(*api.Clientset)

	id, _ := strconv.Atoi(d.Id())
	var l4lb *l4lb.LoadBalancer

	l4lbs, err := clientset.L4LB().Get()
	if err != nil {
		return nil
	}
	for _, lb := range l4lbs {
		if lb.ID == id {
			l4lb = lb
			break
		}
	}

	if !(l4lb != nil && l4lb.ID > 0) {
		return fmt.Errorf("Coudn't find l4lb with id '%d'", id)
	}

	d.SetId(strconv.Itoa(l4lb.ID))
	err = d.Set("name", l4lb.Name)
	if err != nil {
		return err
	}
	err = d.Set("tenantid", l4lb.TenantID)
	if err != nil {
		return err
	}
	err = d.Set("siteid", l4lb.SiteID)
	if err != nil {
		return err
	}
	state := "disable"
	if l4lb.Status == "enable" {
		state = "active"
	}
	err = d.Set("state", state)
	if err != nil {
		return err
	}
	err = d.Set("protocol", strings.ToLower(l4lb.Protocol))
	if err != nil {
		return err
	}
	err = d.Set("frontend", l4lb.IP)
	if err != nil {
		return err
	}
	err = d.Set("port", l4lb.Port)
	if err != nil {
		return err
	}

	check := make(map[string]interface{})
	lbCheckType := "None"
	lbCheckTimeout := ""
	if l4lb.HealthCheck.HTTP.Timeout != "" {
		lbCheckType = "http"
		check["requestPath"] = l4lb.HealthCheck.HTTP.RequestPath
		lbCheckTimeout = l4lb.HealthCheck.HTTP.Timeout
	}
	if l4lb.HealthCheck.TCP.Timeout != "" {
		lbCheckType = "tcp"
		lbCheckTimeout = l4lb.HealthCheck.TCP.Timeout
	}
	check["type"] = lbCheckType
	check["timeout"] = lbCheckTimeout
	err = d.Set("check", check)
	if err != nil {
		return err
	}

	backends := []interface{}{}
	for _, b := range l4lb.BackendIPs {
		backends = append(backends, fmt.Sprintf("%s:%s", b.IP, b.Port))
	}
	err = d.Set("backend", backends)
	if err != nil {
		return err
	}

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

	l4lbMetaBackends := d.Get("backend").(*schema.Set).List()
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

	checkType := ""
	checkTimeout := 0
	checkRequestPath := ""

	check := d.Get("check").(map[string]interface{})
	if v, ok := check["type"]; ok {
		checkType = v.(string)
	}
	if v, ok := check["timeout"]; ok {
		checkTimeout, _ = strconv.Atoi(v.(string))
	}
	if v, ok := check["requestPath"]; ok {
		checkRequestPath = v.(string)
	}

	healthCheck := "None"

	if proto == "tcp" {
		if checkTimeout == 0 {
			timeout = "2000"
		} else {
			timeout = strconv.Itoa(checkTimeout)
		}

		if checkType == "tcp" || checkType == "" {
			healthCheck = "TCP"
		} else {
			healthCheck = "HTTP"
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
		Name:        d.Get("name").(string),
		TenantID:    d.Get("tenantid").(int),
		SiteID:      d.Get("siteid").(int),
		Automatic:   automatic,
		Protocol:    strings.ToUpper(proto),
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

	reply, err := clientset.L4LB().Update(id, l4lbUpdate)
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

	return false, nil
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
