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

package acl2

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/netrisai/netriswebapi/http"
	"github.com/netrisai/netriswebapi/v1/types/acl2"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	api "github.com/netrisai/netriswebapi/v2"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: "Creates and manages ACLs 2.0",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ACL 2.0 unique name",
			},
			"privacy": {
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
				Description: "Valid values are `public`, `private`, `hidden`. Public - Service is visible to all users and every user can subscribe instances and get access without approval. Private - Service is visible to all users, instances can be subscribed either by service owning tenant members or will require approval. Hidden - Service is not visible to any user except those who are part of tenant owning the service, instances can be subscribed only by service owning tenant members.",
			},
			"tenantid": {
				Required:    true,
				Type:        schema.TypeInt,
				Description: "ID of tenant. Users of this tenant will be permitted to manage this acl",
			},
			"state": {
				Optional:    true,
				Type:        schema.TypeString,
				Description: "State of the resource. Valid values are `enabled` or `disabled`",
			},
			"publishers": {
				Optional:    true,
				Type:        schema.TypeList,
				Description: "The block of publisher configurations",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"instanceids": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "List of Instances ID (ROH)",
							Elem: &schema.Schema{
								Type: schema.TypeInt,
							},
						},
						"lbvips": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "List of LB VIPs ID",
							Elem: &schema.Schema{
								Type: schema.TypeInt,
							},
						},
						"prefixes": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "List with prefixes",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"protocol": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "The block of protocol configurations",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Custom name for the current protocol",
									},
									"protocol": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Valid protocol. Possible values: `ip`, `tcp`, `udp`, `icmp`, `all`",
									},
									"port": {
										Optional:    true,
										Type:        schema.TypeString,
										Description: "Port number. Example `80`. Or protocol number when protocol == `ip`",
									},
									"portgroupid": {
										Optional:    true,
										Type:        schema.TypeInt,
										Description: "ID of Port Group. Use instead of port key",
									},
								},
							},
						},
					},
				},
			},
			"subscribers": {
				Optional:    true,
				Type:        schema.TypeList,
				Description: "The block of subscriber configurations",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"instanceids": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "List of Instances ID (ROH)",
							Elem: &schema.Schema{
								Type: schema.TypeInt,
							},
						},
						"prefix": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "List of prefixes",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"prefix": {
										Required:    true,
										Type:        schema.TypeString,
										Description: "Valid prefix",
									},
									"comment": {
										Optional:    true,
										Type:        schema.TypeString,
										Description: "Optional comment",
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
	// rs := reflect.ValueOf(d).Elem()
	// fs := rs.FieldByName("schema")
	// a := reflect.NewAt(fs.Type(), unsafe.Pointer(fs.UnsafeAddr())).Elem()
	// log.Println("[DEBUG] REFLECTTTTTT", a)

	name := d.Get("name").(string)
	privacy := d.Get("privacy").(string)
	tenantid := d.Get("tenantid").(int)

	acl2Create := &acl2.ACLw{
		Name:     name,
		Privacy:  privacy,
		TenantID: tenantid,
	}

	js, _ := json.Marshal(acl2Create)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.ACL2().Add(acl2Create)
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

	if reply.StatusCode != 200 && reply.StatusCode != 201 {
		return fmt.Errorf(string(reply.Data))
	}

	d.SetId(strconv.Itoa(idStruct.ID))

	err = changeStatus(d, m)
	if err != nil {
		return err
	}
	err = editPublishers(d, m)
	if err != nil {
		return err
	}
	err = editPubProtocols(d, m)
	if err != nil {
		return err
	}
	err = editSubscribers(d, m)
	if err != nil {
		return err
	}

	return nil
}

func changeStatus(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	id, _ := strconv.Atoi(d.Id())
	reply, err := clientset.ACL2().ChangeStatus(&acl2.ACLStatusW{
		ID:       id,
		Status:   d.Get("state").(string),
		TenantID: d.Get("tenantid").(int),
	})
	if err != nil {
		log.Println("[DEBUG]", err)
		return err
	}

	if reply.StatusCode != 200 {
		return fmt.Errorf(string(reply.Data))
	}

	return nil
}

func editPubProtocols(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	id, _ := strconv.Atoi(d.Id())

	netrisProtocols := getNetrisPubProtocols(d, m)
	protocolMap := make(map[string]acl2.Protoport)
	for _, p := range netrisProtocols {
		protocolMap[fmt.Sprintf("%s_%s", p.Proto, p.Port)] = p
	}

	publishersAdd := &acl2.PublisherW{
		ID:        id,
		Protocols: getPubProtocols(d, m),
		Lbs:       []acl2.PublisherWLB{},
		TenantID:  d.Get("tenantid").(int),
		Type:      "protocol",
	}

	for i, p := range publishersAdd.Protocols {
		if protocol, ok := protocolMap[fmt.Sprintf("%s_%s", p.Proto, p.Port)]; ok {
			publishersAdd.Protocols[i].ID = protocol.ID
		}
	}

	js, _ := json.Marshal(publishersAdd)
	log.Println("[DEBUG] Pub Protocols edit", string(js))

	reply, err := clientset.ACL2().EditPublishers(publishersAdd)
	if err != nil {
		log.Println("[DEBUG]", err)
		return err
	}

	if reply.StatusCode != 200 {
		return fmt.Errorf(string(reply.Data))
	}

	return nil
}

func editPublishers(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)
	fmt.Println(clientset)

	publishersAdd := getPublishers(d, m)

	netrisPrefixes := getNetrisPubPrefixes(d, m)
	prefixMap := make(map[string]acl2.PublisherPrefix)
	for _, p := range netrisPrefixes {
		prefixMap[fmt.Sprintf("%s/%s", p.Prefix, p.Length)] = p
	}
	for i, p := range publishersAdd.Prefixes {
		prefix := fmt.Sprintf("%s/%s", p.Prefix, p.Length)
		if prefix, ok := prefixMap[prefix]; ok {
			publishersAdd.Prefixes[i].ID = prefix.ID
		}
	}

	js, _ := json.Marshal(publishersAdd)
	log.Println("[DEBUG] Publisher edit", string(js))

	reply, err := clientset.ACL2().EditPublishers(publishersAdd)
	if err != nil {
		log.Println("[DEBUG]", err)
		return err
	}

	if reply.StatusCode != 200 {
		return fmt.Errorf(string(reply.Data))
	}

	return nil
}

func editSubscribers(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)
	fmt.Println(clientset)

	subscribers := getSubscribers(d, m)

	netrisPrefixes := getNetrisSubPrefixes(d, m)
	prefixMap := make(map[string]acl2.Prefix)
	for _, p := range netrisPrefixes {
		prefixMap[fmt.Sprintf("%s/%d", p.Prefix, p.Length)] = p
	}
	for i, p := range subscribers.Prefixes {
		prefix := fmt.Sprintf("%s/%s", p.Prefix, p.Length)
		if prefix, ok := prefixMap[prefix]; ok {
			subscribers.Prefixes[i].ID = prefix.ID
		}
	}

	js, _ := json.Marshal(subscribers)
	log.Println("[DEBUG] Subscriber edit", string(js))

	reply, err := clientset.ACL2().SubscribersEdit(subscribers)
	if err != nil {
		log.Println("[DEBUG]", err)
		return err
	}

	if reply.StatusCode != 200 {
		return fmt.Errorf(string(reply.Data))
	}

	return nil
}

func resourceRead(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)
	var acl *acl2.ACL2

	acls, err := clientset.ACL2().Get()
	if err != nil {
		return err
	}

	id, _ := strconv.Atoi(d.Id())

	for _, a := range acls {
		if a.ID == id {
			acl = a
		}
	}

	if !(acl != nil && acl.ID > 0) {
		return fmt.Errorf("Coudn't find acl2.0 %s", d.Get("name").(string))
	}

	d.SetId(strconv.Itoa(acl.ID))
	err = d.Set("name", acl.Name)
	if err != nil {
		return err
	}
	err = d.Set("privacy", acl.Privacy)
	if err != nil {
		return err
	}
	err = d.Set("state", acl.Status)
	if err != nil {
		return err
	}
	pubInstances := []int{}
	for _, i := range acl.PubInstances {
		pubInstances = append(pubInstances, i.ID)
	}
	pubPrefixes := []string{}
	for _, p := range acl.PublisherPrefixes {
		pubPrefixes = append(pubPrefixes, fmt.Sprintf("%s/%s", p.Prefix, p.Length))
	}
	lbVips := []int{}
	// for _, l := range acl.Lbs{

	// }

	var protocols []map[string]interface{}
	for _, p := range acl.Protoports {
		protocol := make(map[string]interface{})
		protocol["name"] = p.Description
		protocol["protocol"] = p.Proto
		portGroupID, _ := strconv.Atoi(p.PortGroupID)
		protocol["portgroupid"] = portGroupID
		if portGroupID == 0 {
			protocol["port"] = p.Port
		}
		protocols = append(protocols, protocol)
	}

	var publishers []map[string]interface{}
	publisher := make(map[string]interface{})
	publisher["instanceids"] = pubInstances
	publisher["lbvips"] = lbVips
	publisher["prefixes"] = pubPrefixes
	publisher["protocol"] = protocols
	publishers = append(publishers, publisher)
	err = d.Set("publishers", publishers)
	if err != nil {
		return err
	}

	var subscribers []map[string]interface{}
	subscriber := make(map[string]interface{})
	subInstances := []int{}
	for _, i := range acl.SubInstances {
		subInstances = append(subInstances, i.ID)
	}
	var subPrefixes []map[string]interface{}
	for _, p := range acl.Prefixes {
		prefix := make(map[string]interface{})
		prefix["prefix"] = fmt.Sprintf("%s/%d", p.Prefix, p.Length)
		prefix["comment"] = p.Comment
		subPrefixes = append(subPrefixes, prefix)
	}
	subscriber["instanceids"] = subInstances
	subscriber["prefix"] = subPrefixes
	subscribers = append(subscribers, subscriber)
	err = d.Set("subscribers", subscribers)
	if err != nil {
		return err
	}

	return nil
}

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	name := d.Get("name").(string)
	privacy := d.Get("privacy").(string)
	tenantid := d.Get("tenantid").(int)
	id, _ := strconv.Atoi(d.Id())

	acl2Update := &acl2.ACLw{
		ID:       id,
		Name:     name,
		Privacy:  privacy,
		TenantID: tenantid,
	}

	js, _ := json.Marshal(acl2Update)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.ACL2().Update(acl2Update)
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

	err = changeStatus(d, m)
	if err != nil {
		return err
	}
	err = editPublishers(d, m)
	if err != nil {
		return err
	}
	err = editPubProtocols(d, m)
	if err != nil {
		return err
	}
	err = editSubscribers(d, m)
	if err != nil {
		return err
	}

	return nil
}

func resourceDelete(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	id, _ := strconv.Atoi(d.Id())
	reply, err := clientset.ACL2().Delete(id)
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
	var acl *acl2.ACL2

	acls, err := clientset.ACL2().Get()
	if err != nil {
		return false, err
	}

	id, _ := strconv.Atoi(d.Id())

	for _, a := range acls {
		if a.ID == id {
			acl = a
		}
	}

	if acl != nil && acl.ID > 0 {
		return true, nil
	}

	return false, fmt.Errorf("Coudn't find acl2.0 %s", d.Get("name").(string))
}

func resourceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	return []*schema.ResourceData{d}, nil
}
