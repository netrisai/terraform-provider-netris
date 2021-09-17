package subnet

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/netrisai/netriswebapi/http"
	"github.com/netrisai/netriswebapi/v2/types/ipam"

	api "github.com/netrisai/netriswebapi/v2"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"subnetid": {
				Type:             schema.TypeInt,
				Optional:         true,
				DiffSuppressFunc: DiffSuppress,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the resource, also acts as it's unique ID",
			},
			"prefix": {
				ForceNew: true,
				Required: true,
				Type:     schema.TypeString,
			},
			"tenant": {
				Required: true,
				Type:     schema.TypeString,
			},
			"purpose": {
				Required: true,
				Type:     schema.TypeString,
			},
			"defaultgateway": {
				Required: true,
				Type:     schema.TypeString,
			},
			"sites": {
				Optional:    true,
				Type:        schema.TypeList,
				Description: "Sites",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
		Create: resourceCreate,
		Read:   resourceRead,
		Update: resourceUpdate,
		Delete: resourceDelete,
		Exists: resourceExists,
		Importer: &schema.ResourceImporter{
			State: resourceImport,
		},
	}
}

func DiffSuppress(k, old, new string, d *schema.ResourceData) bool {
	return true
}

func resourceCreate(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	name := d.Get("name").(string)
	prefix := d.Get("prefix").(string)
	tenant := d.Get("tenant").(string)
	purpose := d.Get("purpose").(string)
	defaultgw := ""
	sitesList := d.Get("sites").([]interface{})
	sites := []ipam.IDName{}
	for _, s := range sitesList {
		sites = append(sites, ipam.IDName{Name: s.(string)})
	}

	if purpose == "management" {
		defaultgw = d.Get("defaultgateway").(string)
	}

	subnetAdd := &ipam.Subnet{
		Name:           name,
		Prefix:         prefix,
		Tenant:         ipam.IDName{Name: tenant},
		Purpose:        purpose,
		Sites:          sites,
		DefaultGateway: defaultgw,
	}

	js, _ := json.Marshal(subnetAdd)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.IPAM().AddSubnet(subnetAdd)
	if err != nil {
		log.Println("[DEBUG]", err)
		return err
	}

	js, _ = json.Marshal(reply)
	log.Println("[DEBUG]", string(js))

	log.Println("[DEBUG]", string(reply.Data))

	idStruct := struct {
		ID int `json:"id"`
	}{}

	data, err := reply.Parse()
	if err != nil {
		log.Println("[DEBUG]", err)
		return err
	}

	err = http.Decode(data.Data, &idStruct)
	if err != nil {
		log.Println("[DEBUG]", err)
		return err
	}

	log.Println("[DEBUG] ID:", idStruct.ID)

	if reply.StatusCode != 200 {
		return fmt.Errorf(string(reply.Data))
	}

	_ = d.Set("subnetid", idStruct.ID)
	d.SetId(subnetAdd.Name)

	return nil
}

func resourceRead(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	ipams, err := clientset.IPAM().GetSubnets()
	if err != nil {
		return err
	}
	id := d.Get("subnetid").(int)
	ipam := GetByID(ipams, id)
	if ipam == nil {
		return nil
	}

	d.SetId(ipam.Name)
	err = d.Set("name", ipam.Name)
	if err != nil {
		return err
	}
	err = d.Set("prefix", ipam.Prefix)
	if err != nil {
		return err
	}
	err = d.Set("tenant", ipam.Tenant.Name)
	if err != nil {
		return err
	}
	err = d.Set("purpose", ipam.Purpose)
	if err != nil {
		return err
	}
	err = d.Set("defaultgateway", ipam.DefaultGateway)
	if err != nil {
		return err
	}
	sites := []string{}
	for _, s := range ipam.Sites {
		sites = append(sites, s.Name)
	}
	err = d.Set("sites", sites)
	if err != nil {
		return err
	}
	return nil
}

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	name := d.Get("name").(string)
	prefix := d.Get("prefix").(string)
	tenant := d.Get("tenant").(string)
	purpose := d.Get("purpose").(string)
	defaultgw := ""
	sitesList := d.Get("sites").([]interface{})
	sites := []ipam.IDName{}
	for _, s := range sitesList {
		sites = append(sites, ipam.IDName{Name: s.(string)})
	}

	if purpose == "management" {
		defaultgw = d.Get("defaultgateway").(string)
	}

	subnetUpdate := &ipam.Subnet{
		Name:           name,
		Prefix:         prefix,
		Tenant:         ipam.IDName{Name: tenant},
		Purpose:        purpose,
		Sites:          sites,
		DefaultGateway: defaultgw,
	}

	js, _ := json.Marshal(subnetUpdate)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.IPAM().UpdateSubnet(d.Get("subnetid").(int), subnetUpdate)
	if err != nil {
		log.Println("[DEBUG]", err)
		return err
	}

	js, _ = json.Marshal(reply)
	log.Println("[DEBUG]", string(js))

	log.Println("[DEBUG]", string(reply.Data))

	if reply.StatusCode != 200 {
		return fmt.Errorf(string(reply.Data))
	}

	return nil
}

func resourceDelete(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	reply, err := clientset.IPAM().Delete("subnet", d.Get("subnetid").(int))
	if err != nil {
		return err
	}

	if reply.StatusCode != 200 {
		return fmt.Errorf(string(reply.Data))
	}

	d.SetId("")
	return nil
}

func resourceExists(d *schema.ResourceData, m interface{}) (bool, error) {
	clientset := m.(*api.Clientset)

	ipams, err := clientset.IPAM().GetSubnets()
	if err != nil {
		return false, err
	}
	id := d.Get("subnetid").(int)
	if ipam := GetByID(ipams, id); ipam == nil {
		return false, nil
	}

	return true, nil
}

func resourceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	clientset := m.(*api.Clientset)

	ipams, err := clientset.IPAM().GetSubnets()
	if err != nil {
		return []*schema.ResourceData{d}, err
	}
	prefix := d.Id()
	ipam := GetByPrefix(ipams, prefix)
	if ipam == nil {
		return []*schema.ResourceData{d}, fmt.Errorf("Allocation '%s' not found", prefix)
	}

	err = d.Set("subnetid", ipam.ID)
	if err != nil {
		return []*schema.ResourceData{d}, err
	}

	return []*schema.ResourceData{d}, nil
}

func GetByPrefix(list []*ipam.IPAM, prefix string) *ipam.IPAM {
	for _, s := range list {
		if s.Prefix == prefix {
			return s
		} else if len(s.Children) > 0 {
			if p := GetByPrefix(s.Children, prefix); p != nil {
				return p
			}
		}
	}
	return nil
}

func GetByID(list []*ipam.IPAM, id int) *ipam.IPAM {
	for _, s := range list {
		if s.ID == id {
			return s
		} else if len(s.Children) > 0 {
			if p := GetByID(s.Children, id); p != nil {
				return p
			}
		}
	}
	return nil
}
