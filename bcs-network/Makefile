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

cloud-netagent:
	mkdir -p ../${INNER_PACKAGEPATH}/bcs-network/bcs-cloud-netagent
	cp -R ../install/conf/bcs-network/bcs-cloud-netagent/* ../${INNER_PACKAGEPATH}/bcs-network/bcs-cloud-netagent
	GOOS=linux go build ${LDFLAG} -o ../${INNER_PACKAGEPATH}/bcs-network/bcs-cloud-netagent/bcs-cloud-netagent ./bcs-cloud-netagent/cmd/netagent/main.go
	GOOS=linux go build ${LDFLAG} -o ../${INNER_PACKAGEPATH}/bcs-network/bcs-cloud-netagent/bcs-eni-cni ./bcs-cloud-netagent/cmd/bcs-eni-cni/main.go
	GOOS=linux go build ${LDFLAG} -o ../${INNER_PACKAGEPATH}/bcs-network/bcs-cloud-netagent/bcs-eni-ipam ./bcs-cloud-netagent/cmd/bcs-eni-ipam/main.go

cloud-netcontroller:
	mkdir -p ../${INNER_PACKAGEPATH}/bcs-network/bcs-cloud-netcontroller
	cp -R ../install/conf/bcs-network/bcs-cloud-netcontroller/* ../${INNER_PACKAGEPATH}/bcs-network/bcs-cloud-netcontroller
	GOOS=linux go build ${LDFLAG} -o ../${INNER_PACKAGEPATH}/bcs-network/bcs-cloud-netcontroller/bcs-cloud-netcontroller ./bcs-cloud-netcontroller/main.go

cloud-netservice:
	mkdir -p ../${INNER_PACKAGEPATH}/bcs-network/bcs-cloud-netservice
	cp -R ../install/conf/bcs-network/bcs-cloud-netservice/* ../${INNER_PACKAGEPATH}/bcs-network/bcs-cloud-netservice
	cp ./api/protocol/cloudnetservice/cloudnetservice.swagger.json ../${INNER_PACKAGEPATH}/bcs-network/bcs-cloud-netservice
	GOOS=linux go build ${LDFLAG} -o ../${INNER_PACKAGEPATH}/bcs-network/bcs-cloud-netservice/bcs-cloud-netservice ./bcs-cloud-netservice/cmd/cloudnetservice/main.go

# qcloud-eip
network:
	mkdir -p ../${INNER_PACKAGEPATH}/bcs-mesos-node/bcs-cni/bin
	mkdir -p ../${INNER_PACKAGEPATH}/bcs-mesos-node/bcs-cni/conf
	cp -R ../install/conf/bcs-mesos-node/qcloud-eip/* ../${INNER_PACKAGEPATH}/bcs-mesos-node/bcs-cni/conf
	GOOS=linux go build ${LDFLAG} -o ../${INNER_PACKAGEPATH}/bcs-mesos-node/bcs-cni/bin/qcloud-eip ./qcloud-eip/main.go

# old version network agent
bcs-cloud-network-agent:
	mkdir -p ../${INNER_PACKAGEPATH}/bcs-network/bcs-cloud-network-agent
	cp -R ../install/conf/bcs-network/bcs-cloud-network-agent/* ../${INNER_PACKAGEPATH}/bcs-network/bcs-cloud-network-agent
	GOOS=linux go build ${LDFLAG} -o ../${INNER_PACKAGEPATH}/bcs-network/bcs-cloud-network-agent/bcs-cloud-network-agent ./bcs-cloudnetwork/cloud-network-agent/main.go
	GOOS=linux go build ${LDFLAG} -o ../${INNER_PACKAGEPATH}/bcs-network/bcs-cloud-network-agent/bcs-eni ./bcs-cloud-netagent/cmd/bcs-eni-cni/main.go

networkpolicy:
	mkdir -p ../${INNER_PACKAGEPATH}/bcs-network/bcs-networkpolicy
	cp -R ../install/conf/bcs-network/bcs-networkpolicy ../${INNER_PACKAGEPATH}/bcs-network
	go build ${LDFLAG} -o ../${INNER_PACKAGEPATH}/bcs-network/bcs-networkpolicy/bcs-networkpolicy ./bcs-networkpolicy/main.go

ingress-controller:
	mkdir -p ../${INNER_PACKAGEPATH}/bcs-network/bcs-ingress-controller
	cp -R ../install/conf/bcs-network/bcs-ingress-controller/* ../${INNER_PACKAGEPATH}/bcs-network/bcs-ingress-controller
	GOOS=linux go build ${LDFLAG} -o ../${INNER_PACKAGEPATH}/bcs-network/bcs-ingress-controller/bcs-ingress-controller ./bcs-ingress-controller/main.go