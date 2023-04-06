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

package site

import (
	"fmt"
	"strconv"
	"strings"
)

func validateRoutingProfile(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if !(v == "default" || v == "default_agg" || v == "full_table") {
		errs = append(errs, fmt.Errorf("Routing profile available values are (default, default_agg, full_table)"))
		return warns, errs
	}
	return warns, errs
}

func validateSiteMesh(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if !(v == "disabled" || v == "hub" || v == "spoke" || v == "dspoke") {
		errs = append(errs, fmt.Errorf("Site mesh available values are (disabled, hub, spoke, dspoke)"))
		return warns, errs
	}
	return warns, errs
}

func validateACLPolicy(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if !(v == "permit" || v == "deny") {
		errs = append(errs, fmt.Errorf("ACL policy available values are (permit, deny)"))
		return warns, errs
	}
	return warns, errs
}

func validateSwitchFabric(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if !(v == "equinix_metal" || v == "dot1q_trunk" || v == "netris" || v == "phoenixnap_bmc") {
		errs = append(errs, fmt.Errorf("Switch fabric available values are (equinix_metal, phoenixnap_bmc, dot1q_trunk, netris)"))
		return warns, errs
	}
	return warns, errs
}

func validateVlanRange(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if err := valVlanRange(v); err != nil {
		errs = append(errs, err)
		return warns, errs
	}
	return warns, errs
}

func valVlanRange(vlan string) error {
	rg := strings.Split(vlan, "-")
	if len(rg) == 2 {
		p1, _ := strconv.Atoi(rg[0])
		err1 := valPort(p1)
		if err1 != nil {
			return err1
		}
		p2, _ := strconv.Atoi(rg[1])
		err2 := valPort(p2)
		if err2 != nil {
			return err2
		}

		if p1 >= p2 {
			return fmt.Errorf("Invalid vlan range")
		}
	}
	return nil
}

func valPort(vlan int) error {
	if !(vlan >= 2 && vlan <= 4094) {
		return fmt.Errorf("Port should be in range 2-4094")
	}

	return nil
}

func valEquinixVlanRange(s string) error {
	vlanSplited := strings.Split(s, "-")
	lastVlan := vlanSplited[len(vlanSplited)-1]
	if id, _ := strconv.Atoi(lastVlan); id > 3999 {
		return fmt.Errorf("Invalid vlan range %s", s)
	}
	return nil
}

func valPhoenixVlanRange(s string) error {
	vlanSplited := strings.Split(s, "-")
	lastVlan := vlanSplited[len(vlanSplited)-1]
	if id, _ := strconv.Atoi(lastVlan); id > 4094 {
		return fmt.Errorf("invalid vlan range %s", s)
	}
	return nil
}

func validateEquinixLocation(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if _, ok := equinixLocationsMap[v]; !ok {
		keys := make([]string, 0, len(equinixLocationsMap))
		for k := range equinixLocationsMap {
			keys = append(keys, k)
		}
		errs = append(errs, fmt.Errorf("Invalid equinixlocation, Possible Values are (%s)", strings.Join(keys, ", ")))
		return warns, errs
	}
	if err := valVlanRange(v); err != nil {
		errs = append(errs, err)
		return warns, errs
	}
	return warns, errs
}

func validatephoenixLocation(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if _, ok := phoenixLocationsMap[v]; !ok {
		keys := make([]string, 0, len(phoenixLocationsMap))
		for k := range phoenixLocationsMap {
			keys = append(keys, k)
		}
		errs = append(errs, fmt.Errorf("Invalid phoenixlocation, Possible Values are (%s)", strings.Join(keys, ", ")))
		return warns, errs
	}
	if err := valVlanRange(v); err != nil {
		errs = append(errs, err)
		return warns, errs
	}
	return warns, errs
}

var equinixLocationsMap = map[string]struct{}{
	"se": {},
	"dc": {},
	"at": {},
	"hk": {},
	"am": {},
	"ny": {},
	"ty": {},
	"sl": {},
	"md": {},
	"sp": {},
	"fr": {},
	"sy": {},
	"ld": {},
	"sg": {},
	"pa": {},
	"tr": {},
	"sv": {},
	"la": {},
	"ch": {},
	"da": {},
}

var phoenixLocationsMap = map[string]struct{}{
	"phx": {},
	"chi": {},
	"aus": {},
	"sgp": {},
	"ash": {},
	"sea": {},
	"nld": {},
}
