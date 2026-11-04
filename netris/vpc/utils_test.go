/*
Copyright 2026. Netris, Inc.

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
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/netrisai/netriswebapi/v2/types/vpc"
)

func testRouteTargetBlock(values ...string) []interface{} {
	rows := make([]interface{}, 0, len(values))
	for _, value := range values {
		rows = append(rows, map[string]interface{}{"value": value})
	}
	return []interface{}{map[string]interface{}{"target": rows}}
}

func TestRouteTargetDiff(t *testing.T) {
	const unknown = "74D93920-ED26-11E3-AC10-0800200C9A66"
	for _, test := range []struct {
		name             string
		imports, exports interface{}
		wantError        bool
	}{
		{"omitted", nil, nil, false},
		{"empty", testRouteTargetBlock(), nil, false},
		{"duplicate import", testRouteTargetBlock("65000:1", "65000:1"), nil, true},
		{"duplicate export", nil, testRouteTargetBlock("65000:1", "65000:1"), true},
		{"same across directions", testRouteTargetBlock("65000:1"), testRouteTargetBlock("65000:1"), false},
		{"unknown wrapper", unknown, nil, false},
		{"unknown targets", []interface{}{map[string]interface{}{"target": unknown}}, nil, false},
		{"unknown leaves", testRouteTargetBlock(unknown, unknown), nil, false},
		{"known duplicate beside unknown", testRouteTargetBlock("65000:1", unknown, "65000:1"), nil, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			config := map[string]interface{}{"name": "test", "tenantid": 1, "importrt": test.imports, "exportrt": test.exports}
			_, err := Resource().Diff(nil, terraform.NewResourceConfigRaw(config), nil)
			if (err != nil) != test.wantError {
				t.Fatalf("Diff error = %v, want error %t", err, test.wantError)
			}
		})
	}
}

func TestExpandRouteTargetsCreate(t *testing.T) {
	for _, test := range []struct {
		name, vni string
		block     []interface{}
		want      []vpc.RouteTargetWrite
	}{
		{"auto custom", "auto", testRouteTargetBlock("65000:1"), []vpc.RouteTargetWrite{{Value: "65000:1"}}},
		{"auto empty", "auto", testRouteTargetBlock(), []vpc.RouteTargetWrite{}},
		{"manual empty", "7001", testRouteTargetBlock(), []vpc.RouteTargetWrite{{Value: "0:7001"}}},
		{"manual alias", "7001", testRouteTargetBlock("0:7001", "65000:1"), []vpc.RouteTargetWrite{{Value: "0:7001"}, {Value: "65000:1"}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, Resource().Schema, map[string]interface{}{"vni": test.vni, "importrt": test.block})
			routes, err := expandRouteTargets(d, nil)
			if err != nil || routes == nil || routes.Import == nil || !reflect.DeepEqual(*routes.Import, test.want) || routes.Export != nil {
				t.Fatalf("routes = %#v, error = %v; want import %#v and omitted export", routes, err, test.want)
			}
		})
	}
	d := schema.TestResourceDataRaw(t, Resource().Schema, nil)
	if routes, err := expandRouteTargets(d, nil); routes != nil || err != nil {
		t.Fatalf("omitted routes = %#v, error = %v", routes, err)
	}
}

func TestExpandRouteTargetsUpdate(t *testing.T) {
	vni := 7001
	defaultRT := vpc.RouteTarget{ID: 10, Value: "0:7001", Type: "default-RT"}
	peering := vpc.RouteTarget{ID: 20, Value: "64512:99", Type: "vpc-peering-RT"}
	custom := vpc.RouteTarget{ID: 30, Value: "65000:1", Type: "custom-RT"}
	second := vpc.RouteTarget{ID: 31, Value: "65000:2", Type: "custom-RT"}
	holder := vpc.RouteTarget{ID: 32, Value: "0:7002", Type: "custom-RT"}
	for _, test := range []struct {
		name, vni         string
		current           []vpc.RouteTarget
		previous, desired []string
		want              []vpc.RouteTargetWrite
		wantError         bool
	}{
		{"clear custom", "7001", []vpc.RouteTarget{defaultRT, peering, custom}, []string{"65000:1"}, nil, []vpc.RouteTargetWrite{{ID: 10, Value: "0:7001"}, {ID: 20, Value: "64512:99"}}, false},
		{"generated alias", "7001", []vpc.RouteTarget{defaultRT, peering}, []string{"0:7001"}, []string{"0:7001", "64512:99"}, []vpc.RouteTargetWrite{{ID: 10, Value: "0:7001"}, {ID: 20, Value: "64512:99"}}, false},
		{"generated position is not editable", "7001", []vpc.RouteTarget{defaultRT}, []string{"0:7001"}, []string{"65000:9"}, []vpc.RouteTargetWrite{{ID: 10, Value: "0:7001"}, {Value: "65000:9"}}, false},
		{"reserve retained custom before edit", "7001", []vpc.RouteTarget{custom, second}, []string{"65000:1", "65000:2"}, []string{"65000:2", "65000:9"}, []vpc.RouteTargetWrite{{ID: 31, Value: "65000:2"}, {Value: "65000:9"}}, false},
		{"custom edit", "7001", []vpc.RouteTarget{custom}, []string{"65000:1"}, []string{"65000:9"}, []vpc.RouteTargetWrite{{ID: 30, Value: "65000:9"}}, false},
		{"new default alias", "7002", []vpc.RouteTarget{defaultRT, custom}, []string{"65000:1"}, []string{"0:7002", "65000:1"}, []vpc.RouteTargetWrite{{ID: 10, Value: "0:7001"}, {ID: 30, Value: "65000:1"}}, false},
		{"old default pinned", "7002", []vpc.RouteTarget{defaultRT}, []string{"0:7001"}, []string{"0:7001"}, nil, true},
		{"retained holder", "7002", []vpc.RouteTarget{defaultRT, holder}, []string{"0:7002"}, []string{"0:7002"}, []vpc.RouteTargetWrite{{ID: 32, Value: "0:7002"}}, false},
		{"removed holder", "7002", []vpc.RouteTarget{defaultRT, holder}, []string{"0:7002"}, nil, nil, true},
		{"auto removed potential holder", "auto", []vpc.RouteTarget{defaultRT, holder}, []string{"0:7002"}, nil, nil, true},
		{"auto retains potential holder", "auto", []vpc.RouteTarget{defaultRT, holder}, []string{"0:7002"}, []string{"0:7002"}, []vpc.RouteTargetWrite{{ID: 10, Value: "0:7001"}, {ID: 32, Value: "0:7002"}}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			r := Resource()
			before := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{"name": "before", "tenantid": 1, "vni": "7001", "importrt": testRouteTargetBlock(test.previous...)})
			before.SetId("1")
			if err := before.Set("importrt", flattenRouteTargets(test.current, routeTargetRows(before.Get("importrt")), false)); err != nil {
				t.Fatal(err)
			}
			state := before.State()
			config := terraform.NewResourceConfigRaw(map[string]interface{}{"name": "after", "tenantid": 1, "vni": test.vni, "importrt": testRouteTargetBlock(test.desired...)})
			diff, err := r.Diff(state, config, nil)
			if err != nil {
				t.Fatal(err)
			}
			var routes *vpc.RouteTargetsWrite
			r.Update = func(d *schema.ResourceData, m interface{}) error {
				var err error
				routes, err = expandRouteTargets(d, &vpc.VPC{ID: 1, VNI: &vni, VNIMode: "manual", RouteTargets: &vpc.RouteTargets{Import: test.current}})
				return err
			}
			_, err = r.Apply(state, diff, nil)
			if (err != nil) != test.wantError {
				t.Fatalf("error = %v, want error %t", err, test.wantError)
			}
			if !test.wantError && (routes == nil || routes.Import == nil || !reflect.DeepEqual(*routes.Import, test.want)) {
				t.Fatalf("routes = %#v, want import %#v", routes, test.want)
			}
		})
	}
}

func TestFlattenRouteTargets(t *testing.T) {
	rows := []vpc.RouteTarget{{ID: 10, Value: "0:7001", Type: "default-RT"}, {ID: 20, Value: "64512:99", Type: "vpc-peering-RT"}, {ID: 30, Value: "65000:1", Type: "custom-RT"}}
	for _, test := range []struct {
		name       string
		previous   []interface{}
		dataSource bool
		want       []int
	}{
		{"resource custom only", nil, false, []int{30}},
		{"explicit generated alias", routeTargetRows(testRouteTargetBlock("64512:99")), false, []int{20, 30}},
		{"datasource complete", nil, true, []int{10, 20, 30}},
		{"changed generated alias is not silently rewritten", []interface{}{map[string]interface{}{"id": 10, "value": "0:7000", "type": "default-RT"}}, false, []int{30}},
	} {
		t.Run(test.name, func(t *testing.T) {
			result := flattenRouteTargets(rows, test.previous, test.dataSource)
			ids := []int{}
			for _, row := range result[0]["target"].([]map[string]interface{}) {
				ids = append(ids, row["id"].(int))
			}
			if !reflect.DeepEqual(ids, test.want) {
				t.Fatalf("IDs = %v, want %v", ids, test.want)
			}
		})
	}
}

func TestRouteTargetStateOwnership(t *testing.T) {
	item := &vpc.VPC{RouteTargets: &vpc.RouteTargets{
		Import: []vpc.RouteTarget{{ID: 10, Value: "0:7001", Type: "default-RT"}},
	}}
	for _, test := range []struct {
		name       string
		block      []interface{}
		wantBlocks int
	}{
		{"imported without configuration", nil, 0},
		{"empty managed block", testRouteTargetBlock(), 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, Resource().Schema, map[string]interface{}{"importrt": test.block})
			if err := flattenVNIAndRouteTargets(d, item, false); err != nil {
				t.Fatal(err)
			}
			if got := len(d.Get("importrt").([]interface{})); got != test.wantBlocks {
				t.Fatalf("blocks = %d, want %d", got, test.wantBlocks)
			}
			if got := routeTargetRows(d.Get("importrt")); len(got) != 0 {
				t.Fatalf("generated target entered managed state: %#v", got)
			}
		})
	}
}

func TestRouteTargetAndVNIValidation(t *testing.T) {
	for _, test := range []struct {
		value string
		valid bool
	}{
		{"0:7001", true}, {"65535:4294967295", true}, {"65536:65535", true}, {"192.0.2.1:65535", true},
		{"65536:65536", false}, {"192.0.2.1:65536", false}, {"065000:1", false}, {"65000:01", false}, {"invalid", false},
	} {
		if _, errors := validateRouteTarget(test.value, "target"); (len(errors) == 0) != test.valid {
			t.Errorf("route target %q errors = %v", test.value, errors)
		}
	}
	for _, test := range []struct {
		value string
		valid bool
	}{
		{"auto", true}, {"1", true}, {"16777215", true}, {"0", false}, {"16777216", false}, {"01", false}, {"", false},
	} {
		if _, errors := validateVNI(test.value, "vni"); (len(errors) == 0) != test.valid {
			t.Errorf("VNI %q errors = %v", test.value, errors)
		}
	}
}
