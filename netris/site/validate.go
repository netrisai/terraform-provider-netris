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
