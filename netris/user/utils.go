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

package user

import (
	"fmt"

	"github.com/netrisai/netriswebapi/v1/types/permission"
	"github.com/netrisai/netriswebapi/v1/types/tenant"
	"github.com/netrisai/netriswebapi/v1/types/userrole"

	api "github.com/netrisai/netriswebapi/v2"
)

func findRoleByName(name string, clientset *api.Clientset) (*userrole.UserRole, bool) {
	uroles, err := clientset.UserRole().Get()
	if err != nil {
		return nil, false
	}
	for _, role := range uroles {
		if role.Name == name {
			return role, true
		}
	}

	return nil, false
}

func findPgroupByName(name string, clientset *api.Clientset) (*permission.PermissionGroup, bool) {
	pgrps, err := clientset.Permission().Get()
	if err != nil {
		return nil, false
	}
	for _, pgrp := range pgrps {
		if pgrp.Name == name {
			return pgrp, true
		}
	}

	return nil, false
}

func findTenatsByNames(names []string, clientset *api.Clientset) ([]*tenant.Tenant, error) {
	tenants, err := clientset.Tenant().Get()
	if err != nil {
		return nil, err
	}

	ts := map[string]int{}
	for _, name := range names {
		ts[name] = 1
	}

	list := []*tenant.Tenant{}

	for _, tenant := range tenants {
		if _, ok := ts[tenant.Name]; ok {
			list = append(list, tenant)
		}
	}

	if len(names) != len(list) {
		return nil, fmt.Errorf("invalid tenants")
	}

	return list, nil
}
