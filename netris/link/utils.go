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

package link

import (
	"fmt"
	"strconv"

	api "github.com/netrisai/netriswebapi/v2"
	"github.com/netrisai/netriswebapi/v2/types/inventory"
	"github.com/netrisai/netriswebapi/v2/types/port"
)

func findPortByID(ports []*port.Port, id int, clientset *api.Clientset) (*port.Port, bool) {
	for _, port := range ports {
		if port.ID == id {
			return port, true
		}
	}
	return nil, false
}

func findPortByName(ports []*port.Port, name string, clientset *api.Clientset) (*port.Port, bool) {
	for _, port := range ports {
		if fmt.Sprintf("%s@%s", port.Port, port.SwitchName) == name {
			return port, true
		}
	}
	return nil, false
}

func hwToSoftgateUpdate(hw *inventory.HW) *inventory.HWSoftgateUpdate {
	sg := &inventory.HWSoftgateUpdate{
		Description: hw.Description,
		Links:       hw.Links,
		MainAddress: hw.MainAddress,
		MgmtAddress: hw.MgmtAddress,
		Name:        hw.Name,
		Profile:     hw.Profile,
		Site:        hw.Site,
		Tenant:      hw.Tenant,
	}
	return sg
}

func hwToSwitchUpdate(hw *inventory.HW) *inventory.HWSwitchUpdate {
	sg := &inventory.HWSwitchUpdate{
		Asn:         strconv.Itoa(hw.Asn),
		Description: hw.Description,
		Links:       hw.Links,
		MainAddress: hw.MainAddress,
		MgmtAddress: hw.MgmtAddress,
		Name:        hw.Name,
		Nos:         inventory.NOS(hw.Nos),
		PortCount:   hw.PortCount,
		Profile:     hw.Profile,
		Site:        hw.Site,
		Tenant:      hw.Tenant,
		Type:        hw.Type,
	}
	return sg
}
