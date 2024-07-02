<a href="https://terraform.io">
    <img src=".github/terraform_logo.svg" alt="Terraform logo" title="Terraform" align="right" height="50" />
</a>

# Terraform Provider for Netris

[![GitHub release](https://img.shields.io/github/tag/netrisai/terraform-provider-netris.svg?label=release)](https://github.com/netrisai/terraform-provider-netris/releases/latest)
[![Actions Status](https://github.com/netrisai/terraform-provider-netris/workflows/release/badge.svg)](https://github.com/netrisai/terraform-provider-netris/actions)
- Website: https://www.terraform.io
- Documentation: https://registry.terraform.io/providers/netrisai/netris/latest

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.13.x +
-	[Go](https://golang.org/doc/install) 1.20.x (to build the provider plugin)

Using the provider
----------------------

See the [Netris Provider documentation](https://registry.terraform.io/providers/netrisai/netris/latest/docs) to get started using the Netris provider.

Compatibility with Netris-Controller
------------------------------------
  | Provider version | Controller version |
  | -----------------| -------------------|
  | `v1.X`           | `v3.0`             |
  | `v2.X`           | `v3.1+`            |
  | `v3.X`           | `v4.0+`            |
  | `v3.3`           | `v4.3+`            |


Manual Build and Install
------------
If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.20+ is *required*).

To compile the provider, make sure that `OS_ARCH` in the Makefile is correct and run `make install`. This will build the provider and put the provider binary in the `~/.terraform.d/plugins/` directory.

```sh
make install
```
