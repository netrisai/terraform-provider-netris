resource "netris_nat" "my_snat" {
  name       = "MY SNAT"
  comment    = "Terraform Test SNAT"
  state      = "enabled"
  siteid     = netris_site.santa-clara.id
  action     = "SNAT"
  protocol   = "all"
  srcaddress = "100.71.56.0/24"
  dstaddress = "0.0.0.0/0"
  snattoip   = "203.0.113.192"
  # snattopool = "203.0.113.192/26"
  depends_on = [netris_subnet.my-subnet-nat]
}

resource "netris_nat" "my_dnat" {
  name       = "MY DNAT"
  state      = "enabled"
  siteid     = netris_site.santa-clara.id
  action     = "DNAT"
  protocol   = "tcp"
  srcaddress = "0.0.0.0/0"
  srcport    = "1-65535"
  dstaddress = "203.0.113.193/32"
  dstport    = "8080"
  dnattoip   = "100.71.56.60/32"
  dnattoport = 80
  depends_on = [netris_subnet.my-subnet-nat]
}

resource "netris_nat" "my_snat_accept" {
  name       = "MY SNAT ACCEPT"
  state      = "enabled"
  siteid     = netris_site.santa-clara.id
  action     = "ACCEPT_SNAT"
  protocol   = "all"
  srcaddress = "100.71.56.0/24"
  dstaddress = "10.10.0.0/16"
}
