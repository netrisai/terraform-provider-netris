resource "netris_servercluster" "my-servercluster1" {
  name    = "my-servercluster1"
  adminid = data.netris_tenant.admin.id
  siteid  = netris_site.santa-clara.id
  # vpcid      = netris_vpc.my-vpc.id
  templateid = netris_serverclustertemplate.my-serverclustertemplate1.id
  tags       = ["boo", "foo"]

  # vpcmapping {
  #   postfix = "East-West"
  #   id      = netris_vpc.my-vpc.id
  # }

  # vpcmapping {
  #   postfix = "North-South-in-band-and-storage"
  #   id      = 0
  #   group   = "cluster-private"
  # }

  # vpcmapping {
  #   postfix = "OOB-Management"
  #   id      = 0
  #   group   = "cluster-private"
  # }

  servers = [
    netris_server.my-server01.id,
    netris_server.my-server02.id,
  ]

  # Only needed when the referenced template has a VNet whose VLAN mode is
  # "specify". One "vlan" block per such VNet postfix.
  vlan {
    postfix = "East-West"
    vlan_id = 100
  }
}

resource "netris_servercluster" "dynamic-access" {
  name              = "dynamic-access"
  adminid           = 1
  siteid            = 1
  templateid        = 10
  servers           = [101, 102]
  dynamicvpcgroupid = netris_dynamic_vpc_group.access.id
}

resource "netris_servercluster" "mixed" {
  name       = "mixed"
  adminid    = 1
  siteid     = 1
  templateid = 11
  servers    = [103, 104]
  vpcmapping {
    postfix           = "access"
    dynamicvpcgroupid = netris_dynamic_vpc_group.access.id
  }
  vpcmapping {
    postfix = "storage"
    id      = 0
    group   = "storage"
    name    = "storage-vpc"
  }
}
