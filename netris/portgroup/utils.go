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
	"github.com/netrisai/netriswebapi/v1/types/portgroup"

	api "github.com/netrisai/netriswebapi/v2"
)

func findPortGroupByID(id int, clientset *api.Clientset) (*portgroup.PortGroup, bool) {
	list, err := clientset.PortGroup().Get()
	if err != nil {
		return nil, false
	}
	for _, p := range list {
		if id == p.ID {
			return p, true
		}
	}

	return nil, false
}

func findPortGroupByName(name string, clientset *api.Clientset) (*portgroup.PortGroup, bool) {
	list, err := clientset.PortGroup().Get()
	if err != nil {
		return nil, false
	}
	for _, p := range list {
		if name == p.Name {
			return p, true
		}
	}

	return nil, false
}

func comparePorts(newPorts, oldPorts []string) (forAdd, forDelete []string) {
	newPortsMap := make(map[string]int)
	oldPortsMap := make(map[string]int)
	for _, n := range newPorts {
		newPortsMap[n] = 1
	}
	for _, o := range oldPorts {
		oldPortsMap[o] = 1
	}

	forAdd = []string{}
	forDelete = []string{}

	for _, n := range newPorts {
		if _, ok := oldPortsMap[n]; !ok {
			forAdd = append(forAdd, n)
		}
	}

	for _, n := range oldPorts {
		if _, ok := newPortsMap[n]; !ok {
			forDelete = append(forDelete, n)
		}
	}

	return forAdd, forDelete
}
