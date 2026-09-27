HOSTNAME=registry.terraform.io
NAMESPACE=kepsic
NAME=zoneeu
BINARY=terraform-provider-${NAME}
VERSION=1.0.1
OS_ARCH=$(shell go env GOOS)_$(shell go env GOARCH)

default: build

build:
	go build -o ${BINARY}

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	cp ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}/${BINARY}_v${VERSION}

test:
	go test -v -cover ./...

testacc:
	TF_ACC=1 go test -v ./internal/provider/... -timeout 120m

fmt:
	go fmt ./...
	terraform fmt -recursive ./examples/

lint:
	golangci-lint run ./...

docs:
	go generate ./tools/...

generate:
	go generate ./...

clean:
	rm -f ${BINARY}

tidy:
	go mod tidy

.PHONY: build install test testacc fmt lint docs generate clean tidy

