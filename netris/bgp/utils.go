package bgp

import (
	"fmt"

	api "github.com/netrisai/netriswebapi/v2"
	"github.com/netrisai/netriswebapi/v2/types/bgp"
)

func findPort(clientset *api.Clientset, siteID int, portName string) (*bgp.EBGPPort, bool) {
	ports, err := clientset.BGP().GetPorts(siteID)
	if err != nil {
		return nil, false
	}
	for _, port := range ports {
		if fmt.Sprintf("%s@%s", port.Port, port.SwitchName) == portName {
			return port, true
		}
	}
	return nil, false
}

func findVNetByName(clientset *api.Clientset, name string) (*bgp.EBGPVNet, bool) {
	vnets, err := clientset.BGP().GetVNets()
	if err != nil {
		return nil, false
	}
	for _, vnet := range vnets {
		if vnet.Name == name {
			return vnet, true
		}
	}
	return nil, false
}

func findSwitchByName(clientset *api.Clientset, siteID int, name string) (*bgp.EBGPSwitch, bool) {
	switches, err := clientset.BGP().GetSwitches(siteID)
	if err != nil {
		return nil, false
	}
	for _, item := range switches {
		if item.Location == name {
			return item, true
		}
	}

	return nil, false
}
