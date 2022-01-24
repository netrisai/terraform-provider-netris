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

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func getSubInstances(d *schema.ResourceData, m interface{}) (instances []int) {
	subscribersList := d.Get("subscribers").([]interface{})
	if len(subscribersList) == 0 {
		return instances
	}
	subscribers := subscribersList[0].(map[string]interface{})
	instancesList := subscribers["instanceids"].([]interface{})
	for _, i := range instancesList {
		instances = append(instances, i.(int))
	}
	return instances
}

func getSubPrefixes(d *schema.ResourceData, m interface{}) (subPrefixes []acl2.SubscriberWPrefix) {
	subscribersList := d.Get("subscribers").([]interface{})
	if len(subscribersList) == 0 {
		return subPrefixes
	}
	subscribers := subscribersList[0].(map[string]interface{})
	prefixesList := subscribers["prefix"].([]interface{})
	for _, p := range prefixesList {
		prefixMap := p.(map[string]interface{})
		prefix := strings.Split(prefixMap["prefix"].(string), "/")
		comment := prefixMap["comment"].(string)
		subPrefixes = append(subPrefixes, acl2.SubscriberWPrefix{
			Comment: comment,
			Prefix:  prefix[0],
			Length:  prefix[1],
		})
	}
	return subPrefixes
}

func getSubscribers(d *schema.ResourceData, m interface{}) (subscribers *acl2.SubscriberW) {
	id, _ := strconv.Atoi(d.Id())

	return &acl2.SubscriberW{
		ID:        id,
		Instances: getSubInstances(d, m),
		Prefixes:  getSubPrefixes(d, m),
		TenantID:  d.Get("tenantid").(int),
	}
}
