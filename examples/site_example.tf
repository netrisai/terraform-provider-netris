resource "netris_site" "santa-clara" {
  name                 = "Santa Clara"
  publicasn            = 65001
  rohasn               = 65502
  vmasn                = 65503
  rohroutingprofile    = "default"
  sitemesh             = "hub"
  acldefaultpolicy     = "permit"
  # switchfabric         = "equinix_metal"
  # # vlanrange            = "2-3999"
  # equinixprojectid     = "xxxx"
  # equinixprojectapikey = "yyyy"
  # equinixlocation      = "sv"
}
