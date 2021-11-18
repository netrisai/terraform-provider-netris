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

package inventoryprofile

import (
	"encoding/json"

	"github.com/netrisai/netriswebapi/v1/types/inventoryprofile"
	api "github.com/netrisai/netriswebapi/v2"
)

func findByID(id int, clientset *api.Clientset) (*inventoryprofile.Profile, bool) {
	list, err := clientset.InventoryProfile().Get()
	if err != nil {
		return nil, false
	}
	for _, profile := range list {
		if profile.ID == id {
			return profile, true
		}
	}
	return nil, false
}

func findByName(name string, clientset *api.Clientset) (*inventoryprofile.Profile, bool) {
	list, err := clientset.InventoryProfile().Get()
	if err != nil {
		return nil, false
	}
	for _, profile := range list {
		if profile.Name == name {
			return profile, true
		}
	}
	return nil, false
}

func unmarshalTimezone(s string) *inventoryprofile.Timezone {
	timezone := &inventoryprofile.Timezone{}
	_ = json.Unmarshal([]byte(s), timezone)
	return timezone
}
