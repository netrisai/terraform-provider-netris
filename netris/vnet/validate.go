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
)

func validateState(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if !(v == "active" || v == "disabled") {
		errs = append(errs, fmt.Errorf("'%s' must be active or disabled, got: %s", key, v))
	}
	return warns, errs
}

func validateVlanID(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if v == "auto" {
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
		errs = append(errs, fmt.Errorf("'%s' must be enabled or disabled, got: %s", key, v))
	}
	return warns, errs
}
