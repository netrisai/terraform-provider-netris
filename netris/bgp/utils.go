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

package bgp

import (
	"fmt"

	api "github.com/netrisai/netriswebapi/v2"
	"github.com/netrisai/netriswebapi/v2/types/bgp"
)

func findPort(clientset *api.Clientset, siteID int, portName string) (*bgp.EBGPPort, bool) {
	ports, err := clientset.BGP().GetPorts(siteID)
	if err != nil {
		return nil, false
	}
	for _, port := range ports {
		if fmt.Sprintf("%s@%s", port.Port, port.SwitchName) == portName {
			return port, true
		}
	}
	return nil, false
}

func findVNetByName(clientset *api.Clientset, name string) (*bgp.EBGPVNet, bool) {
	vnets, err := clientset.BGP().GetVNets()
	if err != nil {
		return nil, false
	}
	for _, vnet := range vnets {
		if vnet.Name == name {
			return vnet, true
		}
	}
	return nil, false
}

func findSwitchByName(clientset *api.Clientset, siteID int, name string) (*bgp.EBGPSwitch, bool) {
	switches, err := clientset.BGP().GetSwitches(siteID)
	if err != nil {
		return nil, false
	}
	for _, item := range switches {
		if item.Location == name {
			return item, true
		}
	}

	return nil, false
}
