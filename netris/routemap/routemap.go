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
		Schema: map[string]*schema.Schema{
			"itemid": {
				Type:             schema.TypeInt,
				Optional:         true,
				DiffSuppressFunc: DiffSuppress,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"sequence": {
				Optional: true,
				Type:     schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"policy": {
							Type:     schema.TypeString,
							Required: true,
						},
						"match": {
							Optional: true,
							Type:     schema.TypeList,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:         schema.TypeString,
										ValidateFunc: validateMatchType,
										Required:     true,
									},
									"value": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"objectid": {
										Type:     schema.TypeInt,
										Optional: true,
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
										Type:     schema.TypeString,
										Required: true,
									},
									"parameter": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"value": {
										Type:     schema.TypeString,
										Required: true,
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

	_ = d.Set("itemid", idStruct.ID)
	d.SetId(routeMapAdd.Name)
	return nil
}

func resourceRead(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	obj, ok := findByID(d.Get("itemid").(int), clientset)
	if !ok {
		return fmt.Errorf("Coudn't find routemap '%s'", d.Get("name").(string))
	}

	d.SetId(obj.Name)
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
			action["value"] = a.Value
			if a.Parameter != "" {
				action["parameter"] = a.Parameter
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

	routeMapUpdate := &routemap.RouteMap{
		ID:        d.Get("itemid").(int),
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

	reply, err := clientset.RouteMap().Delete(d.Get("itemid").(int))
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
	var ok bool
	_, ok = findByID(d.Get("itemid").(int), clientset)
	if !ok {
		return false, fmt.Errorf("Coudn't find routemap '%s'", d.Get("name").(string))
	}

	return true, nil
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
	err := d.Set("itemid", obj.ID)
	if err != nil {
		return []*schema.ResourceData{d}, err
	}

	return []*schema.ResourceData{d}, nil
}
