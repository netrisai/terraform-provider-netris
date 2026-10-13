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

package vnet

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func validateState(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if !(v == "active" || v == "disabled") {
		errs = append(errs, fmt.Errorf("'%s' must be active or disabled, got: %s", key, v))
	}
	return warns, errs
}

func validateIPFamily(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if !(v == "dual" || v == "ipv4" || v == "ipv6") {
		errs = append(errs, fmt.Errorf("'%s' must be dual, ipv4, or ipv6, got: %s", key, v))
	}
	return warns, errs
}

func validateIPv6NDRAMode(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if !(v == "default" || v == "enabled" || v == "disabled") {
		errs = append(errs, fmt.Errorf("'%s' must be default, enabled, or disabled, got: %s", key, v))
	}
	return warns, errs
}

func validateNonNegativeIntString(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if v == "" {
		return warns, errs
	}
	i, err := strconv.Atoi(v)
	if err != nil || i < 0 {
		errs = append(errs, fmt.Errorf("'%s' must be an empty string or a non-negative integer, got: %s", key, v))
	}
	return warns, errs
}

func validateVlanID(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if v == "auto" || v == "disabled" {
		// No further checks are needed for "auto" or "disabled"
	} else {
		vlan, _ := strconv.Atoi(v)
		if !(vlan >= 2 && vlan <= 4094) {
			errs = append(errs, fmt.Errorf("port should be in range 2-4094"))
		}
	}
	return warns, errs
}

func validateLACP(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if !(v == "on" || v == "off") {
		errs = append(errs, fmt.Errorf("'%s' must be on or off, got: %s", key, v))
	}
	return warns, errs
}

func validateGateway(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	re := regexp.MustCompile(`(^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\/([0-9]|[12]\d|3[0-2]))?$)|(^((([0-9A-Fa-f]{1,4}:){7}([0-9A-Fa-f]{1,4}|:))|(([0-9A-Fa-f]{1,4}:){6}(:[0-9A-Fa-f]{1,4}|((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9A-Fa-f]{1,4}:){5}(((:[0-9A-Fa-f]{1,4}){1,2})|:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9A-Fa-f]{1,4}:){4}(((:[0-9A-Fa-f]{1,4}){1,3})|((:[0-9A-Fa-f]{1,4})?:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){3}(((:[0-9A-Fa-f]{1,4}){1,4})|((:[0-9A-Fa-f]{1,4}){0,2}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){2}(((:[0-9A-Fa-f]{1,4}){1,5})|((:[0-9A-Fa-f]{1,4}){0,3}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){1}(((:[0-9A-Fa-f]{1,4}){1,6})|((:[0-9A-Fa-f]{1,4}){0,4}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(:(((:[0-9A-Fa-f]{1,4}){1,7})|((:[0-9A-Fa-f]{1,4}){0,5}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:)))(%.+)?(\/([0-9]{1,2}|1[01][0-9]|12[0-8]))?$)`)
	if !re.Match([]byte(v)) {
		errs = append(errs, fmt.Errorf("invalid gateway: %s", v))
	}
	return warns, errs
}

func validateDHCP(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if !(v == "enabled" || v == "disabled") {
		errs = append(errs, fmt.Errorf("'%s' must be enabled or disabled, got - %s", key, v))
	}
	return warns, errs
}

func validateUntagged(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if !(v == "yes" || v == "no") {
		errs = append(errs, fmt.Errorf("'%s' must be either 'yes' or 'no', got - %s", key, v))
	}
	return warns, errs
}

// validateDynamicVPCGroup checks the effective access mode using the same VLAN defaults
// as the port writer. It runs again at apply once unknown values have resolved.
func validateDynamicVPCGroup(get func(string) interface{}, known func(string) bool) error {
	if !known("dynamicvpcgroupid") || get("dynamicvpcgroupid").(int) == 0 {
		return nil
	}
	if known("vpcid") && get("vpcid").(int) > 0 {
		return fmt.Errorf("dynamicvpcgroupid conflicts with vpcid")
	}
	if known("l3vpn") && !get("l3vpn").(bool) {
		return fmt.Errorf("dynamicvpcgroupid requires l3vpn = true")
	}
	for _, key := range []string{"dhcprelay", "dhcpv6relay"} {
		if !known(key) {
			continue
		}
		for _, value := range get(key).([]interface{}) {
			if value != nil && value.(map[string]interface{})["enabled"].(bool) {
				return fmt.Errorf("dynamicvpcgroupid does not support enabled %s", key)
			}
		}
	}
	if !known("sites.#") {
		return nil
	}
	vlanID := get("vlanid").(string)
	for i, value := range get("sites").([]interface{}) {
		site := value.(map[string]interface{})
		if site["gateways"].(*schema.Set).Len() > 0 {
			return fmt.Errorf("dynamicvpcgroupid does not support gateways")
		}
		portKey := fmt.Sprintf("sites.%d.interface", i)
		portsKnown := known(portKey)
		ports := site["interface"].(*schema.Set).List()
		if portsKnown && len(ports) == 0 {
			portKey = fmt.Sprintf("sites.%d.ports", i)
			portsKnown = known(portKey)
			ports = site["ports"].(*schema.Set).List()
		}
		if portsKnown && known("vlanid") {
			for _, value := range ports {
				port := value.(map[string]interface{})
				untagged := port["untagged"].(string)
				accessMode := untagged == "yes" || (vlanID == "" && port["vlanid"].(string) == "1") || (vlanID != "" && untagged != "no")
				if !accessMode {
					return fmt.Errorf("dynamicvpcgroupid requires untagged member ports")
				}
			}
		}
		if known(fmt.Sprintf("sites.%d.interfacetag.#", i)) && site["interfacetag"].(*schema.Set).Len() > 0 {
			return fmt.Errorf("dynamicvpcgroupid does not support interface tags")
		}
	}
	return nil
}
