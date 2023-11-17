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

package lag

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

func validateName(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	splited := strings.Split(v, "@")
	if len(splited) != 2 {
		errs = append(errs, fmt.Errorf("invalid name format. Example: agg1@switch1"))
		return warns, errs
	}
	return warns, errs
}

func validateAutoneg(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if !(v == "default" || v == "on" || v == "off") {
		errs = append(errs, fmt.Errorf("autoneg available values are (default, on, off)"))
		return warns, errs
	}
	return warns, errs
}

func validateExtension(val interface{}, key string) (warns []string, errs []error) {
	ext := val.(map[string]interface{})
	if v, ok := ext["vlanrange"]; ok {
		vlanrange := strings.Split(v.(string), "-")
		if len(vlanrange) != 2 {
			errs = append(errs, fmt.Errorf("invalid vlan range in port extension"))
			return warns, errs
		}
	}
	if v, ok := ext["extensionname"]; ok && v == "" {
		errs = append(errs, fmt.Errorf("empty extension name"))
		return warns, errs
	}
	return warns, errs
}

func validatePort(val interface{}, key string) (warns []string, errs []error) {
	if err := valPort(val.(string)); err != nil {
		errs = append(errs, fmt.Errorf(`invalid value "%s". %s`, val.(string), err))
	}
	return warns, errs
}

func valPort(port string) error {
	log.Println("[DEBUG] port", port)
	rg := strings.Split(port, "-")
	if len(rg) == 2 {
		_, err1 := strconv.Atoi(rg[0])
		if err1 != nil {
			return err1
		}
		_, err2 := strconv.Atoi(rg[1])
		if err2 != nil {
			return err2
		}
	} else {
		return fmt.Errorf(`port should be a range. Example: "20-22"`)
	}

	return nil
}
