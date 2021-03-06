
# Image URL to use all building/pushing image targets
OS := $(shell uname )
VERSION := $(shell git describe --tags --always --dirty)
REPOSITORY := wonderix/shalm
IMG ?= ${REPOSITORY}:${VERSION}
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: manager

# Run tests
test: generate fmt vet 
	go test ./... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet 
	go run ./main.go

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: 
	go run ./main.go apply charts/shalm


# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths="./..."

# Build the docker image
docker-build: test
	docker build . -t ${IMG}
	docker tag ${IMG} ${REPOSITORY}:latest

# Push the docker image
docker-push:
	docker push ${IMG}
	docker push ${REPOSITORY}:latest

chart:
	rm -rf /tmp/shalm
	cp -r charts/shalm /tmp/shalm
ifeq ($(OS),Darwin)
	sed -i '' -e 's|version:.*|version: ${VERSION}|g' /tmp/shalm/Chart.yaml
	sed -i '' -e 's|image: wonderix/shalm:.*|image: wonderix/shalm:${VERSION}|g' /tmp/shalm/ytt/deployment.yaml
else
	sed -i -e 's|version:.*|version: ${VERSION}|g' /tmp/shalm/Chart.yaml
	sed -i -e 's|image: wonderix/shalm:.*|image: wonderix/shalm:${VERSION}|g' /tmp/shalm/ytt/deployment.yaml
endif
	mkdir -p bin
	cd bin && go run .. package /tmp/shalm

shalm::
	go build -ldflags "-X github.com/kramerul/shalm/cmd.version=${VERSION}" -o bin/shalm . 

binaries:
	mkdir -p bin
	cd bin; \
	for GOOS in linux darwin windows; do \
	  CGO_ENABLED=0 GOOS=$$GOOS GOARCH=amd64 GO111MODULE=on go build -ldflags "-X github.com/kramerul/shalm/cmd.version=${VERSION}" -o shalm ..; \
		tar czf shalm-binary-$$GOOS.tgz shalm; \
	done


# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.4 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif
