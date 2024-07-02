resource "netris_serverclustertemplate" "my-serverclustertemplate1" {
  template = <<EOF
{
    "name": "GPU3-cluster-template-terraform",
    "vnets": [
        {
            "postfix": "East-West",
            "type": "l3vpn",
            "vlan": "untagged",
            "vlanID": "auto",
            "serverNics": [
                "eth1",
                "eth2",
                "eth3",
                "eth4",
                "eth5",
                "eth6",
                "eth7",
                "eth8"
            ]
        },
        {
            "postfix": "North-South-in-band-and-storage",
            "type": "l2vpn",
            "vlan": "untagged",
            "vlanID": "auto",
            "serverNics": [
                "eth9",
                "eth10"
            ],
            "ipv4Gateway": "192.168.0.254/24"
        },
        {
            "postfix": "OOB-Management",
            "type": "l2vpn",
            "vlan": "untagged",
            "vlanID": "auto",
            "serverNics": [
                "eth13"
            ],
            "ipv4Gateway": "192.168.10.254/24"
        }
    ]
}
EOF
}
