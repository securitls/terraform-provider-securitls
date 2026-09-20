TEST?=$$(go list ./... | grep -v 'vendor')
HOSTNAME=registry.terraform.io
NAMESPACE=securitls
NAME=securitls
VERSION?=0.3.0

GOOS:=$(shell go env GOOS)
GOARCH:=$(shell go env GOARCH)

BINARY=terraform-provider-${NAME}_v${VERSION}.exe
PLUGIN_DIR=C:/Users/jacob/AppData/Roaming/terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${GOOS}_${GOARCH}

.PHONY: build test testacc

build:
	go build -o ${BINARY}
	mkdir -p "${PLUGIN_DIR}"
	mv -f "${BINARY}" "${PLUGIN_DIR}/${BINARY}"

test:
	go test ./...

testacc:
	TF_ACC=1 go test ./... -v -timeout 120m
