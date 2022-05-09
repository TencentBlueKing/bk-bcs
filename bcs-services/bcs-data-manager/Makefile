.PHONY: proto build test docker

bcs_edition?=inner_edition

# init the build information
ifdef HASTAG
	GITTAG=${HASTAG}
else
	GITTAG=$(shell git describe --always)
endif

BUILDTIME = $(shell date +%Y-%m-%dT%T%z)
GITHASH=$(shell git rev-parse HEAD)
VERSION?=${GITTAG}-$(shell date +%y.%m.%d)
WORKSPACE=$(shell pwd)

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

# build path config
INNER_PACKAGEPATH=build/bcs.${VERSION}

build:
	mkdir -p ../${INNER_PACKAGEPATH}/bcs-services/bcs-data-manager/swagger
	cp -R ./third_party/swagger-ui/* ../${INNER_PACKAGEPATH}/bcs-services/bcs-data-manager/swagger/
	cp ./proto/bcs-data-manager/bcs-data-manager.swagger.json ../${INNER_PACKAGEPATH}/bcs-services/bcs-data-manager/swagger/bcs-data-manager.swagger.json
	GOOS=linux GOARCH=amd64 go build ${LDFLAG} -o ../${INNER_PACKAGEPATH}/bcs-services/bcs-data-manager/bcs-data-manager ./main.go

client:
	GOOS=linux GOARCH=amd64 go build ${LDFLAG} -o ../${INNER_PACKAGEPATH}/bcs-services/bcs-data-manager/kubectl-bcs-data-manager pkg/client/main.go

test:
	go test -v ./... -cover

docker: build
	cp ./image/* ../${INNER_PACKAGEPATH}/bcs-services/bcs-data-manager/
	cd ../${INNER_PACKAGEPATH}/bcs-services/bcs-data-manager/ && docker build . -t bcs-data-manager:${VERSION}
