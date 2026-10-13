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

package dynamicvpcgroup

import (
	"fmt"
	"net"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf16"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/netrisai/netriswebapi/v2/types/dynamicvpcgroup"
)

func validateUsername(value interface{}, key string) ([]string, []error) {
	username := value.(string)
	if !regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._@-]{0,62}$`).MatchString(username) || strings.EqualFold(username, "DEFAULT") {
		return nil, []error{fmt.Errorf("%s must contain 1 to 63 ASCII letters, digits, dots, underscores, @ or hyphens, start with a letter or digit, and cannot be DEFAULT", key)}
	}
	return nil, nil
}

func validatePassword(value interface{}, key string) ([]string, []error) {
	password := value.(string)
	if password != "" && !regexp.MustCompile(`^[!#-\[\]-~]{8,128}$`).MatchString(password) {
		return nil, []error{fmt.Errorf("%s must contain 8 to 128 printable ASCII characters excluding spaces, double quotes and backslashes", key)}
	}
	return nil, nil
}

func validateSecret(value interface{}, key string) ([]string, []error) {
	if len(utf16.Encode([]rune(value.(string)))) > 512 {
		return nil, []error{fmt.Errorf("%s must contain at most 512 characters", key)}
	}
	return nil, nil
}

func validateCollections(radiusServer string, members, servers []interface{}) error {
	if len(members) == 0 {
		return fmt.Errorf("at least one member block is required")
	}
	if radiusServer == "netris-hosted" && len(servers) != 0 {
		return fmt.Errorf("netris-hosted groups cannot have server blocks")
	}
	if radiusServer == "third-party" && (len(servers) < 1 || len(servers) > 3) {
		return fmt.Errorf("third-party groups require one to three server blocks")
	}
	vpcs := map[int]bool{}
	usernames := map[string]bool{}
	for _, value := range members {
		member := value.(map[string]interface{})
		id := member["vpcid"].(int)
		if id <= 0 {
			return fmt.Errorf("member vpcid must be a positive integer")
		}
		if vpcs[id] {
			return fmt.Errorf("member blocks must have unique vpcid values")
		}
		vpcs[id] = true
		users := member["user"].(*schema.Set).List()
		if radiusServer == "third-party" && len(users) > 0 {
			return fmt.Errorf("third-party groups cannot have user blocks")
		}
		for _, value := range users {
			username := value.(map[string]interface{})["username"].(string)
			if usernames[username] {
				return fmt.Errorf("user blocks must have unique usernames across all members")
			}
			usernames[username] = true
		}
	}
	identities := map[string]bool{}
	for _, value := range servers {
		server := value.(map[string]interface{})
		identity := serverIdentity(server["host"].(string), server["authport"].(int))
		if identities[identity] {
			return fmt.Errorf("server blocks must have unique host/authport pairs")
		}
		identities[identity] = true
	}
	return nil
}

func serverIdentity(host string, port int) string {
	return net.JoinHostPort(strings.ToLower(host), strconv.Itoa(port))
}

func expandTags(d *schema.ResourceData) []string {
	tags := []string{}
	for _, tag := range d.Get("tags").(*schema.Set).List() {
		tags = append(tags, tag.(string))
	}
	return tags
}

func usersByUsername(members []interface{}) map[string]map[string]interface{} {
	users := map[string]map[string]interface{}{}
	for _, value := range members {
		for _, value := range value.(map[string]interface{})["user"].(*schema.Set).List() {
			user := value.(map[string]interface{})
			users[user["username"].(string)] = user
		}
	}
	return users
}

func expandMembers(members, old []interface{}) ([]dynamicvpcgroup.MemberWrite, error) {
	previous := usersByUsername(old)
	result := []dynamicvpcgroup.MemberWrite{}
	for _, value := range members {
		member := value.(map[string]interface{})
		item := dynamicvpcgroup.MemberWrite{
			VPC:   dynamicvpcgroup.IDName{ID: member["vpcid"].(int)},
			Users: []dynamicvpcgroup.UserWrite{},
		}
		for _, value := range member["user"].(*schema.Set).List() {
			user := value.(map[string]interface{})
			write := dynamicvpcgroup.UserWrite{
				Username: user["username"].(string),
				Password: user["password"].(string),
			}
			if _, errors := validateUsername(write.Username, "username"); len(errors) > 0 {
				return nil, errors[0]
			}
			if _, errors := validatePassword(write.Password, "password"); len(errors) > 0 {
				return nil, errors[0]
			}
			if retained, ok := previous[write.Username]; ok {
				write.ID = retained["id"].(int)
				if write.Password == retained["password"].(string) {
					write.Password = ""
				}
			}
			if write.ID == 0 && write.Password == "" {
				return nil, fmt.Errorf("a password is required for each new username")
			}
			item.Users = append(item.Users, write)
		}
		sort.Slice(item.Users, func(i, j int) bool { return item.Users[i].Username < item.Users[j].Username })
		result = append(result, item)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].VPC.ID < result[j].VPC.ID })
	return result, nil
}

func expandServers(servers, old []interface{}, create bool) ([]dynamicvpcgroup.ServerWrite, error) {
	previous := map[string]map[string]interface{}{}
	for _, value := range old {
		server := value.(map[string]interface{})
		previous[serverIdentity(server["host"].(string), server["authport"].(int))] = server
	}
	result := []dynamicvpcgroup.ServerWrite{}
	for _, value := range servers {
		server := value.(map[string]interface{})
		item := dynamicvpcgroup.ServerWrite{
			Host:     server["host"].(string),
			AuthPort: server["authport"].(int),
			AcctPort: server["acctport"].(int),
			Secret:   server["secret"].(string),
		}
		if net.ParseIP(item.Host) == nil {
			return nil, fmt.Errorf("server host must be an IPv4 or IPv6 address")
		}
		if item.AuthPort < 1 || item.AuthPort > 65535 || item.AcctPort < 1 || item.AcctPort > 65535 {
			return nil, fmt.Errorf("server ports must be between 1 and 65535")
		}
		if _, errors := validateSecret(item.Secret, "server secret"); len(errors) > 0 {
			return nil, errors[0]
		}
		if create && item.Secret == "" {
			return nil, fmt.Errorf("a secret is required for each server on create")
		}
		if retained, ok := previous[serverIdentity(item.Host, item.AuthPort)]; ok && item.Secret == retained["secret"].(string) {
			item.Secret = ""
		}
		// On update the backend also permits same-position secret preservation for edited endpoints.
		result = append(result, item)
	}
	return result, nil
}

func flatten(d *schema.ResourceData, item *dynamicvpcgroup.DynamicVPCGroup, credentials, newUsers bool) error {
	previousUsers := map[string]map[string]interface{}{}
	previousServers := map[string]map[string]interface{}{}
	if credentials {
		previousUsers = usersByUsername(d.Get("member").(*schema.Set).List())
		for _, value := range d.Get("server").([]interface{}) {
			server := value.(map[string]interface{})
			previousServers[serverIdentity(server["host"].(string), server["authport"].(int))] = server
		}
	}
	sort.Slice(item.Members, func(i, j int) bool { return item.Members[i].VPC.ID < item.Members[j].VPC.ID })
	members := []map[string]interface{}{}
	for _, member := range item.Members {
		sort.Slice(member.Users, func(i, j int) bool { return member.Users[i].Username < member.Users[j].Username })
		users := []map[string]interface{}{}
		for _, user := range member.Users {
			entry := map[string]interface{}{
				"id":       user.ID,
				"username": user.Username,
			}
			if credentials {
				entry["password"] = ""
				if previous, ok := previousUsers[user.Username]; ok && (previous["id"].(int) == user.ID || newUsers && previous["id"].(int) == 0) {
					entry["password"] = previous["password"]
				}
			}
			users = append(users, entry)
		}
		vni := 0
		if member.VPC.VNI != nil {
			vni = *member.VPC.VNI
		}
		members = append(members, map[string]interface{}{
			"vpcid":   member.VPC.ID,
			"vpcname": member.VPC.Name,
			"vni":     vni,
			"vrfname": member.VPC.VrfName,
			"user":    users,
		})
	}
	servers := []map[string]interface{}{}
	for _, server := range item.Servers {
		entry := map[string]interface{}{
			"host":     server.Host,
			"authport": server.AuthPort,
			"acctport": server.AcctPort,
		}
		if credentials {
			entry["secret"] = ""
			if previous, ok := previousServers[serverIdentity(server.Host, server.AuthPort)]; ok {
				entry["secret"] = previous["secret"]
			}
		}
		servers = append(servers, entry)
	}
	sort.Slice(item.RadiusClients, func(i, j int) bool {
		if item.RadiusClients[i].NASIdentifier == item.RadiusClients[j].NASIdentifier {
			return item.RadiusClients[i].Switch.ID < item.RadiusClients[j].Switch.ID
		}
		return item.RadiusClients[i].NASIdentifier < item.RadiusClients[j].NASIdentifier
	})
	clients := []map[string]interface{}{}
	for _, client := range item.RadiusClients {
		clients = append(clients, map[string]interface{}{
			"nasidentifier": client.NASIdentifier,
			"nasip":         client.NASIP,
			"switchid":      client.Switch.ID,
			"switchname":    client.Switch.Name,
		})
	}
	err := d.Set("name", item.Name)
	if err != nil {
		return err
	}
	err = d.Set("resourcegroupid", item.ResourceGroup.ID)
	if err != nil {
		return err
	}
	err = d.Set("radiusserver", item.RadiusServer)
	if err != nil {
		return err
	}
	err = d.Set("tags", item.Tags)
	if err != nil {
		return err
	}
	err = d.Set("displayid", item.DisplayID)
	if err != nil {
		return err
	}
	err = d.Set("member", members)
	if err != nil {
		return err
	}
	err = d.Set("server", servers)
	if err != nil {
		return err
	}
	err = d.Set("radiusclients", clients)
	if err != nil {
		return err
	}
	d.SetId(strconv.Itoa(item.ID))
	return nil
}
