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

package portgroup

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

func validatePort(val interface{}, key string) (warns []string, errs []error) {
	if _, err := valPort(val.(string)); err != nil {
		errs = append(errs, fmt.Errorf(`Invalid value "%s". %s`, val.(string), err))
	}
	return warns, errs
}

func valPort(port string) (int, error) {
	log.Println("[DEBUG] port", port)
	v, err := strconv.Atoi(port)
	if err != nil {
		rg := strings.Split(port, "-")
		if len(rg) == 2 {
			_, err1 := valPort(rg[0])
			if err1 != nil {
				return 0, err1
			}
			_, err2 := valPort(rg[1])
			if err2 != nil {
				return 0, err2
			}
		} else {
			return 0, fmt.Errorf(`Port should be a number or range. Example: "80", "20-22"`)
		}
	} else if !(v >= 1 && v <= 65535) {
		return 0, fmt.Errorf("Port should be in range 1-65535")
	}

	return v, nil
}
