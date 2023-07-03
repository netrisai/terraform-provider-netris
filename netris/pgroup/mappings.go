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

package pgroup

type mapping struct {
	m map[string]map[string]string
}

var mappings = &mapping{
	m: map[string]map[string]string{
		"accounts": {
			"main":             "ACCOUNTS",
			"permissiongroups": "PERMISSION_GROUPS",
			"tenants":          "TENANTS",
			"userroles":        "USER_ROLES",
			"users":            "USERS",
		},
		"api": {
			"main": "GENERAL",
			"docs": "API_DOCS",
		},
		"net": {
			"main":          "NET",
			"ebgp":          "E_BGP",
			"ebgpobjects":   "E_BGP_OBJECTS",
			"ebgproutemaps": "E_BGP_ROUTE_MAPS",
			"inventory":     "INVENTORY",
			"ipam":          "SUBNETS",
			"lookinglass":   "LOOKING_GLASS",
			"nat":           "NAT",
			"routes":        "ROUTES",
			"sites":         "SITES",
			"switchports":   "SWITCH_PORTS",
			"topology":      "TOPOLOGY",
			"vpn":           "VPN",
		},
		"services": {
			"main":           "SERVICES",
			"acl":            "ACL",
			"aclportgroups":  "ACL_PORT_GROUPS",
			"acltwozero":     "ACL_2.0",
			"instances":      "INSTANCES",
			"l4loadbalancer": "L4_LOAD_BALANCER",
			"loadbalancer":   "LOAD_BALANCER",
			"vnet":           "CIRCUITS",
		},
		"settings": {
			"authentication": "AUTHENTICATION",
			"checks":         "MONITORING_CHECK_THRESHOLDS",
			"general":        "GS_GENERAL",
			"main":           "GLOBAL_SETTINGS",
			"whitelist":      "LOGIN_WHITELIST",
		},
	},
}

func (m *mapping) getMap() map[string]map[string]string {
	return m.m
}

var sectionNames = []string{
	"SERVICES",
	"INSTANCES",
	"CIRCUITS",
	"ACL",
	"ACL_2.0",
	"ACL_PORT_GROUPS",
	"LOAD_BALANCER",
	"L4_LOAD_BALANCER",
	"NET",
	"TOPOLOGY",
	"INVENTORY",
	"SWITCH_PORTS",
	"SITES",
	"E_BGP",
	"E_BGP_OBJECTS",
	"E_BGP_ROUTE_MAPS",
	"SUBNETS",
	"NAT",
	"VPN",
	"ROUTES",
	"LOOKING_GLASS",
	"ACCOUNTS",
	"USERS",
	"TENANTS",
	"USER_ROLES",
	"PERMISSION_GROUPS",
	"GLOBAL_SETTINGS",
	"GS_GENERAL",
	"LOGIN_WHITELIST",
	"AUTHENTICATION",
	"GENERAL",
	"API_DOCS",
	"MONITORING_CHECK_THRESHOLDS",
}
