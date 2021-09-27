package sw

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"

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
			"nos": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the resource, also acts as it's unique ID",
			},
			"asnumber": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the resource, also acts as it's unique ID",
			},
			"profile": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of the resource, also acts as it's unique ID",
			},
			"mainip": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the resource, also acts as it's unique ID",
			},
			"mgmtip": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the resource, also acts as it's unique ID",
			},
			"macaddress": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of the resource, also acts as it's unique ID",
			},
			"portcount": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The name of the resource, also acts as it's unique ID",
			},
			"links": {
				Optional: true,
				Type:     schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"localport": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"remoteport": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
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

	links := d.Get("links").([]interface{})
	var linksInterfaces []map[string]interface{}
	for _, site := range links {
		linksInterfaces = append(linksInterfaces, site.(map[string]interface{}))
	}

	re := regexp.MustCompile(`^[a-zA-Z]+(?P<id>\d+)@\w+`)

	linksList := []inventory.HWLink{}
	for _, l := range linksInterfaces {
		link := inventory.HWLink{}
		if local, localok := l["localport"]; localok && len(local.(string)) > 0 {
			valueMatch := re.FindStringSubmatch(local.(string))
			result := regParser(valueMatch, re.SubexpNames())
			localID, _ := strconv.Atoi(result["id"])
			link.Local = inventory.IDName{ID: localID}
		}
		if remote, remoteok := l["remoteport"]; remoteok && len(remote.(string)) > 0 {
			link.Remote = inventory.IDName{Name: remote.(string)}
		}
		if link.Local.ID == 0 || link.Remote.Name == "" {
			return fmt.Errorf("invalid links")
		}
		linksList = append(linksList, link)
	}

	nosList, err := clientset.Inventory().GetNOS()
	if err != nil {
		return err
	}

	nosMap := make(map[string]inventory.NOS)
	for _, nos := range nosList {
		nosMap[nos.Tag] = *nos
	}

	nos := nosMap[d.Get("nos").(string)]

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

	profileID := 0
	profiles, err := clientset.Inventory().GetProfiles()
	if err != nil {
		return err
	}
	for _, profile := range profiles {
		if profile.Name == d.Get("profile").(string) {
			profileID = profile.ID
			break
		}
	}

	swAdd := &inventory.HWSwitchAdd{
		Name:        d.Get("name").(string),
		Tenant:      inventory.IDName{Name: d.Get("tenant").(string)},
		Site:        inventory.IDName{ID: siteID},
		Description: d.Get("description").(string),
		Nos:         nos,
		Asn:         d.Get("asnumber").(string),
		Profile:     inventory.IDName{ID: profileID},
		MainAddress: d.Get("mainip").(string),
		MgmtAddress: d.Get("mgmtip").(string),
		MacAddress:  d.Get("macaddress").(string),
		PortCount:   d.Get("portcount").(int),
		Links:       linksList,
	}

	js, _ := json.Marshal(swAdd)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.Inventory().AddSwitch(swAdd)
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
	d.SetId(swAdd.Name)
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
	err = d.Set("nos", sw.Nos.Tag)
	if err != nil {
		return err
	}
	err = d.Set("asnumber", strconv.Itoa(sw.Asn))
	if err != nil {
		return err
	}
	err = d.Set("profile", sw.Profile.Name)
	if err != nil {
		return err
	}
	if main := d.Get("mainip"); main.(string) != "auto" {
		err = d.Set("mainip", sw.MainAddress)
		if err != nil {
			return err
		}
	}
	if main := d.Get("mgmtip"); main.(string) != "auto" {
		err = d.Set("mgmtip", sw.MainAddress)
		if err != nil {
			return err
		}
	}
	err = d.Set("macaddress", sw.MacAddress)
	if err != nil {
		return err
	}
	err = d.Set("portcount", sw.PortCount)
	if err != nil {
		return err
	}

	var links []map[string]interface{}
	for _, link := range sw.Links {
		l := make(map[string]interface{})
		localPort, err := clientset.Port().GetByID(link.Local.ID)
		if err != nil {
			return err
		}
		remotePort, err := clientset.Port().GetByID(link.Remote.ID)
		if err != nil {
			return err
		}
		l["localport"] = localPort.ShortName
		l["remoteport"] = remotePort.ShortName
		links = append(links, l)
	}
	err = d.Set("links", links)
	if err != nil {
		return err
	}

	return nil
}

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	links := d.Get("links").([]interface{})
	var linksInterfaces []map[string]interface{}
	for _, site := range links {
		linksInterfaces = append(linksInterfaces, site.(map[string]interface{}))
	}

	linksList := []inventory.HWLink{}
	for _, l := range linksInterfaces {
		link := inventory.HWLink{}
		if local, localok := l["localport"]; localok && len(local.(string)) > 0 {
			link.Local = inventory.IDName{Name: local.(string)}
		}
		if remote, remoteok := l["remoteport"]; remoteok && len(remote.(string)) > 0 {
			link.Remote = inventory.IDName{Name: remote.(string)}
		}
		if link.Local.Name == "" || link.Remote.Name == "" {
			return fmt.Errorf("invalid links")
		}
		linksList = append(linksList, link)
	}

	nosList, err := clientset.Inventory().GetNOS()
	if err != nil {
		return err
	}

	nosMap := make(map[string]inventory.NOS)
	for _, nos := range nosList {
		nosMap[nos.Tag] = *nos
	}

	nos := nosMap[d.Get("nos").(string)]

	profileID := 0
	profiles, err := clientset.Inventory().GetProfiles()
	if err != nil {
		return err
	}
	for _, profile := range profiles {
		if profile.Name == d.Get("profile").(string) {
			profileID = profile.ID
			break
		}
	}

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

	swUpdate := &inventory.HWSwitchUpdate{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Tenant:      inventory.IDName{Name: d.Get("tenant").(string)},
		Site:        inventory.IDName{ID: siteID},
		Nos:         nos,
		Asn:         d.Get("asnumber").(string),
		Profile:     inventory.IDName{ID: profileID},
		MainAddress: d.Get("mainip").(string),
		MgmtAddress: d.Get("mgmtip").(string),
		MacAddress:  d.Get("macaddress").(string),
		PortCount:   d.Get("portcount").(int),
		Links:       linksList,
		Type:        "switch",
	}

	js, _ := json.Marshal(swUpdate)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.Inventory().UpdateSwitch(d.Get("itemid").(int), swUpdate)
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

	reply, err := clientset.Inventory().Delete("switch", d.Get("itemid").(int))
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

func regParser(valueMatch []string, subexpNames []string) map[string]string {
	result := make(map[string]string)
	for i, name := range subexpNames {
		if i != 0 && name != "" && len(valueMatch) >= len(subexpNames) {
			result[name] = valueMatch[i]
		}
	}
	return result
}
