Terraform Provider for Netris
==================
[![GitHub release](https://img.shields.io/github/tag/netrisai/terraform-provider-netris.svg?label=release)](https://github.com/netrisai/terraform-provider-netris/releases/latest)
[![Actions Status](https://github.com/netrisai/terraform-provider-netris/workflows/release/badge.svg)](https://github.com/netrisai/terraform-provider-netris/actions)
- Website: https://www.terraform.io
- Documentation: https://registry.terraform.io/providers/netrisai/netris/latest

<img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" width="600px">

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.13.x
-	[Go](https://golang.org/doc/install) 1.16.x (to build the provider plugin)

Using the provider
----------------------

See the [Netris Provider documentation](https://registry.terraform.io/providers/netrisai/netris/latest/docs) to get started using the Netris provider.

Manual Build and Install
------------
If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.16+ is *required*).

To compile the provider, run `make install`. This will build the provider and put the provider binary in the `~/.terraform.d/plugins/` directory.

```sh
make install
```
