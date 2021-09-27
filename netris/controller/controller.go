package controller

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/netrisai/netriswebapi/http"
	api "github.com/netrisai/netriswebapi/v2"
	"github.com/netrisai/netriswebapi/v2/types/inventory"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"itemid": {
				Type:             schema.TypeInt,
				Optional:         true,
				DiffSuppressFunc: DiffSuppress,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the resource, also acts as it's unique ID",
			},
			"tenant": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the resource, also acts as it's unique ID",
			},
			"site": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the resource, also acts as it's unique ID",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of the resource, also acts as it's unique ID",
			},
			"mainip": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the resource, also acts as it's unique ID",
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

	siteID := 0
	sites, err := clientset.Site().Get()
	if err != nil {
		return err
	}
	for _, site := range sites {
		if site.Name == d.Get("site").(string) {
			siteID = site.ID
			break
		}
	}

	controllerAdd := &inventory.HWController{
		Name:        d.Get("name").(string),
		Tenant:      inventory.IDName{Name: d.Get("tenant").(string)},
		Site:        inventory.IDName{ID: siteID},
		Description: d.Get("description").(string),
		MainAddress: d.Get("mainip").(string),
	}

	js, _ := json.Marshal(controllerAdd)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.Inventory().AddController(controllerAdd)
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

	_ = d.Set("itemid", idStruct.ID)
	d.SetId(controllerAdd.Name)
	return nil
}

func resourceRead(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	sw, err := clientset.Inventory().GetByID(d.Get("itemid").(int))
	if err != nil {
		return err
	}

	d.SetId(sw.Name)
	err = d.Set("name", sw.Name)
	if err != nil {
		return err
	}
	err = d.Set("tenant", sw.Tenant.Name)
	if err != nil {
		return err
	}
	err = d.Set("site", sw.Site.Name)
	if err != nil {
		return err
	}
	err = d.Set("description", sw.Description)
	if err != nil {
		return err
	}
	if main := d.Get("mainip"); main.(string) != "auto" {
		err = d.Set("mainip", sw.MainAddress)
		if err != nil {
			return err
		}
	}

	return nil
}

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	controllerUpdate := &inventory.HWControllerUpdate{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		MainAddress: d.Get("mainip").(string),
	}

	js, _ := json.Marshal(controllerUpdate)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.Inventory().UpdateController(d.Get("itemid").(int), controllerUpdate)
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

	reply, err := clientset.Inventory().Delete("controller", d.Get("itemid").(int))
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

	sw, err := clientset.Inventory().GetByID(d.Get("itemid").(int))
	if err != nil {
		return false, err
	}

	if sw == nil {
		return false, nil
	}

	return true, nil
}

func resourceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	clientset := m.(*api.Clientset)

	sws, err := clientset.Inventory().Get()
	if err != nil {
		return []*schema.ResourceData{d}, err
	}
	name := d.Id()
	for _, sw := range sws {
		if sw.Name == name {
			err := d.Set("itemid", sw.ID)
			if err != nil {
				return []*schema.ResourceData{d}, err
			}
			return []*schema.ResourceData{d}, nil
		}
	}

	return []*schema.ResourceData{d}, nil
}
