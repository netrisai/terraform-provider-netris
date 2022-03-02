resource "netris_portgroup" "my_portgroup" {
  name  = "my_portgroup"
  ports = ["22", "1024-2048", "33554"]
}
