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

package networkinterface

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

func validateBreakout(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	allowedValues := []string{"off", "disabled", "4x10", "4x25", "2x50", "4x50", "2x100", "4x100", "2x200", "4x200", "2x400"}

	// Check if the provided value is in the list of allowed values
	isValid := false
	for _, allowedValue := range allowedValues {
		if v == allowedValue {
			isValid = true
			break
		}
	}

	if !isValid {
		errs = append(errs, fmt.Errorf("Breakout available values are %v", allowedValues))
	}

	return warns, errs
}

func validateAutoneg(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if !(v == "default" || v == "on" || v == "off") {
		errs = append(errs, fmt.Errorf("Autoneg available values are (default, on, off)"))
		return warns, errs
	}
	return warns, errs
}

func validateExtension(val interface{}, key string) (warns []string, errs []error) {
	ext := val.(map[string]interface{})
	if v, ok := ext["vlanrange"]; ok {
		vlanrange := strings.Split(v.(string), "-")
		if len(vlanrange) != 2 {
			errs = append(errs, fmt.Errorf("Invalid vlan range in network interface extension"))
			return warns, errs
		}
	}
	if v, ok := ext["extensionname"]; ok && v == "" {
		errs = append(errs, fmt.Errorf("Empty extension name"))
		return warns, errs
	}
	return warns, errs
}

func validateSpeed(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if !(v == "auto" || v == "1g" || v == "10g" || v == "25g" || v == "40g" || v == "50g" || v == "100g" || v == "200g" || v == "400g") {
		errs = append(errs, fmt.Errorf("Speed available values are (auto, 1g, 10g, 25g, 40g, 50g, 100g, 200g, 400g)"))
		return warns, errs
	}
	return warns, errs
}

func validatePort(val interface{}, key string) (warns []string, errs []error) {
	if err := valPort(val.(string)); err != nil {
		errs = append(errs, fmt.Errorf(`Invalid value "%s". %s`, val.(string), err))
	}
	return warns, errs
}

func valPort(port string) error {
	log.Println("[DEBUG] network interface", port)
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
		return fmt.Errorf(`Port should be a range. Example: "20-22"`)
	}

	return nil
}
