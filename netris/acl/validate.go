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

package acl

import (
	"fmt"
	"regexp"
	"time"
)

func validateIPPrefix(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	re := regexp.MustCompile(`(^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\/([0-9]|[12]\d|3[0-2]))?$)|(^((([0-9A-Fa-f]{1,4}:){7}([0-9A-Fa-f]{1,4}|:))|(([0-9A-Fa-f]{1,4}:){6}(:[0-9A-Fa-f]{1,4}|((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9A-Fa-f]{1,4}:){5}(((:[0-9A-Fa-f]{1,4}){1,2})|:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9A-Fa-f]{1,4}:){4}(((:[0-9A-Fa-f]{1,4}){1,3})|((:[0-9A-Fa-f]{1,4})?:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){3}(((:[0-9A-Fa-f]{1,4}){1,4})|((:[0-9A-Fa-f]{1,4}){0,2}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){2}(((:[0-9A-Fa-f]{1,4}){1,5})|((:[0-9A-Fa-f]{1,4}){0,3}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){1}(((:[0-9A-Fa-f]{1,4}){1,6})|((:[0-9A-Fa-f]{1,4}){0,4}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(:(((:[0-9A-Fa-f]{1,4}){1,7})|((:[0-9A-Fa-f]{1,4}){0,5}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:)))(%.+)?(\/([1-9]|[1-5][0-9]|6[0-4]))?$)`)
	if !re.Match([]byte(v)) {
		errs = append(errs, fmt.Errorf("invalid %s: %s", key, v))
	}
	return warns, errs
}

func validateProto(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if !(v == "all" || v == "ip" || v == "tcp" || v == "udp" || v == "icmp" || v == "icmpv6") {
		errs = append(errs, fmt.Errorf("proto available values are (all, ip, tcp, udp, icmp, icmpv6)"))
		return warns, errs
	}
	return warns, errs
}

func validatePort(val interface{}, key string) (warns []string, errs []error) {
	v := val.(int)
	if !(v >= 1 && v <= 65535) {
		errs = append(errs, fmt.Errorf("Port should be in range 1-65535"))
		return warns, errs
	}
	return warns, errs
}

func validateEstablished(val interface{}, key string) (warns []string, errs []error) {
	v := val.(int)
	if !(v == 0 || v == 1) {
		errs = append(errs, fmt.Errorf("Established should be 0 or 1"))
		return warns, errs
	}
	return warns, errs
}

func validateICMP(val interface{}, key string) (warns []string, errs []error) {
	v := val.(int)
	if !(v >= 1 && v <= 37) {
		errs = append(errs, fmt.Errorf("ICMP type should be in range 1-37 according to RFC 1700"))
		return warns, errs
	}
	return warns, errs
}

func validateDate(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if _, err := time.Parse(time.RFC3339, v); err != nil {
		errs = append(errs, fmt.Errorf("invalid validuntil field. Date should be according to RFC 3339. Example 2006-01-02T15:04:05Z"))
	}
	return warns, errs
}
