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

package port

import "github.com/netrisai/netriswebapi/v2/types/port"

var portDefault = &port.PortUpdate{
	AdminDown: "no",
	AutoNeg:   "none",
	Breakout:  "off",
	Duplex:    "full",
	Mtu:       9000,
	Speed:     "auto",
}

var speedMap = map[string]string{
	"auto": "auto",
	"1g":   "1000",
	"10g":  "10000",
	"25g":  "25000",
	"40g":  "40000",
	"50g":  "50000",
	"100g": "100000",
	"200g": "200000",
	"400g": "400000",
}

var speedMapReversed = map[string]string{
	"auto":   "auto",
	"1000":   "1g",
	"10000":  "10g",
	"25000":  "25g",
	"40000":  "40g",
	"50000":  "50g",
	"100000": "100g",
	"200000": "200g",
	"400000": "400g",
}
