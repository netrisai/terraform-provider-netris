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
	"fmt"
	"net"

	"github.com/netrisai/netriswebapi/v1/types/site"
	"github.com/netrisai/netriswebapi/v1/types/tenant"
	api "github.com/netrisai/netriswebapi/v2"
	"github.com/netrisai/netriswebapi/v2/types/ipam"
)

func validateState(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if !(v == "active" || v == "disabled") {
		errs = append(errs, fmt.Errorf("'%s' must be active or disabled, got: %s", key, v))
	}
	return warns, errs
}

func regParser(valueMatch []string, subexpNames []string) map[string]string {
	result := make(map[string]string)
	for i, name := range subexpNames {
		if i != 0 && name != "" {
			result[name] = valueMatch[i]
		}
	}
	return result
}

func findTenantByIP(c *api.Clientset, ip string) (int, error) {
	tenantID := 0
	subnets, err := c.IPAM().Get()
	if err != nil {
		return tenantID, err
	}

	subnetChilds := []*ipam.IPAM{}
	for _, subnet := range subnets {
		subnetChilds = append(subnetChilds, subnet.Children...)
	}

	for _, subnet := range subnetChilds {
		ipAddr := net.ParseIP(ip)
		_, ipNet, err := net.ParseCIDR(subnet.Prefix)
		if err != nil {
			return tenantID, err
		}
		if ipNet.Contains(ipAddr) {
			return subnet.Tenant.ID, nil
		}
	}

	return tenantID, fmt.Errorf("There are no subnets for specified IP address %s", ip)
}

func findSiteByIP(c *api.Clientset, ip string) (int, error) {
	siteID := 0
	subnets, err := c.IPAM().Get()
	if err != nil {
		return siteID, err
	}

	subnetChilds := []*ipam.IPAM{}
	for _, subnet := range subnets {
		subnetChilds = append(subnetChilds, subnet.Children...)
	}

	for _, subnet := range subnetChilds {
		ipAddr := net.ParseIP(ip)
		_, ipNet, err := net.ParseCIDR(subnet.Prefix)
		if err != nil {
			return siteID, err
		}
		if ipNet.Contains(ipAddr) {
			if len(subnet.Sites) > 0 {
				return subnet.Sites[0].ID, nil
			}
		}
	}

	return siteID, fmt.Errorf("There are no sites  for specified IP address %s", ip)
}

func findTenantByName(c *api.Clientset, name string) (*tenant.Tenant, bool) {
	items, err := c.Tenant().Get()
	if err != nil {
		return nil, false
	}
	for _, item := range items {
		if item.Name == name {
			return item, true
		}
	}
	return nil, false
}

func findSiteByName(c *api.Clientset, name string) (*site.Site, bool) {
	items, err := c.Site().Get()
	if err != nil {
		return nil, false
	}
	for _, item := range items {
		if item.Name == name {
			return item, true
		}
	}
	return nil, false
}
