package allocation

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
			"ipamid": {
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
				Required: true,
				Type:     schema.TypeString,
			},
			"tenant": {
				Required: true,
				Type:     schema.TypeString,
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

	allAdd := &ipam.Allocation{
		Name:   name,
		Prefix: prefix,
		Tenant: ipam.IDName{Name: tenant},
	}

	js, _ := json.Marshal(allAdd)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.IPAM().AddAllocation(allAdd)
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

	_ = d.Set("ipamid", idStruct.ID)
	d.SetId(allAdd.Name)

	return nil
}

func resourceRead(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	ipams, err := clientset.IPAM().Get()
	if err != nil {
		return err
	}
	prefix := d.Get("prefix").(string)
	ipam := getByPrefix(ipams, prefix)
	if ipam == nil {
		return fmt.Errorf("prefix '%s' not found", prefix)
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
	return nil
}

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	name := d.Get("name").(string)
	prefix := d.Get("prefix").(string)
	tenant := d.Get("tenant").(string)

	allUpdate := &ipam.Allocation{
		Name:   name,
		Prefix: prefix,
		Tenant: ipam.IDName{Name: tenant},
	}

	js, _ := json.Marshal(allUpdate)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.IPAM().UpdateAllocation(d.Get("ipamid").(int), allUpdate)
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

	reply, err := clientset.IPAM().Delete("allocation", d.Get("ipamid").(int))
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

	ipams, err := clientset.IPAM().Get()
	if err != nil {
		return false, err
	}
	prefix := d.Get("prefix").(string)
	if ipam := getByPrefix(ipams, prefix); ipam == nil {
		return false, fmt.Errorf("prefix '%s' not found", prefix)
	}

	return true, nil
}

func resourceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	clientset := m.(*api.Clientset)

	ipams, err := clientset.IPAM().Get()
	if err != nil {
		return []*schema.ResourceData{d}, err
	}
	prefix := d.Get("prefix").(string)
	ipam := getByPrefix(ipams, prefix)
	if ipam == nil {
		return []*schema.ResourceData{d}, fmt.Errorf("prefix '%s' not found", prefix)
	}

	err = d.Set("ipamid", ipam.ID)
	if err != nil {
		return []*schema.ResourceData{d}, err
	}

	return []*schema.ResourceData{d}, nil
}

func getByPrefix(list []*ipam.IPAM, prefix string) *ipam.IPAM {
	for _, s := range list {
		if s.Prefix == prefix {
			return s
		} else if len(s.Children) > 0 {
			if p := getByPrefix(s.Children, prefix); p != nil {
				return p
			}
		}
	}
	return nil
}
