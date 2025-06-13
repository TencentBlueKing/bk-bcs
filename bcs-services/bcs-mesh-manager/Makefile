VERSION ?=  $(shell git describe --tags --match='v*' --dirty='.dirty')
REGISTRY = "mirrors.tencent.com"
REPO ?= $(REGISTRY)/bcs
TAG ?= $(VERSION)
GITTAG ?= $(shell git describe --tags --match='v*' --dirty='.dirty')
BUILDTIME ?= $(shell date +%Y-%m-%dT%T%z)
GITHASH ?= $(shell git rev-parse HEAD)

bcs_edition ?= inner_edition


.PHONY: proto build test tidy lint fmt vendor fmttool proto-fmt push-image

LDFLAG="-X github.com/Tencent/bk-bcs/bcs-common/common/static.ZookeeperClientUser=${bcs_zk_client_user} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/static.ZookeeperClientPwd=${bcs_zk_client_pwd} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/static.EncryptionKey=${bcs_encryption_key} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/static.ServerCertPwd=${bcs_server_cert_pwd} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/static.ClientCertPwd=${bcs_client_cert_pwd} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/static.LicenseServerClientCertPwd=${bcs_license_server_client_cert_pwd} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/static.BcsDefaultUser=${bcs_registry_default_user} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/static.BcsDefaultPasswd=${bcs_registry_default_pwd} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/version.BcsVersion=${VERSION} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/version.BcsBuildTime=${BUILDTIME} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/version.BcsGitHash=${GITHASH} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/version.BcsTag=${GITTAG} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/version.BcsEdition=${bcs_edition}"


# install fmt tool
fmttool:
	go install github.com/incu6us/goimports-reviser/v3@latest

# format
fmt:fmttool
	goimports-reviser -rm-unused -project-name github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager -excludes ./third_party/,./proto/,./vendor  ./...

# generate proto
proto:
	$(MAKE) -C proto/bcs-mesh-manager clean
	$(MAKE) -C proto/bcs-mesh-manager

# format proto
proto-fmt:
	clang-format -i proto/bcs-mesh-manager/bcs-mesh-manager.proto

# build
build:fmt
	go build  -ldflags $(LDFLAG) -o bin/bcs-mesh-manager ./cmd/mesh-manager/main.go

# test
test:
	go test -v ./cmd/... ./internal/... -cover -count=1 -vet=off

# tidy
tidy:
	go mod tidy

# lint
lint:
	golangci-lint run --fix ./...

# vendor
vendor:
	go mod vendor

# use buildx build arm64/amd64 image and push
.PHONY: push-image
push-image: tidy fmt vendor
	docker buildx build --platform linux/amd64 --build-arg VERSION=$(VERSION) --build-arg LDFLAG=$(LDFLAG) -f ./Dockerfile -t $(REPO)/bcs-mesh-manager:$(TAG) --push .
