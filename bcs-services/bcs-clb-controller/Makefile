# default
MAKE:=make
bmsf_edition?=inner_edition

# build info
BUILDTIME=$(shell date +%Y-%m-%dT%T%z)
GITHASH=$(shell git rev-parse  HEAD)
GITTAG=$(shell git describe --always)
VERSION=${GITTAG}-$(shell date +%y.%m.%d)
TVERSION=$(shell date +%y.%m.%d)

BINARYPATH=./build/k8s.${TVERSION}/bin
CERTSPATH=./build/k8s.${TVERSION}/certs
CONFPATH=./build/k8s.${TVERSION}/conf
COMMONPATH=./build/k8s.${TVERSION}/common
IMAGEPATH =./build/k8s.${TVERSION}/images

LDFLAG=-ldflags "-X bk-bcs/bcs-services/bcs-clb-controller/pkg/version.Version=${VERSION} \
 -X bk-bcs/bcs-services/bcs-clb-controller/pkg/version.BuildTime=${BUILDTIME} \
 -X bk-bcs/bcs-services/bcs-clb-controller/pkg/version.GitHash=${GITHASH} \
 -X bk-bcs/bcs-services/bcs-clb-controller/pkg/version.Tag=${GITTAG}"

# Image URL to use all building/pushing image targets
IMG ?= controller:latest

all: test manager

# Run tests
test: generate fmt vet manifests
	go test ./pkg/... ./cmd/... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager bk-bcs/bcs-services/bcs-clb-controller/cmd/manager

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet
	go run ./cmd/manager/main.go

# Install CRDs into a cluster
install: manifests
	kubectl apply -f config/crds

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	kubectl apply -f config/crds
	kustomize build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests:
	go run ../../vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go all
	cp ./config/crds/* ../../docs/features/bcs-clb-controller/deploy/helm/clb-controller-init/templates/

# Run go fmt against code
fmt:
	go fmt ./pkg/... ./cmd/...

# Run go vet against code
vet:
	go vet ./pkg/... ./cmd/...

# Generate code
generate:
	go generate ./pkg/... ./cmd/...

# Build the docker image
# docker-build: test
# 	docker build . -t ${IMG}
# 	@echo "updating kustomize image patch file for manager resource"
# 	sed -i'' -e 's@image: .*@image: '"${IMG}"'@' ./config/default/manager_image_patch.yaml

# Push the docker image
docker-push:
	docker push ${IMG}

# Build clb controller
binpath:
	mkdir -p ${BINARYPATH}

clb-controller:
	GOOS=linux go build ${LDFLAG} -o ${BINARYPATH}/clb-controller main.go
	cp ${BINARYPATH}/bcs-clb-controller ./docker/clb-controller
