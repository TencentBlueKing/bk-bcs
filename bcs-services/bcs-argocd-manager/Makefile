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

.PHONY: build-proxy
build-proxy:
	GOOS=linux GOARCH=amd64 go build ${LDFLAG} -o ../${INNER_PACKAGEPATH}/bcs-services/bcs-argocd-manager/bcs-argocd-proxy/bcs-argocd-proxy ./bcs-argocd-proxy/main.go

.PHONY: docker-proxy
docker-proxy: build-proxy
	cp -R ./bcs-argocd-proxy/conf/* ../${INNER_PACKAGEPATH}/bcs-services/bcs-argocd-manager/bcs-argocd-proxy/
	cd ../${INNER_PACKAGEPATH}/bcs-services/bcs-argocd-manager/bcs-argocd-proxy/ && docker build . -t bcs-argocd-proxy:${VERSION}


.PHONY: build-proxy-example-agent
build-proxy-example-agent:
	GOOS=linux GOARCH=amd64 go build ${LDFLAG} -o ../${INNER_PACKAGEPATH}/bcs-services/bcs-argocd-manager/bcs-argocd-proxy/bcs-argocd-example-agent ./bcs-argocd-proxy/example/example.go

.PHONY: build-all
build-all:build-server build-proxy

.PHONY: docker-server
docker-server: build-server
	cp ./bcs-argocd-server/conf/* ../${INNER_PACKAGEPATH}/bcs-services/bcs-argocd-manager/bcs-argocd-server/
	cd ../${INNER_PACKAGEPATH}/bcs-services/bcs-argocd-manager/bcs-argocd-server/ && docker build . -t bcs-argocd-server:${VERSION}

.PHONY: build-controller
build-controller:
	mkdir -p ../${INNER_PACKAGEPATH}/bcs-services/bcs-argocd-manager/bcs-argocd-controller
	GOOS=linux GOARCH=amd64 go build ${LDFLAG} -o ../${INNER_PACKAGEPATH}/bcs-services/bcs-argocd-manager/bcs-argocd-controller/bcs-argocd-controller ./bcs-argocd-controller/main.go

.PHONY: docker-controller
docker-controller: build-controller
	cp -R ./bcs-argocd-controller/conf/* ../${INNER_PACKAGEPATH}/bcs-services/bcs-argocd-manager/bcs-argocd-controller/
	cd ../${INNER_PACKAGEPATH}/bcs-services/bcs-argocd-manager/bcs-argocd-controller/ && docker build . -t bcs-argocd-controller:${VERSION}

.PHONY: build-repo-sidecar-client
build-repo-sidecar-client:
	GOOS=linux GOARCH=amd64 go build ${LDFLAG} -o ../${INNER_PACKAGEPATH}/bcs-services/bcs-argocd-manager/plugins/client/bcs-repo-sidecar-client ./plugins/repo-sidecar/client/main.go

.PHONY: build-repo-sidecar-server
build-repo-sidecar-server:
	GOOS=linux GOARCH=amd64 go build ${LDFLAG} -o ../${INNER_PACKAGEPATH}/bcs-services/bcs-argocd-manager/plugins/server/bcs-repo-sidecar-server ./plugins/repo-sidecar/server/main.go

.PHONY: build-repo-sidecar
build-repo-sidecar: build-repo-sidecar-client build-repo-sidecar-server

.PHONY: docker-repo-sidecar-client
docker-repo-sidecar-client: build-repo-sidecar-client
	cp -R ./plugins/repo-sidecar/client/conf/* ../${INNER_PACKAGEPATH}/bcs-services/bcs-argocd-manager/plugins/client/
	cd ../${INNER_PACKAGEPATH}/bcs-services/bcs-argocd-manager/plugins/client && docker build . -t bcs-repo-sidecar-client:${VERSION}

.PHONY: docker-repo-sidecar-server
docker-repo-sidecar-server: build-repo-sidecar-server
	cp -R ./plugins/repo-sidecar/server/conf/* ../${INNER_PACKAGEPATH}/bcs-services/bcs-argocd-manager/plugins/server/
	cd ../${INNER_PACKAGEPATH}/bcs-services/bcs-argocd-manager/plugins/server && docker build . -t bcs-repo-sidecar-server:${VERSION}

.PHONY: docker-repo-sidecar
docker-repo-sidecar: docker-repo-sidecar-client docker-repo-sidecar-server

.PHONY: build-plugin-example
build-plugin-example:
	GOOS=linux GOARCH=amd64 go build ${LDFLAG} -o ../${INNER_PACKAGEPATH}/bcs-services/bcs-argocd-manager/plugins/example/bcs-argocd-example-plugin ./plugins/example/main.go

.PHONY: docker-plugin-example
docker-plugin-example: build-plugin-example
	cp -R ./plugins/example/conf/* ../${INNER_PACKAGEPATH}/bcs-services/bcs-argocd-manager/plugins/example/
	cd ../${INNER_PACKAGEPATH}/bcs-services/bcs-argocd-manager/plugins/example && docker build . -t bcs-argocd-example-plugin:${VERSION}
