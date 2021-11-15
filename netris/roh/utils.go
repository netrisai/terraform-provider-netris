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

package roh

import (
	"regexp"

	"github.com/netrisai/netriswebapi/v2/types/roh"
)

func parsePrefixList(prefix string) roh.InboundPrefixW {
	re := regexp.MustCompile(`(?P<action>permit|deny) +(?P<ip>(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\/\b([0-9]|[12][0-9]|3[0-2])\b) +(?P<cond>(le|ge) +\b([0-9]|[12][0-9]|3[0-2])\b)`)
	sub := re.SubexpNames()
	valueMatch := re.FindStringSubmatch(prefix)
	v := regParser(valueMatch, sub)
	return roh.InboundPrefixW{
		Action:    v["action"],
		Condition: v["cond"],
		Subnet:    v["ip"],
	}
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
