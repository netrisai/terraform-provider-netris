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

package routemap

import "fmt"

func validateMatchType(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if !typeExist(v) {
		errs = append(errs, fmt.Errorf("invalid match type. Available values are (%s)", getTypesListString()))
	}
	return warns, errs
}
