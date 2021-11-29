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

import (
	"github.com/netrisai/netriswebapi/v1/types/routemap"
	api "github.com/netrisai/netriswebapi/v2"
)

func findByID(id int, clientset *api.Clientset) (*routemap.RouteMap, bool) {
	list, err := clientset.RouteMap().Get()
	if err != nil {
		return nil, false
	}
	for _, obj := range list {
		if obj.ID == id {
			return obj, true
		}
	}
	return nil, false
}

func findByName(name string, clientset *api.Clientset) (*routemap.RouteMap, bool) {
	list, err := clientset.RouteMap().Get()
	if err != nil {
		return nil, false
	}
	for _, obj := range list {
		if obj.Name == name {
			return obj, true
		}
	}
	return nil, false
}
