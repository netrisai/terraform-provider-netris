TEST?=$$(go list ./... | grep -v 'vendor')
HOSTNAME=registry.terraform.io
NAMESPACE=netrisai
NAME=netris
BINARY=terraform-provider-${NAME}
VERSION=1.0.3
OS_ARCH=darwin_arm64
WORKDIRECTORY=examples

default: install

build:
	go build -o ${BINARY}

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

uninstall:
	rm -rf ${WORKDIRECTORY}/.terraform*

uninstall-all:
	rm -rf ${WORKDIRECTORY}/.terraform*
	rm -rf ${WORKDIRECTORY}/*.tfstate*

init: install
	cd ${WORKDIRECTORY} && terraform init

apply: init
	cd ${WORKDIRECTORY} && terraform apply -auto-approve

plan: init
	cd ${WORKDIRECTORY} && terraform plan

destroy:
	cd ${WORKDIRECTORY} && terraform destroy -auto-approve

reapply: destroy uninstall apply

test: 
	go test -i $(TEST) || exit 1                                                   
	echo $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4                    

testacc: 
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m   
