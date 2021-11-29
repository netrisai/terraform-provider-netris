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

package routemap

import "strings"

func getType(t string) string {
	return typeMappings[t]
}

func typeExist(t string) bool {
	_, ok := typeMappings[t]
	return ok
}

func getTypesListString() string {
	list := []string{}
	for k := range typeMappings {
		list = append(list, k)
	}
	return strings.Join(list, ", ")
}

var typeMappings = map[string]string{
	"as_path":            "object",
	"community":          "object",
	"extended_community": "object",
	"large_community":    "object",
	"ipv4_prefix_list":   "object",
	"ipv4_next_hop":      "object",
	"route_source":       "object",
	"ipv6_prefix_list":   "object",
	"ipv6_next_hop":      "string",
	"local_preference":   "string",
	"med":                "string",
	"origin":             "string",
	"route_tag":          "string",
}
