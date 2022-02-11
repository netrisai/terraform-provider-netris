# Netris Terraform Provider 

This directory stores terraform examples for provisioning a demo infrastructure.


Usage
------

1. Download and install the [terraform](https://www.terraform.io/downloads)

2. Clone repo
```sh
git clone https://github.com/netrisai/terraform-provider-netris.git
```

3. Go to examples directory 
```sh
cd terraform-provider-netris/examples
```

4. Specify controller address and credentials in `terraform.tf` file or set environment variables
```sh
export NETRIS_ADDRESS=http://example.com
export NETRIS_LOGIN=netris
export NETRIS_PASSWORD=newNet0ps
```

5. Init terraform and apply
```sh
terraform init
terraform apply
```
