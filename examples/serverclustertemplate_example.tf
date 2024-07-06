resource "netris_serverclustertemplate" "my-serverclustertemplate1" {
  name  = "my-serverclustertemplate1"
  vnets = jsonencode(
[
    {
        "postfix": "East-West",
        "serverNics": [
            "eth1",
            "eth2",
            "eth3",
            "eth4",
            "eth5",
            "eth6",
            "eth7",
            "eth8"
        ],
        "type": "l3vpn",
        "vlan": "untagged",
        "vlanID": "auto"
    },
    {
        "ipv4Gateway": "192.168.0.254/24",
        "postfix": "North-South-in-band-and-storage",
        "serverNics": [
            "eth9",
            "eth10"
        ],
        "type": "l2vpn",
        "vlan": "untagged",
        "vlanID": "auto"
    },
    {
        "ipv4Gateway": "192.168.10.254/24",
        "postfix": "OOB-Management",
        "serverNics": [
            "eth13"
        ],
        "type": "l2vpn",
        "vlan": "untagged",
        "vlanID": "auto"
    }
]
  )
}
