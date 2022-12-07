resource "netris_dhcp_option_set" "dhcp_option_set" {
  name         = "my-dhcp-option-set"
  description  = "My DHCP Option Set"
  domainsearch = "dhcp.example.local"
  dnsservers   = ["1.1.1.1", "8.8.8.8"]
  ntpservers   = ["0.pool.ntp.org", "132.163.96.5"]
  # leasetime    = 86400
  standardtoption {
    code  = 67
    value = "ipxe.efi"
  }
  customoption {
    code  = 239
    type  = "string"
    value = "http://192.0.2.55/script.sh"
  }
}
