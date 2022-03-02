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
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/netrisai/netriswebapi/http"
	"github.com/netrisai/netriswebapi/v1/types/routemap"
	api "github.com/netrisai/netriswebapi/v2"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: "Creates and manages BGP Route-maps",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of route-map",
			},
			"sequence": {
				Optional:    true,
				Type:        schema.TypeList,
				Description: "The block of sequence. The sequence number will be assigned automatically",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Current black free description",
						},
						"policy": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Permit or deny the routes which match below all match clauses within the current sequence. Possible values: `permit` or `deny`",
						},
						"match": {
							Optional:    true,
							Type:        schema.TypeList,
							Description: "Block of Rules for route matching.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:         schema.TypeString,
										ValidateFunc: validateMatchType,
										Required:     true,
										Description:  "Type of the object to match: `as_path`, `community`, `extended_community`, `large_community`, `ipv4_prefix_list`, `ipv4_next_hop`, `route_source`, `ipv6_prefix_list`, `ipv6_next_hop`, `local_preference`, `med`, `origin`, `route_tag`",
									},
									"value": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Value of the object. Only for types: `ipv6_next_hop`, `local_preference`, `med`, `origin`, `route_tag`. Possible value for type `origin` is: `egp`, `incomplete`, `igp`",
									},
									"objectid": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "The ID of bgp object. Only for types: `as_path`, `community`, `extended_community`, `large_community`, `ipv4_prefix_list`, `ipv4_next_hop`, `route_source`, `ipv6_prefix_list`",
									},
								},
							},
						},
						"action": {
							Optional: true,
							Type:     schema.TypeList,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Define whether to manipulate a particular BGP attribute or go to another sequence. Possible values: `set`, `goto`, `next`",
									},
									"parameter": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "The attribute to be manipulated. Possible values: `as_path`, `community`, `large_community`, `ipv4_next_hop`, `ipv6_next_hop`, `local_preference`, `med`, `origin`, `route_tag`, `weight`",
									},
									"value": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "New attribute value",
									},
								},
							},
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

	seqNum := 5
	sequences := []routemap.Sequence{}
	seqList := d.Get("sequence").([]interface{})
	for _, seq := range seqList {
		sequence := routemap.Sequence{}
		seqItem := seq.(map[string]interface{})
		actions := []routemap.SequenceAction{}
		actionList := seqItem["action"].([]interface{})
		for _, act := range actionList {
			action := routemap.SequenceAction{}
			actionItem := act.(map[string]interface{})
			action.Type = actionItem["type"].(string)
			action.Parameter = actionItem["parameter"].(string)
			if action.Type == "goto" || action.Type == "next" {
				action.Parameter = "community"
			}
			action.Value = actionItem["value"].(string)
			actions = append(actions, action)
		}

		matches := []routemap.SequenceMatch{}
		matchList := seqItem["match"].([]interface{})
		for _, mat := range matchList {
			match := routemap.SequenceMatch{}
			matchItem := mat.(map[string]interface{})
			match.Type = matchItem["type"].(string)
			if getType(matchItem["type"].(string)) == "object" {
				match.EbgpObject = matchItem["objectid"].(int)
			} else if getType(matchItem["type"].(string)) == "string" {
				match.Value = matchItem["value"].(string)
			}
			matches = append(matches, match)
		}
		sequence.Actions = actions
		sequence.Matches = matches
		sequence.Number = seqNum
		seqNum += 5
		sequence.Description = seqItem["description"].(string)
		sequence.Policy = seqItem["policy"].(string)
		sequences = append(sequences, sequence)
	}

	routeMapAdd := &routemap.RouteMap{
		Name:      d.Get("name").(string),
		Sequences: sequences,
	}

	js, _ := json.Marshal(routeMapAdd)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.RouteMap().Add(routeMapAdd)
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

	d.SetId(strconv.Itoa(idStruct.ID))
	return nil
}

func resourceRead(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	id, _ := strconv.Atoi(d.Id())
	obj, ok := findByID(id, clientset)
	if !ok {
		return fmt.Errorf("Coudn't find routemap '%s'", d.Get("name").(string))
	}

	d.SetId(strconv.Itoa(obj.ID))
	err := d.Set("name", obj.Name)
	if err != nil {
		return err
	}
	sequenceList := []interface{}{}
	for _, seq := range obj.Sequences {
		sequence := make(map[string]interface{})
		matches := []interface{}{}
		actions := []interface{}{}
		for _, m := range seq.Matches {
			match := make(map[string]interface{})
			match["type"] = m.Type
			if m.EbgpObject.(string) != "" {
				match["objectid"], _ = strconv.Atoi(m.EbgpObject.(string))
			} else {
				match["value"] = m.Value
			}
			matches = append(matches, match)
		}
		for _, a := range seq.Actions {
			action := make(map[string]interface{})
			action["type"] = a.Type
			if a.Type != "next" {
				action["value"] = a.Value
				if a.Parameter != "" && a.Type != "goto" {
					action["parameter"] = a.Parameter
				}
			}
			actions = append(actions, action)
		}
		sequence["action"] = actions
		sequence["match"] = matches
		sequence["description"] = seq.Description
		sequence["policy"] = seq.Policy
		sequenceList = append(sequenceList, sequence)
	}
	err = d.Set("sequence", sequenceList)
	if err != nil {
		return err
	}

	return nil
}

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	seqNum := 5
	sequences := []routemap.Sequence{}
	seqList := d.Get("sequence").([]interface{})
	for _, seq := range seqList {
		sequence := routemap.Sequence{}
		seqItem := seq.(map[string]interface{})
		actions := []routemap.SequenceAction{}
		actionList := seqItem["action"].([]interface{})
		for _, act := range actionList {
			action := routemap.SequenceAction{}
			actionItem := act.(map[string]interface{})
			action.Type = actionItem["type"].(string)
			action.Parameter = actionItem["parameter"].(string)
			if action.Type == "goto" || action.Type == "next" {
				action.Parameter = "community"
			}
			action.Value = actionItem["value"].(string)
			actions = append(actions, action)
		}

		matches := []routemap.SequenceMatch{}
		matchList := seqItem["match"].([]interface{})
		for _, mat := range matchList {
			match := routemap.SequenceMatch{}
			matchItem := mat.(map[string]interface{})
			match.Type = matchItem["type"].(string)
			if getType(matchItem["type"].(string)) == "object" {
				match.EbgpObject = matchItem["objectid"].(int)
			} else if getType(matchItem["type"].(string)) == "string" {
				match.Value = matchItem["value"].(string)
			}
			matches = append(matches, match)
		}
		sequence.Actions = actions
		sequence.Matches = matches
		sequence.Number = seqNum
		seqNum += 5
		sequence.Description = seqItem["description"].(string)
		sequence.Policy = seqItem["policy"].(string)
		sequences = append(sequences, sequence)
	}

	id, _ := strconv.Atoi(d.Id())
	routeMapUpdate := &routemap.RouteMap{
		ID:        id,
		Name:      d.Get("name").(string),
		Sequences: sequences,
	}

	js, _ := json.Marshal(routeMapUpdate)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.RouteMap().Update(routeMapUpdate)
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

	id, _ := strconv.Atoi(d.Id())
	reply, err := clientset.RouteMap().Delete(id)
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
	id, _ := strconv.Atoi(d.Id())
	_, ok := findByID(id, clientset)
	return ok, nil
}

func resourceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	clientset := m.(*api.Clientset)
	name := d.Id()
	var obj *routemap.RouteMap
	var ok bool
	obj, ok = findByName(name, clientset)
	if !ok {
		return []*schema.ResourceData{d}, fmt.Errorf("Coudn't find routemap '%s'", d.Get("name").(string))
	}
	d.SetId(strconv.Itoa(obj.ID))

	return []*schema.ResourceData{d}, nil
}
