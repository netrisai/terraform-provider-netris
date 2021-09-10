package netris

import (
	api "github.com/netrisai/netriswebapi/v2"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
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
			"netris_vnet": resourceVNet(),
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
