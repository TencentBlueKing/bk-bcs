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

LDFLAG=-ldflags "-X github.com/Tencent/bk-bcs/bcs-common/common/static.EncryptionKey=${bcs_encryption_key} \
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

.PHONY: proto
proto:
	./hack/generate-proto.sh

.PHONY: client
client:
	./hack/update-codegen.sh

.PHONY: tools
tools:
	./hack/install-tools.sh

.PHONY: crds
crds:
	./hack/update-crds.sh

.PHONY: build-server
#build-server: proto client
build-server:
	mkdir -p ../${INNER_PACKAGEPATH}/bcs-services/bcs-argocd-manager/bcs-argocd-server/swagger
	cp -R ./third_party/swagger-ui/* ../${INNER_PACKAGEPATH}/bcs-services/bcs-argocd-manager/bcs-argocd-server/swagger/
	cp ./pkg/sdk/project/project.swagger.json ../${INNER_PACKAGEPATH}/bcs-services/bcs-argocd-manager/bcs-argocd-server/swagger/project.swagger.json
	cp ./pkg/sdk/instance/instance.swagger.json ../${INNER_PACKAGEPATH}/bcs-services/bcs-argocd-manager/bcs-argocd-server/swagger/instance.swagger.json
	cp ./pkg/sdk/plugin/plugin.swagger.json ../${INNER_PACKAGEPATH}/bcs-services/bcs-argocd-manager/bcs-argocd-server/swagger/plugin.swagger.json
	GOOS=linux GOARCH=amd64 go build ${LDFLAG} -o ../${INNER_PACKAGEPATH}/bcs-services/bcs-argocd-manager/bcs-argocd-server/bcs-argocd-server ./bcs-argocd-server/main.go

.PHONY: docker-server
docker-server: build-server
	cp ./bcs-argocd-server/conf/* ../${INNER_PACKAGEPATH}/bcs-services/bcs-argocd-manager/bcs-argocd-server/
	cd ../${INNER_PACKAGEPATH}/bcs-services/bcs-argocd-manager/bcs-argocd-server/ && docker build . -t bcs-argocd-server:${VERSION}

build-controller:
	mkdir -p ../${INNER_PACKAGEPATH}/bcs-services/bcs-argocd-manager/bcs-argocd-controller
	GOOS=linux GOARCH=amd64 go build ${LDFLAG} -o ../${INNER_PACKAGEPATH}/bcs-services/bcs-argocd-manager/bcs-argocd-controller/bcs-argocd-controller ./bcs-argocd-controller/main.go

.PHONY: docker-controller
docker-controller: build-controller
	cp -R ./bcs-argocd-controller/conf/* ../${INNER_PACKAGEPATH}/bcs-services/bcs-argocd-manager/bcs-argocd-controller/
	cd ../${INNER_PACKAGEPATH}/bcs-services/bcs-argocd-manager/bcs-argocd-controller/ && docker build . -t bcs-argocd-controller:${VERSION}