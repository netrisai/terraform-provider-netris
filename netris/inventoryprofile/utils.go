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
	"sort"
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

// parseSyslogDestinations reads the syslog_destinations block and builds the SyslogDestinations payload.
func parseSyslogDestinations(d *schema.ResourceData) (inventoryprofile.SyslogDestinations, error) {
	syslog := inventoryprofile.SyslogDestinations{Servers: []inventoryprofile.SyslogServer{}}
	syslogList := d.Get("syslog_destinations").([]interface{})
	if len(syslogList) == 0 {
		return syslog, nil
	}
	if len(syslogList) > 1 {
		return syslog, fmt.Errorf("please specify only one syslog_destinations")
	}
	syslogtmp, ok := syslogList[0].(map[string]interface{})
	if !ok {
		return syslog, nil
	}
	if v, ok := syslogtmp["enabled"].(bool); ok {
		syslog.Enabled = v
	}
	if v, ok := syslogtmp["use_rfc5424"].(bool); ok {
		syslog.UseRfc5424 = v
	}
	if rawServers, ok := syslogtmp["servers"].([]interface{}); ok {
		for _, rs := range rawServers {
			server, ok := rs.(map[string]interface{})
			if !ok {
				continue
			}
			port, _ := server["port"].(int)
			syslog.Servers = append(syslog.Servers, inventoryprofile.SyslogServer{
				Host:     getStringFromMap("host", server),
				Port:     int32(port),
				Protocol: getStringFromMap("protocol", server),
				Severity: getStringFromMap("severity", server),
			})
		}
	}
	return syslog, nil
}

// syslogDestinationsToMap converts the API's SyslogDestinations into the
// map shape expected by the syslog_destinations Terraform block.
func syslogDestinationsToMap(syslog inventoryprofile.SyslogDestinations) map[string]interface{} {
	var servers []map[string]interface{}
	for _, s := range syslog.Servers {
		servers = append(servers, map[string]interface{}{
			"host":     s.Host,
			"port":     int(s.Port),
			"protocol": s.Protocol,
			"severity": s.Severity,
		})
	}
	return map[string]interface{}{
		"enabled":     syslog.Enabled,
		"use_rfc5424": syslog.UseRfc5424,
		"servers":     servers,
	}
}

// dataGetter is satisfied by both *schema.ResourceData and *schema.ResourceDiff,
// letting parseAAA validate the aaa block during CustomizeDiff (so `terraform
// plan` catches errors) and reuse the exact same logic to build the API
// payload during Create/Update.
type dataGetter interface {
	Get(key string) interface{}
}

// defaultAAAProps is the backward-compatible default applied when no aaa
// block is configured: local-only authentication, matching every profile's
// behavior before this feature existed (HLD §6.2 "aaa Default").
func defaultAAAProps() inventoryprofile.AAAProps {
	return inventoryprofile.AAAProps{
		AuthOrder: []string{"local"},
		Radius:    inventoryprofile.RadiusProps{PriorityServers: []inventoryprofile.RadiusServer{}},
		Local:     inventoryprofile.LocalAuthProps{Enabled: true},
	}
}

// parseAAA reads the aaa block and builds the AAAProps payload, applying the
// same validation the controller enforces (R10-R12 and the authorder/enabled
// consistency rules in HLD §6.3) so a misconfiguration fails at `terraform
// plan` instead of surfacing as an opaque API error at apply time.
func parseAAA(d dataGetter) (inventoryprofile.AAAProps, error) {
	aaaList, ok := d.Get("aaa").([]interface{})
	if !ok || len(aaaList) == 0 {
		return defaultAAAProps(), nil
	}
	aaatmp, ok := aaaList[0].(map[string]interface{})
	if !ok {
		return defaultAAAProps(), nil
	}

	authOrder := []string{}
	seenOrder := map[string]bool{}
	if raw, ok := aaatmp["authorder"].([]interface{}); ok {
		for _, s := range raw {
			method := s.(string)
			if seenOrder[method] {
				return inventoryprofile.AAAProps{}, fmt.Errorf("authentication order cannot contain duplicate methods: %q", method)
			}
			seenOrder[method] = true
			authOrder = append(authOrder, method)
		}
	}
	if len(authOrder) == 0 {
		return inventoryprofile.AAAProps{}, fmt.Errorf("select at least one authentication method (local or radius)")
	}

	radiusList, _ := aaatmp["radius"].([]interface{})
	var radiustmp map[string]interface{}
	if len(radiusList) > 0 {
		radiustmp, _ = radiusList[0].(map[string]interface{})
	}
	radiusEnabled := getBoolFromMap("enabled", radiustmp, false)

	servers := []inventoryprofile.RadiusServer{}
	seenHostPort := map[string]bool{}
	seenPriority := map[int]bool{}
	if rawServers, ok := radiustmp["server"].([]interface{}); ok {
		for _, rs := range rawServers {
			server, ok := rs.(map[string]interface{})
			if !ok {
				continue
			}
			host := getStringFromMap("host", server)
			port, _ := server["port"].(int)
			priority, _ := server["priority"].(int)

			hostPort := fmt.Sprintf("%s:%d", host, port)
			if seenHostPort[hostPort] {
				return inventoryprofile.AAAProps{}, fmt.Errorf("%q is already used by another RADIUS server in this profile; each RADIUS server must be unique", hostPort)
			}
			seenHostPort[hostPort] = true

			if seenPriority[priority] {
				return inventoryprofile.AAAProps{}, fmt.Errorf("duplicate RADIUS server priority %d; each RADIUS server must have a unique priority", priority)
			}
			seenPriority[priority] = true

			servers = append(servers, inventoryprofile.RadiusServer{
				Host:     host,
				Port:     int32(port),
				Priority: int32(priority),
				AuthType: getStringFromMap("authtype", server),
				Secret:   getStringFromMap("secret", server),
			})
		}
	}
	sort.Slice(servers, func(i, j int) bool { return servers[i].Priority < servers[j].Priority })

	if radiusEnabled && len(servers) == 0 {
		return inventoryprofile.AAAProps{}, fmt.Errorf("at least one radius server entry is required when radius is enabled")
	}
	if len(servers) > 8 {
		return inventoryprofile.AAAProps{}, fmt.Errorf("maximum 8 radius servers are allowed per inventory profile")
	}

	localList, _ := aaatmp["local"].([]interface{})
	var localtmp map[string]interface{}
	if len(localList) > 0 {
		localtmp, _ = localList[0].(map[string]interface{})
	}
	localEnabled := getBoolFromMap("enabled", localtmp, true)

	if radiusEnabled != seenOrder["radius"] {
		if seenOrder["radius"] {
			return inventoryprofile.AAAProps{}, fmt.Errorf("radius auth must be enabled when it is selected in authorder")
		}
		return inventoryprofile.AAAProps{}, fmt.Errorf("radius auth must be disabled when it is not selected in authorder")
	}
	if localEnabled != seenOrder["local"] {
		if seenOrder["local"] {
			return inventoryprofile.AAAProps{}, fmt.Errorf("local auth must be enabled when it is selected in authorder")
		}
		return inventoryprofile.AAAProps{}, fmt.Errorf("local auth must be disabled when it is not selected in authorder")
	}

	return inventoryprofile.AAAProps{
		AuthOrder: authOrder,
		Radius: inventoryprofile.RadiusProps{
			Enabled:         radiusEnabled,
			PriorityServers: servers,
		},
		Local: inventoryprofile.LocalAuthProps{Enabled: localEnabled},
	}, nil
}

// aaaToMap converts the API's AAAProps into the map shape expected by the aaa
// Terraform block. existingServers, keyed by "host:port" from the prior
// state/config, supplies the secret and priority for servers the API already
// knows about: the API never returns the secret in cleartext (write-only,
// matching the NOS Admin Password pattern), and priority is Terraform-only
// (the API models order via the priorityServers list position, not a stored
// field), so both must be preserved from what Terraform last wrote rather
// than reconstructed from the read.
func aaaToMap(aaa inventoryprofile.AAAProps, existingServers map[string]map[string]interface{}) map[string]interface{} {
	var servers []map[string]interface{}
	for i, s := range aaa.Radius.PriorityServers {
		hostPort := fmt.Sprintf("%s:%d", s.Host, s.Port)
		priority := i + 1
		secret := ""
		if existing, ok := existingServers[hostPort]; ok {
			if p, ok := existing["priority"].(int); ok {
				priority = p
			}
			if sec, ok := existing["secret"].(string); ok && sec != "" {
				secret = sec
			}
		}
		servers = append(servers, map[string]interface{}{
			"host":     s.Host,
			"port":     int(s.Port),
			"priority": priority,
			"authtype": s.AuthType,
			"secret":   secret,
		})
	}

	return map[string]interface{}{
		"authorder": aaa.AuthOrder,
		"radius": []map[string]interface{}{
			{
				"enabled": aaa.Radius.Enabled,
				"server":  servers,
			},
		},
		"local": []map[string]interface{}{
			{"enabled": aaa.Local.Enabled},
		},
	}
}

// existingRadiusServersByHostPort indexes the aaa.radius.server blocks
// currently in state/config by "host:port", for aaaToMap to preserve
// Terraform-only or write-only fields across a read.
func existingRadiusServersByHostPort(d dataGetter) map[string]map[string]interface{} {
	out := map[string]map[string]interface{}{}
	aaaList, ok := d.Get("aaa").([]interface{})
	if !ok || len(aaaList) == 0 {
		return out
	}
	aaatmp, ok := aaaList[0].(map[string]interface{})
	if !ok {
		return out
	}
	radiusList, _ := aaatmp["radius"].([]interface{})
	if len(radiusList) == 0 {
		return out
	}
	radiustmp, ok := radiusList[0].(map[string]interface{})
	if !ok {
		return out
	}
	rawServers, ok := radiustmp["server"].([]interface{})
	if !ok {
		return out
	}
	for _, rs := range rawServers {
		server, ok := rs.(map[string]interface{})
		if !ok {
			continue
		}
		host := getStringFromMap("host", server)
		port, _ := server["port"].(int)
		out[fmt.Sprintf("%s:%d", host, port)] = server
	}
	return out
}

// getBoolFromMap returns the bool for key from m, or defaultVal if missing or not a bool.
func getBoolFromMap(key string, m map[string]interface{}, defaultVal bool) bool {
	if m == nil {
		return defaultVal
	}
	if val, ok := m[key]; ok {
		if boolVal, ok := val.(bool); ok {
			return boolVal
		}
	}
	return defaultVal
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
