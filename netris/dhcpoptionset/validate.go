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

package dhcpoptionset

import (
	"fmt"
	"regexp"
)

func validateState(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if !(v == "enabled" || v == "disabled") {
		errs = append(errs, fmt.Errorf("'%s' must be enabled or disabled, got: %s", key, v))
	}
	return warns, errs
}

func validateIPPrefix(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	re := regexp.MustCompile(`(^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\/([0-9]|[12]\d|3[0-2]))?$)|(^((([0-9A-Fa-f]{1,4}:){7}([0-9A-Fa-f]{1,4}|:))|(([0-9A-Fa-f]{1,4}:){6}(:[0-9A-Fa-f]{1,4}|((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9A-Fa-f]{1,4}:){5}(((:[0-9A-Fa-f]{1,4}){1,2})|:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9A-Fa-f]{1,4}:){4}(((:[0-9A-Fa-f]{1,4}){1,3})|((:[0-9A-Fa-f]{1,4})?:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){3}(((:[0-9A-Fa-f]{1,4}){1,4})|((:[0-9A-Fa-f]{1,4}){0,2}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){2}(((:[0-9A-Fa-f]{1,4}){1,5})|((:[0-9A-Fa-f]{1,4}){0,3}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){1}(((:[0-9A-Fa-f]{1,4}){1,6})|((:[0-9A-Fa-f]{1,4}){0,4}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(:(((:[0-9A-Fa-f]{1,4}){1,7})|((:[0-9A-Fa-f]{1,4}){0,5}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:)))(%.+)?(\/([0-9]{1,2}|1[01][0-9]|12[0-8]))?$)`)
	if !re.Match([]byte(v)) {
		errs = append(errs, fmt.Errorf("invalid %s: %s", key, v))
	}
	return warns, errs
}

func validateIP(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	re := regexp.MustCompile(`(^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\/([0-9]|[12]\d|3[0-2]))?$)|(^((([0-9A-Fa-f]{1,4}:){7}([0-9A-Fa-f]{1,4}|:))|(([0-9A-Fa-f]{1,4}:){6}(:[0-9A-Fa-f]{1,4}|((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9A-Fa-f]{1,4}:){5}(((:[0-9A-Fa-f]{1,4}){1,2})|:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9A-Fa-f]{1,4}:){4}(((:[0-9A-Fa-f]{1,4}){1,3})|((:[0-9A-Fa-f]{1,4})?:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){3}(((:[0-9A-Fa-f]{1,4}){1,4})|((:[0-9A-Fa-f]{1,4}){0,2}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){2}(((:[0-9A-Fa-f]{1,4}){1,5})|((:[0-9A-Fa-f]{1,4}){0,3}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){1}(((:[0-9A-Fa-f]{1,4}){1,6})|((:[0-9A-Fa-f]{1,4}){0,4}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(:(((:[0-9A-Fa-f]{1,4}){1,7})|((:[0-9A-Fa-f]{1,4}){0,5}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:)))(%.+)?(\/([0-9]{1,2}|1[01][0-9]|12[0-8]))?$)`)
	if !re.Match([]byte(v)) {
		errs = append(errs, fmt.Errorf("invalid %s: %s", key, v))
	}
	return warns, errs
}

func validateMultihop(val interface{}, key string) (warns []string, errs []error) {
	if key == "neighboraddress" || key == "updatesource" {
		return validateIP(val, key)
	}
	return warns, errs
}

var dhcpStandardOptionTypes = map[int]string{
	2:   "int32",
	4:   "ipv4-address",
	6:   "ipv4-address",
	7:   "ipv4-address",
	8:   "ipv4-address",
	9:   "ipv4-address",
	10:  "ipv4-address",
	11:  "ipv4-address",
	13:  "uint16",
	14:  "string",
	15:  "fqdn",
	16:  "ipv4-address",
	17:  "string",
	18:  "string",
	19:  "boolean",
	20:  "boolean",
	21:  "ipv4-address",
	22:  "uint16",
	23:  "uint8",
	24:  "uint32",
	25:  "uint16",
	26:  "uint16",
	27:  "boolean",
	28:  "ipv4-address",
	29:  "boolean",
	30:  "boolean",
	31:  "boolean",
	32:  "ipv4-address",
	33:  "ipv4-address",
	34:  "boolean",
	35:  "uint32",
	36:  "boolean",
	37:  "uint8",
	38:  "uint32",
	39:  "boolean",
	40:  "string",
	44:  "ipv4-address",
	45:  "ipv4-address",
	46:  "uint8",
	47:  "string",
	48:  "ipv4-address",
	49:  "ipv4-address",
	52:  "uint8",
	54:  "ipv4-address",
	56:  "string",
	57:  "uint16",
	60:  "string",
	62:  "string",
	64:  "string",
	65:  "ipv4-address",
	66:  "string",
	67:  "string",
	68:  "ipv4-address",
	69:  "ipv4-address",
	70:  "ipv4-address",
	71:  "ipv4-address",
	72:  "ipv4-address",
	73:  "ipv4-address",
	74:  "ipv4-address",
	75:  "ipv4-address",
	76:  "ipv4-address",
	85:  "ipv4-address",
	86:  "string",
	87:  "string",
	88:  "fqdn",
	89:  "ipv4-address",
	93:  "uint16",
	98:  "string",
	100: "string",
	101: "string",
	112: "ipv4-address",
	113: "string",
	114: "string",
	116: "uint8",
	117: "uint16",
	118: "ipv4-address",
	136: "ipv4-address",
	137: "fqdn",
	138: "ipv4-address",
	141: "fqdn",
	160: "string",
	213: "fqdn",
}
