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

package acl

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/netrisai/netriswebapi/http"
	"github.com/netrisai/netriswebapi/v1/types/acl"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	api "github.com/netrisai/netriswebapi/v2"
)

func Resource() *schema.Resource {
	return &schema.Resource{
		Description: "Creates and manages ACLs",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Unique name for the ACL entry.",
			},
			"action": {
				Required:    true,
				Type:        schema.TypeString,
				Description: "Permit or Deny forwarding of matched packets. Valid values are `permit` and `deny`.",
			},
			"comment": {
				Default:     "",
				Optional:    true,
				Type:        schema.TypeString,
				Description: "Descriptive comment, commonly used for approval workflows.",
			},
			"established": {
				ValidateFunc: validateEstablished,
				Default:      1,
				Optional:     true,
				Type:         schema.TypeInt,
				Description:  "For TCP, also match reverse packets except with TCP SYN flag. For non-TCP, also generate a reverse rule with swapped source/destination. Valid values are `0` and `1`. Default value is `1`",
			},
			"icmptype": {
				ValidateFunc: validateICMP,
				Default:      1,
				Optional:     true,
				Type:         schema.TypeInt,
				Description:  "Custom IPv4 ICMP code. Valid values should be in range 1-37 according to RFC 1700. Default value is `1`",
			},
			"proto": {
				ValidateFunc: validateProto,
				Required:     true,
				Type:         schema.TypeString,
				Description:  "IP protocol to match. Valid values are `all`, `ip`, `tcp`, `udp`, `icmp`, `icmpv6`.",
			},
			"reverse": {
				Default:     true,
				Optional:    true,
				Type:        schema.TypeBool,
				Description: "For TCP, also match reverse packets except with TCP SYN flag. For non-TCP, also generate a reverse rule with swapped source/destination. Default value is `true`",
			},
			"srcprefix": {
				ValidateFunc: validateIPPrefix,
				Required:     true,
				Type:         schema.TypeString,
				Description:  "Source IPv4/IPv6 address. Example `192.0.2.0/24`",
			},
			"srcportfrom": {
				ValidateFunc: validatePort,
				Optional:     true,
				Type:         schema.TypeInt,
				Description:  "Source port from. Valid values should be in range 1-65535",
			},
			"srcportto": {
				ValidateFunc: validatePort,
				Optional:     true,
				Type:         schema.TypeInt,
				Description:  "Source port to. Valid values should be in range 1-65535",
			},
			"srcportgroup": {
				Optional:    true,
				Type:        schema.TypeString,
				Description: "Match source ports on a group of ports. Valid value name of ACL Port Group",
			},
			"dstprefix": {
				ValidateFunc: validateIPPrefix,
				Required:     true,
				Type:         schema.TypeString,
				Description:  "Destination IPv4/IPv6 address. Example `0.0.0.0/0`",
			},
			"dstportfrom": {
				ValidateFunc: validatePort,
				Optional:     true,
				Type:         schema.TypeInt,
				Description:  "Destination port from. Valid values should be in range 1-65535",
			},
			"dstportto": {
				ValidateFunc: validatePort,
				Optional:     true,
				Type:         schema.TypeInt,
				Description:  "Destination port to. Valid values should be in range 1-65535",
			},
			"dstportgroup": {
				Optional:    true,
				Type:        schema.TypeString,
				Description: "Match destination ports on a group of ports. Valid value name of ACL Port Group",
			},
			"validuntil": {
				ValidateFunc: validateDate,
				Optional:     true,
				Type:         schema.TypeString,
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

	name := d.Get("name").(string)
	action := d.Get("action").(string)
	comment := d.Get("comment").(string)

	established := d.Get("established").(int)

	icmptype := d.Get("icmptype").(int)
	proto := d.Get("proto").(string)

	if proto != "tcp" {
		established = 0
	}

	reverse := "yes"
	if r := d.Get("reverse").(bool); !r {
		reverse = "no"
	}

	srcprefix := d.Get("srcprefix").(string)
	srcportfrom := d.Get("srcportfrom").(int)
	srcportto := d.Get("srcportto").(int)

	srcPgID := 0
	if s := d.Get("srcportgroup").(string); s != "" {
		if pg, ok := getPortGroupByName(s, clientset); ok {
			srcPgID = pg.ID
		} else {
			return fmt.Errorf("couldn't find port group %s", s)
		}
	}

	dstprefix := d.Get("dstprefix").(string)
	dstportfrom := d.Get("dstportfrom").(int)
	dstportto := d.Get("dstportto").(int)

	dstPgID := 0
	if s := d.Get("dstportgroup").(string); s != "" {
		if pg, ok := getPortGroupByName(s, clientset); ok {
			dstPgID = pg.ID
		} else {
			return fmt.Errorf("couldn't find port group %s", s)
		}
	}

	aclW := &acl.ACLw{
		Name:        name,
		Action:      action,
		Comment:     comment,
		Established: established,
		ICMPType:    icmptype,
		Proto:       proto,
		Reverse:     reverse,
		SrcPrefix:   srcprefix,
		DstPrefix:   dstprefix,
	}

	if srcPgID > 0 {
		aclW.SrcPortGroup = srcPgID
	} else {
		aclW.SrcPortFrom = srcportfrom
		aclW.SrcPortTo = srcportto
	}

	if dstPgID > 0 {
		aclW.DstPortGroup = dstPgID
	} else {
		aclW.DstPortFrom = dstportfrom
		aclW.DstPortTo = dstportto
	}

	if v := d.Get("validuntil").(string); v != "" {
		aclW.ValidUntil = v
	}

	js, _ := json.Marshal(aclW)
	log.Println("[DEBUG]", string(js))

	reply, err := clientset.ACL().Add(aclW)
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
	var acl *acl.ACL
	acls, err := clientset.ACL().Get()
	if err != nil {
		return err
	}
	id, _ := strconv.Atoi(d.Id())
	for _, a := range acls {
		if a.ID == id {
			acl = a
			break
		}
	}

	if acl == nil {
		return nil
	}

	d.SetId(strconv.Itoa(acl.ID))
	err = d.Set("name", acl.Name)
	if err != nil {
		return err
	}
	err = d.Set("action", acl.Action)
	if err != nil {
		return err
	}
	err = d.Set("comment", acl.Comment)
	if err != nil {
		return err
	}
	if acl.Protocol == "tcp" {
		err = d.Set("established", acl.Established)
		if err != nil {
			return err
		}
	}
	err = d.Set("proto", acl.Protocol)
	if err != nil {
		return err
	}
	var reverse bool
	if acl.Reverse == "yes" {
		reverse = true
	}
	err = d.Set("reverse", reverse)
	if err != nil {
		return err
	}
	err = d.Set("srcprefix", fmt.Sprintf("%s/%d", acl.SrcPrefix, acl.SrcLength))
	if err != nil {
		return err
	}
	pgName := ""
	if pg, ok := getPortGroupByID(acl.SrcPortGroup, clientset); ok {
		pgName = pg.Name
	}
	if pgName != "" || d.Get("srcportgroup").(string) != "" {
		err = d.Set("srcportgroup", pgName)
		if err != nil {
			return err
		}
	}

	if acl.SrcPortFrom != 0 || d.Get("srcportfrom").(int) != 0 {
		err = d.Set("srcportfrom", acl.SrcPortFrom)
		if err != nil {
			return err
		}
	}
	if acl.SrcPortTo != 0 || d.Get("srcportto").(int) != 0 {
		err = d.Set("srcportto", acl.SrcPortTo)
		if err != nil {
			return err
		}
	}
	err = d.Set("dstprefix", fmt.Sprintf("%s/%d", acl.DstPrefix, acl.DstLength))
	if err != nil {
		return err
	}
	pgName = ""
	if pg, ok := getPortGroupByID(acl.DstPortGroup, clientset); ok {
		pgName = pg.Name
	}
	if pgName != "" || d.Get("dstportgroup").(string) != "" {
		err = d.Set("dstportgroup", pgName)
		if err != nil {
			return err
		}
	}
	if acl.DstPortFrom != 0 || d.Get("dstportfrom").(int) != 0 {
		err = d.Set("dstportfrom", acl.DstPortFrom)
		if err != nil {
			return err
		}
	}
	if acl.DstPortTo != 0 || d.Get("dstportto").(int) != 0 {
		err = d.Set("dstportto", acl.DstPortTo)
		if err != nil {
			return err
		}
	}

	if acl.ValidUntil != "" {
		if v := d.Get("validuntil").(string); v != "" {
			valMili, err := strconv.Atoi(acl.ValidUntil)
			if err != nil {
				return err
			}
			aclTime := time.UnixMilli(int64(valMili))
			aclStamp := aclTime.UnixMilli()
			terrStamp, err := time.Parse(time.RFC3339, v)
			if err != nil {
				return err
			}
			if aclStamp != terrStamp.UnixMilli() {
				err = d.Set("validuntil", aclTime.Format(time.RFC3339))
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func resourceUpdate(d *schema.ResourceData, m interface{}) error {
	clientset := m.(*api.Clientset)

	name := d.Get("name").(string)
	action := d.Get("action").(string)
	comment := d.Get("comment").(string)

	established := d.Get("established").(int)

	icmptype := d.Get("icmptype").(int)
	proto := d.Get("proto").(string)

	if proto != "tcp" {
		established = 0
	}

	reverse := "yes"
	if r := d.Get("reverse").(bool); !r {
		reverse = "no"
	}

	srcprefix := d.Get("srcprefix").(string)
	srcportfrom := d.Get("srcportfrom").(int)
	srcportto := d.Get("srcportto").(int)

	srcPgID := 0
	if s := d.Get("srcportgroup").(string); s != "" {
		if pg, ok := getPortGroupByName(s, clientset); ok {
			srcPgID = pg.ID
		} else {
			return fmt.Errorf("couldn't find port group %s", s)
		}
	}

	dstprefix := d.Get("dstprefix").(string)
	dstportfrom := d.Get("dstportfrom").(int)
	dstportto := d.Get("dstportto").(int)

	dstPgID := 0
	if s := d.Get("dstportgroup").(string); s != "" {
		if pg, ok := getPortGroupByName(s, clientset); ok {
			dstPgID = pg.ID
		} else {
			return fmt.Errorf("couldn't find port group %s", s)
		}
	}
	id, _ := strconv.Atoi(d.Id())
	aclW := &acl.ACLw{
		ID:          id,
		Name:        name,
		Action:      action,
		Comment:     comment,
		Established: established,
		ICMPType:    icmptype,
		Proto:       proto,
		Reverse:     reverse,
		SrcPrefix:   srcprefix,
		DstPrefix:   dstprefix,
	}

	if srcPgID > 0 {
		aclW.SrcPortGroup = srcPgID
	} else {
		aclW.SrcPortFrom = srcportfrom
		aclW.SrcPortTo = srcportto
	}

	if dstPgID > 0 {
		aclW.DstPortGroup = dstPgID
	} else {
		aclW.DstPortFrom = dstportfrom
		aclW.DstPortTo = dstportto
	}

	if v := d.Get("validuntil").(string); v != "" {
		aclW.ValidUntil = v
	}
	js, _ := json.Marshal(aclW)
	log.Println("[DEBUG] bgpUpdate", string(js))

	reply, err := clientset.ACL().Update(aclW)
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
	reply, err := clientset.ACL().Delete(id)
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
	aclID, _ := strconv.Atoi(d.Id())

	acls, err := clientset.ACL().Get()
	if err != nil {
		return false, err
	}

	for _, acl := range acls {
		if aclID == acl.ID {
			return true, nil
		}
	}

	return false, nil
}

func resourceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	clientset := m.(*api.Clientset)

	acls, _ := clientset.ACL().Get()
	name := d.Id()
	for _, acl := range acls {
		if acl.Name == name {
			d.SetId(strconv.Itoa(acl.ID))
			return []*schema.ResourceData{d}, nil
		}
	}

	return []*schema.ResourceData{d}, nil
}
