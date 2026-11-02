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
		Description: "Managed custom route targets. Omit to preserve; an empty block removes custom targets in this direction. Controller-generated targets are preserved. Requires Controller 4.21 or later.",
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

func validateRouteTargetDiff(d *schema.ResourceDiff, m interface{}) error {
	for _, key := range []string{"importrt", "exportrt"} {
		if !d.NewValueKnown(key) || !d.NewValueKnown(key+".0.target") {
			continue
		}
		seen := make(map[string]bool)
		for i, raw := range routeTargetRows(d.Get(key)) {
			path := fmt.Sprintf("%s.0.target.%d.value", key, i)
			if !d.NewValueKnown(path) {
				continue
			}
			value := raw.(map[string]interface{})["value"].(string)
			if seen[value] {
				return fmt.Errorf("%s contains duplicate route target %q", key, value)
			}
			seen[value] = true
		}
	}
	return nil
}

func expandRouteTargets(d *schema.ResourceData, current *vpc.VPC) (*vpc.RouteTargetsWrite, error) {
	routes := &vpc.RouteTargetsWrite{}
	vni := expandVNI(d)
	changeVNI := current != nil && current.VNI != nil && d.HasChange("vni") &&
		(vni == "auto" && current.VNIMode == "manual" || vni != nil && vni != "auto" && vni != *current.VNI)
	for _, key := range []string{"importrt", "exportrt"} {
		if len(d.Get(key).([]interface{})) == 0 {
			continue
		}
		rows := routeTargetRows(d.Get(key))
		desired := make(map[string]bool)
		for _, raw := range rows {
			desired[raw.(map[string]interface{})["value"].(string)] = true
		}
		var existing []vpc.RouteTarget
		if current != nil && current.RouteTargets != nil {
			existing = current.RouteTargets.Import
			if key == "exportrt" {
				existing = current.RouteTargets.Export
			}
		}
		// Preserve generated rows from the current API response, including edited
		// peering values. They are not owned by the managed custom list.
		targets := make([]vpc.RouteTargetWrite, 0)
		generated := make(map[string]bool)
		for _, row := range existing {
			if row.Type != "default-RT" && row.Type != "vpc-peering-RT" {
				continue
			}
			value := row.Value
			changingDefault := changeVNI && row.Type == "default-RT" && row.Value == fmt.Sprintf("0:%d", *current.VNI)
			if changingDefault {
				if desired[row.Value] {
					return nil, fmt.Errorf("%s contains controller-generated default %q; remove this redundant target before changing vni so the controller can update it", key, row.Value)
				}
				if number, ok := vni.(int); ok {
					value = fmt.Sprintf("0:%d", number)
				}
			}
			autoDefault := changingDefault && vni == "auto"
			if value != row.Value || autoDefault {
				holder := false
				for _, other := range existing {
					if autoDefault {
						number, err := strconv.Atoi(strings.TrimPrefix(other.Value, "0:"))
						if other.Type != "custom-RT" || !strings.HasPrefix(other.Value, "0:") || err != nil || number < 1 || number > 16777215 {
							continue
						}
					} else {
						if other.ID == row.ID || other.Value != value {
							continue
						}
						holder = true
					}
					if other.Type == "custom-RT" && !desired[other.Value] {
						return nil, fmt.Errorf("%s custom target %q may become the new VNI default; remove or edit that target in a separate apply before changing vni", key, other.Value)
					}
				}
				if holder {
					continue
				}
			}
			generated[value] = true
			targets = append(targets, vpc.RouteTargetWrite{ID: row.ID, Value: row.Value})
		}
		// Manual POST replaces the complete direction; unlike auto POST it does
		// not retain the newly allocated default unless it is included explicitly.
		if current == nil {
			if number, ok := vni.(int); ok {
				value := fmt.Sprintf("0:%d", number)
				generated[value] = true
				targets = append(targets, vpc.RouteTargetWrite{Value: value})
			}
		}
		old, _ := d.GetChange(key)
		previous := routeTargetRows(old)
		custom := make([]vpc.RouteTargetWrite, len(rows))
		used := make(map[int]bool)
		// Reserve retained custom IDs before assigning any ID to an edited value.
		for i, raw := range rows {
			value := raw.(map[string]interface{})["value"].(string)
			custom[i].Value = value
			if generated[value] {
				continue
			}
			for _, row := range existing {
				if row.Type == "custom-RT" && row.Value == value && !used[row.ID] {
					custom[i].ID = row.ID
					used[row.ID] = true
					break
				}
			}
		}
		for i, row := range custom {
			if generated[row.Value] {
				continue
			}
			if row.ID == 0 && i < len(previous) {
				prior := previous[i].(map[string]interface{})
				for _, existingRow := range existing {
					if existingRow.ID == prior["id"] && existingRow.Type == "custom-RT" && !used[existingRow.ID] {
						row.ID = existingRow.ID
						used[row.ID] = true
						break
					}
				}
			}
			targets = append(targets, row)
		}
		if key == "importrt" {
			routes.Import = &targets
		} else {
			routes.Export = &targets
		}
	}
	if routes.Import == nil && routes.Export == nil {
		return nil, nil
	}
	return routes, nil
}

func flattenRouteTargets(rows []vpc.RouteTarget, previous []interface{}, dataSource bool) []map[string]interface{} {
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
				if !used[rows[i].ID] && rows[i].ID == prior["id"] && (dataSource || rows[i].Type == "custom-RT") {
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
		if !used[row.ID] && (dataSource || row.Type == "custom-RT") {
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
		if err := d.Set(direction.key, flattenRouteTargets(direction.rows, routeTargetRows(d.Get(direction.key)), dataSource)); err != nil {
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
