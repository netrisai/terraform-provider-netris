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

import (
	"regexp"
)

func parseGroups(s string) map[string]map[string][]string {
	rg := regexp.MustCompile(`(?P<key>[a-z0-9]+)\.?(?P<subkey>[a-z0-9]*):(?P<val>[a-z0-9-]+)`)
	sub := rg.SubexpNames()
	valueMatch := rg.FindAllStringSubmatch(s, -1)

	m := make(map[string]map[string][]string)

	for _, v := range valueMatch {
		a := regParser(v, sub)
		key := a["key"]
		subkey := a["subkey"]
		if subkey == "" {
			subkey = "main"
		}
		val := a["val"]

		if _, ok := m[key]; ok {
			m[key][subkey] = append(m[key][subkey], val)
		} else {
			m[key] = make(map[string][]string)
			m[key][subkey] = append(m[key][subkey], val)
		}
	}
	return m
}

func makeExceptionList(groupParameters map[string]map[string][]string, mappings map[string]map[string]string) (exceptHidden, exceptReadOnly map[string]int) {
	exceptHidden = make(map[string]int)
	exceptReadOnly = make(map[string]int)
	for key, keys := range groupParameters {
		for subkey, subkeys := range keys {
			for _, val := range subkeys {
				if subkey == "main" {
					if val == "view" {
						for s := range mappings[key] {
							exceptHidden[mappings[key][s]] = 1
						}
					} else if val == "edit" {
						for s := range mappings[key] {
							exceptHidden[mappings[key][s]] = 1
							exceptReadOnly[mappings[key][s]] = 1
						}
					}
				} else {
					if val == "view" {
						exceptHidden[mappings[key][subkey]] = 1
						exceptHidden[mappings[key]["main"]] = 1
					} else if val == "edit" {
						exceptHidden[mappings[key][subkey]] = 1
						exceptHidden[mappings[key]["main"]] = 1
						exceptReadOnly[mappings[key][subkey]] = 1
						exceptReadOnly[mappings[key]["main"]] = 1
					}
				}
			}
		}
	}
	return exceptHidden, exceptReadOnly
}

func makePermLists(exceptHidden, exceptReadOnly map[string]int, sections []string) (hiddenList, readOnlyList []string) {
	for _, name := range sectionNames {
		if _, ok := exceptHidden[name]; !ok {
			hiddenList = append(hiddenList, name)
		}
		if _, ok := exceptReadOnly[name]; !ok {
			readOnlyList = append(readOnlyList, name)
		}
	}

	return hiddenList, readOnlyList
}

func regParser(valueMatch []string, subexpNames []string) map[string]string {
	result := make(map[string]string)
	if len(subexpNames) == len(valueMatch) {
		for i, name := range subexpNames {
			if name != "" {
				result[name] = valueMatch[i]
			}
		}
	}
	return result
}
