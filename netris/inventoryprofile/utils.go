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

package inventoryprofile

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/netrisai/netriswebapi/v1/types/inventoryprofile"
	api "github.com/netrisai/netriswebapi/v2"
)

func findByID(id int, clientset *api.Clientset) (*inventoryprofile.Profile, bool) {
	list, err := clientset.InventoryProfile().Get()
	if err != nil {
		return nil, false
	}
	for _, profile := range list {
		if profile.ID == id {
			return profile, true
		}
	}
	return nil, false
}

func findByName(name string, clientset *api.Clientset) (*inventoryprofile.Profile, bool) {
	list, err := clientset.InventoryProfile().Get()
	if err != nil {
		return nil, false
	}
	for _, profile := range list {
		if profile.Name == name {
			return profile, true
		}
	}
	return nil, false
}

func unmarshalTimezone(s string) *inventoryprofile.Timezone {
	timezone := &inventoryprofile.Timezone{}
	_ = json.Unmarshal([]byte(s), timezone)
	return timezone
}

// normalizeTimezoneString trims spaces plus zero‑width/BOM runes controllers sometimes emit.
func normalizeTimezoneString(s string) string {
	return strings.TrimFunc(s, func(r rune) bool {
		return unicode.IsSpace(r) || r == '\u200b' || r == '\u200c' || r == '\u200d' || r == '\ufeff'
	})
}

// effectiveTimezoneForState maps the API timezone field to the value we store in Terraform.
// The controller may return JSON with only label, a JSON string, or a plain zone name.
func effectiveTimezoneForState(apiField string) string {
	s := normalizeTimezoneString(apiField)
	if s == "" {
		return ""
	}
	t := unmarshalTimezone(s)
	var out string
	if tc := normalizeTimezoneString(strings.TrimSpace(t.TzCode)); tc != "" {
		out = tc
	} else if lb := normalizeTimezoneString(strings.TrimSpace(t.Label)); lb != "" {
		out = lb
	} else {
		var decoded string
		if err := json.Unmarshal([]byte(s), &decoded); err == nil {
			out = normalizeTimezoneString(decoded)
		}
		// JSON object (e.g. {"label":"","offset":"","tzCode":""}) with no usable fields — treat as unset.
		if out == "" && strings.HasPrefix(s, "{") {
			return ""
		}
		if out == "" {
			out = s
		}
	}
	out = normalizeTimezoneString(out)
	if out == "" {
		return ""
	}
	return out
}

// parseNetQSettings reads the netqsettings block and builds the NetQProps payload.
func parseNetQSettings(d *schema.ResourceData) (inventoryprofile.NetQProps, error) {
	netq := inventoryprofile.NetQProps{ServerAddrs: []string{}}
	netqList := d.Get("netqsettings").(*schema.Set).List()
	if len(netqList) == 0 {
		// No netqsettings block defined: NetQ is disabled.
		return netq, nil
	}
	if len(netqList) > 1 {
		return netq, fmt.Errorf("please specify only one netqsettings")
	}
	netqtmp, ok := netqList[0].(map[string]interface{})
	if !ok {
		return netq, nil
	}
	// A netqsettings block defaults to enabled (schema Default is true); an
	// explicit enabled = false keeps the config but turns NetQ off.
	netq.Enabled = true
	if v, ok := netqtmp["enabled"].(bool); ok {
		netq.Enabled = v
	}
	if rawAddrs, ok := netqtmp["server_addrs"].([]interface{}); ok {
		for _, s := range rawAddrs {
			netq.ServerAddrs = append(netq.ServerAddrs, s.(string))
		}
	}
	if port, ok := netqtmp["server_port"].(int); ok {
		netq.ServerPort = int32(port)
	}
	return netq, nil
}

// getStringFromMap returns a trimmed string for key from m, or empty if missing or not a string.
func getStringFromMap(key string, m map[string]interface{}) string {
	if m == nil {
		return ""
	}
	val, ok := m[key]
	if !ok || val == nil {
		return ""
	}
	s, ok := val.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(s)
}
