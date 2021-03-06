---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "netris_user Resource - terraform-provider-netris"
subcategory: ""
description: |-
  Creates and manages Users
---

# netris_user

Define a new user in Netris.

~> **Note:** User require `userrole`/`pgroup` to exist prior to resource creation. Use `depends_on` to set an explicit dependency on the `userrole`/`pgroup`/.

## Example Usages

```hcl
data "netris_tenant" "my-tenant" {
  name = "my-tenant"
}

resource "netris_user" "terrraform-user" {
  username = "terraform"
  fullname = "Terraform"
  email = "terraform@netris.ai"
  emailcc = "devops@netris.ai"
  phone = "6504570097"
  company = "Netris, Inc."
  position = "DevOps Engineer"
  userrole = ""
  pgroup = "my-group"
  tenants {
    id = -1
    edit = false
  }
  tenants {
    id = data.netris_tenant.my-tenant.id
    edit = true
  }
  depends_on = [
    netris_permission_group.my-group,
  ]
}
```


<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **email** (String) The email address of the user. Also used for system notifications and for password retrieval.
- **pgroup** (String) Name of Permission Group. User permissions for viewing and editing parts of the Netris Controller. (if User Role is not used).
- **username** (String) Unique username.
- **userrole** (String) Name of User Role. When using a User Role object to define RBAC (role-based access control), `pgroup` and `tenants` fields will be ignoring.

### Optional

- **company** (String) Company the user works for. Usually useful for multi-tenant systems where the company provides Netris Controller access to customers.
- **emailcc** (String) Send copies of email notifications to this address.
- **fullname** (String) Full Name of the user.
- **phone** (String) User’s phone number.
- **position** (String) Position within the company.
- **tenants** (Block List) The block of tenants. (if User Role is not used). (see [below for nested schema](#nestedblock--tenants))

<a id="nestedblock--tenants"></a>
### Nested Schema for `tenants`

Optional:

- **id** (Number) Reference to tenant resource ID. `-1` means `All tenants`
- **edit** (Boolean) When `true` means Full access when `false` - Read-only. Default value: `true` 
