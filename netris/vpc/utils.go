/*
Copyright 2023. Netris, Inc.

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

package vpc

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/netrisai/netriswebapi/http"
	"github.com/netrisai/netriswebapi/v2/types/vpc"
)

func parseWriteReply(reply http.HTTPReply) (*vpc.VPC, error) {
	var result struct {
		IsSuccess bool    `json:"isSuccess"`
		Data      vpc.VPC `json:"data"`
	}
	if err := json.Unmarshal(reply.Data, &result); err != nil {
		return nil, fmt.Errorf("VPC request returned an invalid response (HTTP %d): %s: %s", reply.StatusCode, err, reply.Data)
	}
	if reply.StatusCode != 200 || !result.IsSuccess {
		return nil, fmt.Errorf("VPC request failed (HTTP %d): %s", reply.StatusCode, reply.Data)
	}
	return &result.Data, nil
}

func routeTargetSchema(computed bool) *schema.Schema {
	value := &schema.Schema{Type: schema.TypeString, Required: !computed, Computed: computed}
	if !computed {
		value.ValidateFunc = validateRouteTarget
	}
	result := &schema.Schema{
		Type:        schema.TypeList,
		Optional:    !computed,
		Computed:    computed,
		Description: "Managed route targets. Omit to preserve; an empty block removes all targets in this direction. Requires Controller 4.21 or later.",
		Elem: &schema.Resource{Schema: map[string]*schema.Schema{
			"target": {
				Type:     schema.TypeList,
				Optional: !computed,
				Computed: computed,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"id":    {Type: schema.TypeInt, Computed: true},
					"value": value,
					"type":  {Type: schema.TypeString, Computed: true},
				}},
			},
		}},
	}
	if computed {
		result.Description = "Controller route targets for this direction."
	} else {
		result.MaxItems = 1
	}
	return result
}

func validateVNI(value interface{}, key string) ([]string, []error) {
	s := value.(string)
	if s == "auto" {
		return nil, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 || n > 16777215 || strconv.Itoa(n) != s {
		return nil, []error{fmt.Errorf("%s must be auto or an integer from 1 to 16777215", key)}
	}
	return nil, nil
}

func expandVNI(d *schema.ResourceData) interface{} {
	s := d.Get("vni").(string)
	if s == "" {
		return nil
	}
	if s == "auto" {
		return s
	}
	n, _ := strconv.Atoi(s)
	return n
}

func routeTargetRows(raw interface{}) []interface{} {
	blocks := raw.([]interface{})
	if len(blocks) == 0 || blocks[0] == nil {
		return nil
	}
	return blocks[0].(map[string]interface{})["target"].([]interface{})
}

func expandRouteTargets(d *schema.ResourceData) *vpc.RouteTargetsWrite {
	routes := &vpc.RouteTargetsWrite{}
	for _, key := range []string{"importrt", "exportrt"} {
		if len(d.Get(key).([]interface{})) == 0 {
			continue
		}
		old, _ := d.GetChange(key)
		previous := routeTargetRows(old)
		rows := routeTargetRows(d.Get(key))
		targets := make([]vpc.RouteTargetWrite, len(rows))
		used := make(map[int]bool)
		// Match retained values before assigning IDs to edits, so reorder/deletion
		// cannot accidentally turn a default or peering target into a custom one.
		for i, raw := range rows {
			row := raw.(map[string]interface{})
			targets[i].Value = row["value"].(string)
			for _, oldRaw := range previous {
				prior := oldRaw.(map[string]interface{})
				id := prior["id"].(int)
				if prior["value"] == targets[i].Value && !used[id] {
					targets[i].ID = id
					used[id] = true
					break
				}
			}
		}
		for i := range targets {
			if targets[i].ID != 0 || i >= len(previous) {
				continue
			}
			id := previous[i].(map[string]interface{})["id"].(int)
			if !used[id] {
				targets[i].ID = id
				used[id] = true
			}
		}
		if key == "importrt" {
			routes.Import = &targets
		} else {
			routes.Export = &targets
		}
	}
	if routes.Import == nil && routes.Export == nil {
		return nil
	}
	return routes
}

func remapRouteTargetIDs(writes *vpc.RouteTargetsWrite, before, current *vpc.VPC) {
	if before == nil || current == nil || before.RouteTargets == nil || current.RouteTargets == nil {
		return
	}
	for _, direction := range []struct {
		writes          *[]vpc.RouteTargetWrite
		before, current []vpc.RouteTarget
	}{
		{writes.Import, before.RouteTargets.Import, current.RouteTargets.Import},
		{writes.Export, before.RouteTargets.Export, current.RouteTargets.Export},
	} {
		if direction.writes == nil {
			continue
		}
		for i, row := range *direction.writes {
			for _, old := range direction.before {
				if row.ID == 0 || row.ID != old.ID || old.Type != "default-RT" {
					continue
				}
				(*direction.writes)[i].ID = 0
				for _, replacement := range direction.current {
					if replacement.Type == "default-RT" {
						(*direction.writes)[i].ID = replacement.ID
						break
					}
				}
			}
		}
	}
}

func flattenRouteTargets(rows []vpc.RouteTarget, previous []interface{}) []map[string]interface{} {
	ordered := make([]vpc.RouteTarget, 0, len(rows))
	used := make(map[int]bool)
	for _, raw := range previous {
		prior := raw.(map[string]interface{})
		var match *vpc.RouteTarget
		for i := range rows {
			if !used[rows[i].ID] && rows[i].Value == prior["value"] {
				match = &rows[i]
				break
			}
		}
		if match == nil {
			for i := range rows {
				if !used[rows[i].ID] && rows[i].ID == prior["id"] {
					match = &rows[i]
					break
				}
			}
		}
		// A VNI change can replace both the default target's ID and value.
		if match == nil && prior["type"] == "default-RT" {
			for i := range rows {
				if !used[rows[i].ID] && rows[i].Type == "default-RT" {
					match = &rows[i]
					break
				}
			}
		}
		if match != nil {
			ordered = append(ordered, *match)
			used[match.ID] = true
		}
	}
	for _, row := range rows {
		if !used[row.ID] {
			ordered = append(ordered, row)
		}
	}
	result := make([]map[string]interface{}, 0, len(ordered))
	for _, row := range ordered {
		result = append(result, map[string]interface{}{"id": row.ID, "value": row.Value, "type": row.Type})
	}
	return []map[string]interface{}{{"target": result}}
}

func flattenVNIAndRouteTargets(d *schema.ResourceData, item *vpc.VPC, dataSource bool) error {
	value := ""
	if item.VNI != nil {
		value = strconv.Itoa(*item.VNI)
		if !dataSource && item.VNIMode == "auto" {
			value = "auto"
		}
	}
	if item.VNI != nil || dataSource {
		if err := d.Set("vni", value); err != nil {
			return err
		}
	}
	if dataSource {
		if err := d.Set("vnimode", item.VNIMode); err != nil {
			return err
		}
	}
	if item.RouteTargets == nil {
		if dataSource {
			if err := d.Set("importrt", nil); err != nil {
				return err
			}
			return d.Set("exportrt", nil)
		}
		return nil
	}
	for _, direction := range []struct {
		key  string
		rows []vpc.RouteTarget
	}{
		{"importrt", item.RouteTargets.Import}, {"exportrt", item.RouteTargets.Export},
	} {
		if !dataSource && len(d.Get(direction.key).([]interface{})) == 0 {
			continue
		}
		if err := d.Set(direction.key, flattenRouteTargets(direction.rows, routeTargetRows(d.Get(direction.key)))); err != nil {
			return err
		}
	}
	return nil
}

func validateRouteTarget(value interface{}, key string) ([]string, []error) {
	parts := strings.Split(value.(string), ":")
	if len(parts) == 2 {
		number, err := strconv.ParseUint(parts[1], 10, 32)
		if err == nil && strconv.FormatUint(number, 10) == parts[1] {
			if ip := net.ParseIP(parts[0]); ip != nil && ip.To4() != nil && number <= 65535 {
				return nil, nil
			}
			asn, err := strconv.ParseUint(parts[0], 10, 32)
			if err == nil && strconv.FormatUint(asn, 10) == parts[0] && (asn <= 65535 || number <= 65535) {
				return nil, nil
			}
		}
	}
	return nil, []error{fmt.Errorf("%s must be an ASN:number or IPv4:number route target within the supported ranges, without leading zeros", key)}
}
