GOPATH:=$(shell go env GOPATH)

bcs_edition?=inner_edition
imagetag?=latest

# init the build information
ifdef HASTAG
	GITTAG=${HASTAG}
else
	GITTAG=$(shell git describe --always)
endif

ifndef BUILDTIME
	BUILDTIME = $(shell date +%Y-%m-%dT%T%z)
  	GITHASH=$(shell git rev-parse HEAD)
	VERSION?=${GITTAG}-$(shell date +%y.%m.%d)
endif

ifndef LDFLAG
LDFLAG=-ldflags "-X github.com/Tencent/bk-bcs/bcs-common/common/static.ZookeeperClientUser=${bcs_zk_client_user} \
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
endif

# build path config
OUTPUT=build/bcs.${VERSION}
ifdef SCENARIOSPACKAGE
	OUTPUT=${SCENARIOSPACKAGE}
endif

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: gen
gen:
	./hack/update-codegen.sh
	controller-gen  rbac:roleName=manager-role crd:allowDangerousTypes=true webhook paths=./pkg/apis/... output:crd:artifacts:config=config/crd/bases output:stdout

.PHONY: cli
cli:
	go mod tidy
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bcs-repack-descheduler-cli ./cli/main.go

.PHONY: build
build: tidy
	mkdir -p ${OUTPUT}/bcs-repack-descheduler
	cp -r images/repack-descheduler/* ${OUTPUT}/bcs-repack-descheduler/
	go build ${LDFLAG} -o ${OUTPUT}/bcs-repack-descheduler/bcs-repack-descheduler main.go
