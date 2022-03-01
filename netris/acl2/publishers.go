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
	"strconv"
	"strings"

	"github.com/netrisai/netriswebapi/v1/types/acl2"
	api "github.com/netrisai/netriswebapi/v2"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func getPubInstances(d *schema.ResourceData, m interface{}) (instances []int) {
	publishersList := d.Get("publishers").([]interface{})
	if len(publishersList) == 0 {
		return instances
	}
	publishers := publishersList[0].(map[string]interface{})
	instancesList := publishers["instanceids"].([]interface{})
	for _, i := range instancesList {
		instances = append(instances, i.(int))
	}
	return instances
}

func getPubProtocols(d *schema.ResourceData, m interface{}) (protocols []acl2.PublisherWProtocol) {
	publishersList := d.Get("publishers").([]interface{})
	if len(publishersList) == 0 {
		return protocols
	}
	publishers := publishersList[0].(map[string]interface{})
	protocolList := publishers["protocol"].([]interface{})
	for _, p := range protocolList {
		protocol := p.(map[string]interface{})
		name := protocol["name"].(string)
		proto := protocol["protocol"].(string)
		portgroupid := protocol["portgroupid"].(int)
		port := protocol["port"].(string)
		if portgroupid > 0 {
			port = "1"
		}
		netrisProtocol := acl2.PublisherWProtocol{
			Description: name,
			Port:        port,
			PortGroupID: portgroupid,
			Proto:       proto,
		}
		if portgroupid > 0 {
			netrisProtocol.PortGroupID = portgroupid
		}
		protocols = append(protocols, netrisProtocol)
	}
	return protocols
}

func getPubPrefixes(d *schema.ResourceData, m interface{}) (pubPrefixes []acl2.PublisherWPrefix) {
	publishersList := d.Get("publishers").([]interface{})
	if len(publishersList) == 0 {
		return nil
	}
	publishers := publishersList[0].(map[string]interface{})
	prefixesList := publishers["prefixes"].([]interface{})
	prefixes := []string{}
	for _, p := range prefixesList {
		prefixes = append(prefixes, p.(string))
	}
	for _, p := range prefixes {
		prefix := strings.Split(p, "/")
		pubPrefixes = append(pubPrefixes, acl2.PublisherWPrefix{
			Prefix: prefix[0],
			Length: prefix[1],
		})
	}
	return pubPrefixes
}

func getPublishers(d *schema.ResourceData, m interface{}) (publishers *acl2.PublisherW) {
	id, _ := strconv.Atoi(d.Id())

	return &acl2.PublisherW{
		ID:        id,
		Instances: getPubInstances(d, m),
		Lbs:       []acl2.PublisherWLB{},
		Prefixes:  getPubPrefixes(d, m),
		TenantID:  d.Get("tenantid").(int),
	}
}

func getNetrisPubProtocols(d *schema.ResourceData, m interface{}) (protocols []acl2.Protoport) {
	clientset := m.(*api.Clientset)
	var acl *acl2.ACL2

	acls, err := clientset.ACL2().Get()
	if err != nil {
		return protocols
	}

	id, _ := strconv.Atoi(d.Id())

	for _, a := range acls {
		if a.ID == id {
			acl = a
		}
	}

	if acl.ID > 0 {
		return acl.Protoports
	}

	return protocols
}

func getNetrisPubPrefixes(d *schema.ResourceData, m interface{}) (prefixes []acl2.PublisherPrefix) {
	clientset := m.(*api.Clientset)
	var acl *acl2.ACL2

	acls, err := clientset.ACL2().Get()
	if err != nil {
		return prefixes
	}

	id, _ := strconv.Atoi(d.Id())

	for _, a := range acls {
		if a.ID == id {
			acl = a
		}
	}

	if acl.ID > 0 {
		return acl.PublisherPrefixes
	}

	return prefixes
}

func getNetrisSubPrefixes(d *schema.ResourceData, m interface{}) (prefixes []acl2.Prefix) {
	clientset := m.(*api.Clientset)
	var acl *acl2.ACL2

	acls, err := clientset.ACL2().Get()
	if err != nil {
		return prefixes
	}

	id, _ := strconv.Atoi(d.Id())

	for _, a := range acls {
		if a.ID == id {
			acl = a
		}
	}

	if acl.ID > 0 {
		return acl.Prefixes
	}

	return prefixes
}
