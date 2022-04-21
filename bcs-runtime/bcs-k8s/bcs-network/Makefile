# BlueKing Container System Makefile
# default config

# BlueKing Container System Makefile
# default config
MAKE:=make
bcs_edition?=inner_edition

# init the build information
ifdef HASTAG
	GITTAG=${HASTAG}
else
	GITTAG=$(shell git describe --always)
endif

BUILDTIME = $(shell date +%Y-%m-%dT%T%z)
GITHASH=$(shell git rev-parse HEAD)
VERSION=${GITTAG}-$(shell date +%y.%m.%d)
WORKSPACE=$(shell pwd)
INSTALL_CONF="../../../install/conf"
BCS_CONF_NETWORK_PATH="../../../install/conf/bcs-runtime/bcs-k8s/bcs-network"

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
INNER_PACKAGEPATH="../../../build/bcs.${VERSION}/bcs-runtime/bcs-k8s"

default: cloud-netagent cloud-netcontroller cloud-netservice network bcs-cloud-network-agent networkpolicy ingress-controller ipmasq-cidrsync ipres-webhook

cloud-netagent:
	mkdir -p ${INNER_PACKAGEPATH}/bcs-network/bcs-cloud-netagent
	cp -R ${BCS_CONF_NETWORK_PATH}/bcs-cloud-netagent/* ${INNER_PACKAGEPATH}/bcs-network/bcs-cloud-netagent
	GOOS=linux go build ${LDFLAG} -o ${INNER_PACKAGEPATH}/bcs-network/bcs-cloud-netagent/bcs-cloud-netagent ./bcs-cloud-netagent/cmd/netagent/main.go
	GOOS=linux go build ${LDFLAG} -o ${INNER_PACKAGEPATH}/bcs-network/bcs-cloud-netagent/bcs-eni-cni ./bcs-cloud-netagent/cmd/bcs-eni-cni/main.go
	GOOS=linux go build ${LDFLAG} -o ${INNER_PACKAGEPATH}/bcs-network/bcs-cloud-netagent/bcs-eni-ipam ./bcs-cloud-netagent/cmd/bcs-eni-ipam/main.go

cloud-netcontroller:
	mkdir -p ${INNER_PACKAGEPATH}/bcs-network/bcs-cloud-netcontroller
	cp -R ${BCS_CONF_NETWORK_PATH}/bcs-cloud-netcontroller/* ${INNER_PACKAGEPATH}/bcs-network/bcs-cloud-netcontroller
	GOOS=linux go build ${LDFLAG} -o ${INNER_PACKAGEPATH}/bcs-network/bcs-cloud-netcontroller/bcs-cloud-netcontroller ./bcs-cloud-netcontroller/main.go

cloud-netservice:
	mkdir -p ${INNER_PACKAGEPATH}/bcs-network/bcs-cloud-netservice
	cp -R ${BCS_CONF_NETWORK_PATH}/bcs-cloud-netservice/* ${INNER_PACKAGEPATH}/bcs-network/bcs-cloud-netservice
	cp ./api/protocol/cloudnetservice/cloudnetservice.swagger.json ${INNER_PACKAGEPATH}/bcs-network/bcs-cloud-netservice
	GOOS=linux go build ${LDFLAG} -o ${INNER_PACKAGEPATH}/bcs-network/bcs-cloud-netservice/bcs-cloud-netservice ./bcs-cloud-netservice/cmd/cloudnetservice/main.go

# qcloud-eip
network:
	mkdir -p ../../../build/bcs.${VERSION}/bcs-runtime/bcs-mesos/bcs-mesos-node/bcs-cni/bin
	mkdir -p ../../../build/bcs.${VERSION}/bcs-runtime/bcs-mesos/bcs-mesos-node/bcs-cni/conf
	cp -R ${INSTALL_CONF}/bcs-runtime/bcs-mesos/bcs-mesos-node/qcloud-eip/* ../../../build/bcs.${VERSION}/bcs-runtime/bcs-mesos/bcs-mesos-node/bcs-cni/conf
	GOOS=linux go build ${LDFLAG} -o ../../../build/bcs.${VERSION}/bcs-runtime/bcs-mesos/bcs-mesos-node/bcs-cni/bin/qcloud-eip ./qcloud-eip/main.go

# old version network agent
bcs-cloud-network-agent:
	mkdir -p ${INNER_PACKAGEPATH}/bcs-network/bcs-cloud-network-agent
	cp -R ${BCS_CONF_NETWORK_PATH}/bcs-cloud-network-agent/* ${INNER_PACKAGEPATH}/bcs-network/bcs-cloud-network-agent
	GOOS=linux go build ${LDFLAG} -o ${INNER_PACKAGEPATH}/bcs-network/bcs-cloud-network-agent/bcs-cloud-network-agent ./bcs-cloudnetwork/cloud-network-agent/main.go
	GOOS=linux go build ${LDFLAG} -o ${INNER_PACKAGEPATH}/bcs-network/bcs-cloud-network-agent/bcs-eni ./bcs-cloudnetwork/cloud-network-agent/bcs-eni-cni/main.go

# networkpolicy network policy module
networkpolicy:
	mkdir -p ${INNER_PACKAGEPATH}/bcs-network/bcs-networkpolicy
	cp -R ${BCS_CONF_NETWORK_PATH}/bcs-networkpolicy ${INNER_PACKAGEPATH}/bcs-network
	go build ${LDFLAG} -o ${INNER_PACKAGEPATH}/bcs-network/bcs-networkpolicy/bcs-networkpolicy ./bcs-networkpolicy/main.go

# ingress-controller bcs ingress controller
ingress-controller:
	mkdir -p ${INNER_PACKAGEPATH}/bcs-network/bcs-ingress-controller
	cp -R ${BCS_CONF_NETWORK_PATH}/bcs-ingress-controller/* ${INNER_PACKAGEPATH}/bcs-network/bcs-ingress-controller
	GOOS=linux go build ${LDFLAG} -o ${INNER_PACKAGEPATH}/bcs-network/bcs-ingress-controller/bcs-ingress-controller ./bcs-ingress-controller/main.go

# ipmasq-cidrsync bcs ipmasq cidrsync
ipmasq-cidrsync:
	mkdir -p ${INNER_PACKAGEPATH}/bcs-network/bcs-ipmasq-cidrsync
	cp -R ${BCS_CONF_NETWORK_PATH}/bcs-ipmasq-cidrsync/* ${INNER_PACKAGEPATH}/bcs-network/bcs-ipmasq-cidrsync
	GOOS=linux go build ${LDFLAG} -o ${INNER_PACKAGEPATH}/bcs-network/bcs-ipmasq-cidrsync/bcs-ipmasq-cidrsync ./bcs-ipmasq-cidrsync/main.go

# ipres-webhook bcs ip resource webhook
ipres-webhook:
	mkdir -p ${INNER_PACKAGEPATH}/bcs-network/bcs-ipres-webhook
	cp -R ${BCS_CONF_NETWORK_PATH}/bcs-ipres-webhook/* ${INNER_PACKAGEPATH}/bcs-network/bcs-ipres-webhook
	GOOS=linux go build ${LDFLAG} -o ${INNER_PACKAGEPATH}/bcs-network/bcs-ipres-webhook/bcs-ipres-webhook ./bcs-ipres-webhook/main.go
	GOOS=linux go build ${LDFLAG} -o ${INNER_PACKAGEPATH}/bcs-network/bcs-ipres-webhook/bcs-ipres-nslabel-injector ./bcs-ipres-webhook/nslabelinjector/main.go

# unit test for bcs ingress controller
test-ingress-controller:
	@go test -v -coverprofile=./bcs-ingress-controller/.coverage.out ./bcs-ingress-controller/...
	@go tool cover -func=./bcs-ingress-controller/.coverage.out | tail -n1 | awk '{print "Total test coverage: " $$3}'
	@go tool cover -html=./bcs-ingress-controller/.coverage.out -o ./bcs-ingress-controller/coverage.html