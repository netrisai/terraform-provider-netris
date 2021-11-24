resource "netris_site" "site-terraform-test" {
    name = "site-terraform-test"
    publicasn = 1234
    rohasn = 12345
    vmasn = 12346
    routingprofile = "full_table"
    sitemesh = "disabled"
    acldefaultpolicy = "deny"
}