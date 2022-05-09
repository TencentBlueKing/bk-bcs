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

proto:
	$(MAKE) -C proto/bcs-helm-manager clean
	$(MAKE) -C proto/bcs-helm-manager

build: proto
	mkdir -p ../${INNER_PACKAGEPATH}/bcs-services/bcs-helm-manager/swagger
	cp -R ./third_party/swagger-ui/* ../${INNER_PACKAGEPATH}/bcs-services/bcs-helm-manager/swagger/
	cp ./proto/bcs-helm-manager/bcs-helm-manager.swagger.json ../${INNER_PACKAGEPATH}/bcs-services/bcs-helm-manager/swagger/helmmanager.swagger.json
	GOOS=linux GOARCH=amd64 go build ${LDFLAG} -o ../${INNER_PACKAGEPATH}/bcs-services/bcs-helm-manager/bcs-helm-manager ./main.go

client: proto
	GOOS=linux GOARCH=amd64 go build ${LDFLAG} -o ../${INNER_PACKAGEPATH}/bcs-services/bcs-helm-manager/bcs-helm-client internal/client/main.go

test:
	go test -v ./... -cover

docker: build
	cp helm/images/bcs-helm-manager/* ../${INNER_PACKAGEPATH}/bcs-services/bcs-helm-manager/
	cd ../${INNER_PACKAGEPATH}/bcs-services/bcs-helm-manager/ && docker build . -t bcs-helm-manager:${VERSION}
