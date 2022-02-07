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

package netris

import (
	api "github.com/netrisai/netriswebapi/v2"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/netrisai/terraform-provider-netris/netris/acl"
	"github.com/netrisai/terraform-provider-netris/netris/acl2"
	"github.com/netrisai/terraform-provider-netris/netris/allocation"
	"github.com/netrisai/terraform-provider-netris/netris/bgp"
	"github.com/netrisai/terraform-provider-netris/netris/bgpobject"
	"github.com/netrisai/terraform-provider-netris/netris/controller"
	"github.com/netrisai/terraform-provider-netris/netris/inventoryprofile"
	"github.com/netrisai/terraform-provider-netris/netris/l4lb"
	"github.com/netrisai/terraform-provider-netris/netris/link"
	"github.com/netrisai/terraform-provider-netris/netris/nat"
	"github.com/netrisai/terraform-provider-netris/netris/pgroup"
	"github.com/netrisai/terraform-provider-netris/netris/port"
	"github.com/netrisai/terraform-provider-netris/netris/portgroup"
	"github.com/netrisai/terraform-provider-netris/netris/roh"
	"github.com/netrisai/terraform-provider-netris/netris/route"
	"github.com/netrisai/terraform-provider-netris/netris/routemap"
	"github.com/netrisai/terraform-provider-netris/netris/site"
	"github.com/netrisai/terraform-provider-netris/netris/softgate"
	"github.com/netrisai/terraform-provider-netris/netris/subnet"
	"github.com/netrisai/terraform-provider-netris/netris/sw"
	"github.com/netrisai/terraform-provider-netris/netris/tenant"
	"github.com/netrisai/terraform-provider-netris/netris/user"
	"github.com/netrisai/terraform-provider-netris/netris/userrole"
	"github.com/netrisai/terraform-provider-netris/netris/vnet"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"address": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NETRIS_ADDRESS", ""),
			},
			"login": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NETRIS_LOGIN", ""),
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NETRIS_PASSWORD", ""),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"netris_vnet":              vnet.Resource(),
			"netris_bgp":               bgp.Resource(),
			"netris_l4lb":              l4lb.Resource(),
			"netris_allocation":        allocation.Resource(),
			"netris_subnet":            subnet.Resource(),
			"netris_tenant":            tenant.Resource(),
			"netris_switch":            sw.Resource(),
			"netris_controller":        controller.Resource(),
			"netris_softgate":          softgate.Resource(),
			"netris_user_role":         userrole.Resource(),
			"netris_user":              user.Resource(),
			"netris_permission_group":  pgroup.Resource(),
			"netris_acl":               acl.Resource(),
			"netris_roh":               roh.Resource(),
			"netris_portgroup":         portgroup.Resource(),
			"netris_inventory_profile": inventoryprofile.Resource(),
			"netris_bgp_object":        bgpobject.Resource(),
			"netris_site":              site.Resource(),
			"netris_routemap":          routemap.Resource(),
			"netris_link":              link.Resource(),
			"netris_nat":               nat.Resource(),
			"netris_port":              port.Resource(),
			"netris_route":             route.Resource(),
			"netris_acltwozero":        acl2.Resource(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"netris_site":              site.DataResource(),
			"netris_bgp_object":        bgpobject.DataResource(),
			"netris_tenant":            tenant.DataResource(),
			"netris_port":              port.DataResource(),
			"netris_vnet":              vnet.DataResource(),
			"netris_inventory_profile": inventoryprofile.DataResource(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	address := d.Get("address").(string)
	login := d.Get("login").(string)
	password := d.Get("password").(string)
	clientset, err := api.Client(address, login, password, 15)
	if err != nil {
		return nil, err
	}
	clientset.Client.InsecureVerify(true)

	err = clientset.Client.LoginUser()
	if err != nil {
		return nil, err
	}

	return clientset, nil
}
