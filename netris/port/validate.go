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

package port

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

func validateBreakout(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if !(v == "off" || v == "4x10g" || v == "4x25g" || v == "4x100g" || v == "manual") {
		errs = append(errs, fmt.Errorf("Breakout available values are (off, 4x10g, 4x25g, 4x100g, manual)"))
		return warns, errs
	}
	return warns, errs
}

func validateAutoneg(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if !(v == "none" || v == "on" || v == "off") {
		errs = append(errs, fmt.Errorf("Autoneg available values are (none, on, off)"))
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
		return fmt.Errorf(`Port should be a range. Example: "20-22"`)
	}

	return nil
}
