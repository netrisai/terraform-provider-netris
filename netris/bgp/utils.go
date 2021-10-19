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
	"log"

	api "github.com/netrisai/netriswebapi/v2"
	"github.com/netrisai/netriswebapi/v2/types/bgp"
	"github.com/netrisai/netriswebapi/v2/types/vnet"
)

func findPortByID(clientset *api.Clientset, siteID int, id int) (*bgp.EBGPPort, bool) {
	ports, err := clientset.BGP().GetPorts(siteID)
	if err != nil {
		return nil, false
	}
	for _, port := range ports {
		if port.PortID == id {
			return port, true
		}
	}
	return nil, false
}

func findVNetByName(clientset *api.Clientset, name string) (*vnet.VNet, bool) {
	vnets, err := clientset.VNet().Get()
	log.Println("[DEBUG] vnets", vnets)
	if err != nil {
		return nil, false
	}
	for _, vnet := range vnets {
		log.Println("[DEBUG] ", vnet.Name)
		if vnet.Name == name {
			return vnet, true
		}
	}
	return nil, false
}
