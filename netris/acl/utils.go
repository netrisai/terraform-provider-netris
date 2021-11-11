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

package acl

import (
	"github.com/netrisai/netriswebapi/v1/types/portgroup"
	api "github.com/netrisai/netriswebapi/v2"
)

func getPortGroupByName(name string, clientset *api.Clientset) (*portgroup.PortGroup, bool) {
	list, err := clientset.PortGroup().Get()
	if err != nil {
		return nil, false
	}

	for _, pg := range list {
		if name == pg.Name {
			return pg, true
		}
	}

	return nil, false
}

func getPortGroupByID(id int, clientset *api.Clientset) (*portgroup.PortGroup, bool) {
	list, err := clientset.PortGroup().Get()
	if err != nil {
		return nil, false
	}

	for _, pg := range list {
		if id == pg.ID {
			return pg, true
		}
	}

	return nil, false
}
