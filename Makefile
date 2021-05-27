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
PACKAGEPATH=./build/bcs.${VERSION}
EXPORTPATH=./build/api_export

# options
default:bcs-service bcs-network bcs-mesos bcs-k8s

bcs-k8s:k8s-watch kube-agent k8s-driver gamestatefulset gamedeployment hook-operator \
	cc-agent csi-cbs kube-sche federated-apiserver federated-apiserver-kubectl-agg

bcs-mesos:executor mesos-driver mesos-watch scheduler loadbalance netservice hpacontroller \
	consoleproxy process-executor process-daemon bmsf-mesos-adapter detection

bcs-service:api client bkcmdb-synchronizer clb-controller cpuset gateway gw-controller log-manager \
	mesh-manager logbeat-sidecar netservice sd-prometheus storage \
	user-manager webhook-server cluster-manager tools alert-manager

bcs-network:network networkpolicy ingress-controller cloud-netservice cloud-netcontroller cloud-netagent 

allpack: svcpack k8spack mmpack mnpack netpack
	cd build && tar -czf bcs.${VERSION}.tgz bcs.${VERSION}

# tag for different edition compiling
inner:
	$(MAKE) default bcs_edition=inner_edition
ce:
	$(MAKE) default bcs_edition=community_edition
ee:
	$(MAKE) default bcs_edition=enterprise_edition

clean:
	rm -rf ./build

svcpack:
	cd ./build/bcs.${VERSION}/bcs-services && find . -type f ! -name MD5 | xargs -L1 md5sum > MD5
	
k8spack:
	cd ./build/bcs.${VERSION}/bcs-k8s-master && find . -type f ! -name MD5 | xargs -L1 md5sum > MD5 

mmpack:
	cd ./build/bcs.${VERSION}/bcs-mesos-master && find . -type f ! -name MD5 | xargs -L1 md5sum > MD5

mnpack:
	cd ./build/bcs.${VERSION}/bcs-mesos-node && find . -type f ! -name MD5 | xargs -L1 md5sum > MD5

netpack:
	cd ./build/bcs.${VERSION}/bcs-network && find . -type f ! -name MD5 | xargs -L1 md5sum > MD5

pre:
	@echo "git tag: ${GITTAG}"
	mkdir -p ${PACKAGEPATH}
	mkdir -p ${EXPORTPATH}
	go fmt ./...
	cd ./scripts && chmod +x vet.sh && ./vet.sh

api:pre
	mkdir -p ${PACKAGEPATH}/bcs-services
	cp -R ./install/conf/bcs-services/bcs-api ${PACKAGEPATH}/bcs-services
	cd bcs-services/bcs-api && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-api/bcs-api ./main.go

gateway:pre
	mkdir -p ${PACKAGEPATH}/bcs-services
	cp -R ./install/conf/bcs-services/bcs-gateway-discovery ${PACKAGEPATH}/bcs-services
	cp -R ./bcs-services/bcs-gateway-discovery/plugins/apisix ${PACKAGEPATH}/bcs-services/bcs-gateway-discovery/
	cd bcs-services/bcs-gateway-discovery && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-gateway-discovery/bcs-gateway-discovery ./main.go

gateway-container: gateway
	cd ${PACKAGEPATH}/bcs-services/bcs-gateway-discovery/ && docker build -t bcs/apisix:${GITTAG} -f Dockerfile.apisix .
	cd ${PACKAGEPATH}/bcs-services/bcs-gateway-discovery/ && docker build -t bcs/bcs-gateway-discovery:${GITTAG} -f Dockerfile.gateway .

kube-agent:pre
	mkdir -p ${PACKAGEPATH}/bcs-k8s-master
	cp -R ./install/conf/bcs-k8s-master/bcs-kube-agent ${PACKAGEPATH}/bcs-k8s-master
	cd ./bcs-k8s/bcs-kube-agent && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-k8s-master/bcs-kube-agent/bcs-kube-agent ./main.go
	

client:pre
	mkdir -p ${PACKAGEPATH}/bcs-services
	cp -R ./install/conf/bcs-services/bcs-client ${PACKAGEPATH}/bcs-services
	cd ./bcs-services/bcs-client && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-client/bcs-client ./cmd/main.go

dns:
	mkdir -p ${PACKAGEPATH}/bcs-services
	mkdir -p ${PACKAGEPATH}/bcs-mesos-master
	cp -R ./install/conf/bcs-mesos-master/bcs-dns ${PACKAGEPATH}/bcs-mesos-master
	cp -R ./install/conf/bcs-services/bcs-dns-service ${PACKAGEPATH}/bcs-services
	cd ../coredns && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-dns-service/bcs-dns-service coredns.go
	cd ../coredns && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-mesos-master/bcs-dns/bcs-dns coredns.go

storage:pre
	mkdir -p ${PACKAGEPATH}/bcs-services
	cp -R ./install/conf/bcs-services/bcs-storage ${PACKAGEPATH}/bcs-services
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-services/bcs-storage/bcs-storage ./bcs-services/bcs-storage/storage.go

loadbalance:pre
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-services/bcs-loadbalance/bcs-loadbalance ./bcs-services/bcs-loadbalance/main.go
	cp -r ./bcs-services/bcs-loadbalance/image/* ${PACKAGEPATH}/bcs-services/bcs-loadbalance/

executor:pre
	mkdir -p ${PACKAGEPATH}/bcs-mesos-node
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-mesos-node/bcs-container-executor/bcs-container-executor ./bcs-mesos/bcs-container-executor/main.go

process-executor:pre
	mkdir -p ${PACKAGEPATH}/bcs-mesos-node
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-mesos-node/bcs-process-executor/bcs-process-executor ./bcs-mesos/bcs-process-executor/main.go

process-daemon:pre
	mkdir -p ${PACKAGEPATH}/bcs-mesos-node
	cp -R ./install/conf/bcs-mesos-node/bcs-process-daemon ${PACKAGEPATH}/bcs-mesos-node
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-mesos-node/bcs-process-daemon/bcs-process-daemon ./bcs-mesos/bcs-process-daemon/main.go

netservice:pre
	mkdir -p ${PACKAGEPATH}/bcs-services
	cp -R ./install/conf/bcs-services/bcs-netservice ${PACKAGEPATH}/bcs-services
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-services/bcs-netservice/bcs-netservice ./bcs-services/bcs-netservice/main.go
	
	mkdir -p ${PACKAGEPATH}/bcs-mesos-master
	cp -R ./install/conf/bcs-mesos-master/bcs-netservice ${PACKAGEPATH}/bcs-mesos-master
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-mesos-master/bcs-netservice/bcs-netservice ./bcs-services/bcs-netservice/main.go

	mkdir -p ${PACKAGEPATH}/bcs-mesos-node/bcs-cni/bin/conf
	mkdir -p ${PACKAGEPATH}/bcs-mesos-node/bcs-cni/conf
	cp -R ./install/conf/bcs-mesos-node/bcs-ipam/bcs.conf.template ${PACKAGEPATH}/bcs-mesos-node/bcs-cni/bin/conf
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-mesos-node/bcs-cni/bin/bcs-ipam ./bcs-services/bcs-netservice/bcs-ipam/main.go

mesos-driver:pre
	mkdir -p ${PACKAGEPATH}/bcs-mesos-master
	cp -R ./install/conf/bcs-mesos-master/bcs-mesos-driver ${PACKAGEPATH}/bcs-mesos-master
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-mesos-master/bcs-mesos-driver/bcs-mesos-driver ./bcs-mesos/bcs-mesos-driver/main.go

mesos-watch:pre
	mkdir -p ${PACKAGEPATH}/bcs-mesos-master
	cp -R ./install/conf/bcs-mesos-master/bcs-mesos-watch ${PACKAGEPATH}/bcs-mesos-master
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-mesos-master/bcs-mesos-watch/bcs-mesos-watch ./bcs-mesos/bcs-mesos-watch/main.go

kube-sche:pre
	mkdir -p ${PACKAGEPATH}/bcs-k8s-master
	cp -R ./install/conf/bcs-k8s-master/bcs-k8s-custom-scheduler ${PACKAGEPATH}/bcs-k8s-master
	cd ./bcs-k8s/bcs-k8s-custom-scheduler && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-k8s-master/bcs-k8s-custom-scheduler/bcs-k8s-custom-scheduler ./main.go

csi-cbs:pre
	mkdir -p ${PACKAGEPATH}/bcs-k8s-master
	cp -R ./install/conf/bcs-k8s-master/bcs-k8s-csi-tencentcloud ${PACKAGEPATH}/bcs-k8s-master
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-k8s-master/bcs-k8s-csi-tencentcloud/bcs-k8s-csi-tencentcloud ./bcs-k8s/bcs-k8s-csi-tencentcloud/cmd/cbs/main.go

scheduler:pre
	mkdir -p ${PACKAGEPATH}/bcs-mesos-master
	cp -R ./install/conf/bcs-mesos-master/bcs-scheduler ${PACKAGEPATH}/bcs-mesos-master
	cd ./bcs-mesos/bcs-scheduler && go build ${LDFLAG} -o ../../${PACKAGEPATH}/bcs-mesos-master/bcs-scheduler/bcs-scheduler ./main.go && cd -
	cd ./bcs-mesos/bcs-scheduler && go build -buildmode=plugin -o ../../${PACKAGEPATH}/bcs-mesos-master/bcs-scheduler/plugin/bin/ip-resources/ip-resources.so ./src/plugin/bin/ip-resources/ipResource.go && cd -
	cd ./bcs-mesos/bcs-scheduler && go build ${LDFLAG} -o ../../${PACKAGEPATH}/bcs-mesos-master/bcs-scheduler/bcs-migrate-data ./bcs-migrate-data/main.go && cd -

logbeat-sidecar:pre
	mkdir -p ${PACKAGEPATH}/bcs-services
	cp -R ./install/conf/bcs-services/bcs-logbeat-sidecar ${PACKAGEPATH}/bcs-services
	cd ./bcs-services/bcs-logbeat-sidecar && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-logbeat-sidecar/bcs-logbeat-sidecar ./main.go

log-manager:pre
	mkdir -p ${PACKAGEPATH}/bcs-services
	cp -R ./install/conf/bcs-services/bcs-log-manager ${PACKAGEPATH}/bcs-services
	cd ./bcs-services/bcs-log-manager && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-log-manager/bcs-log-manager ./main.go

mesh-manager:pre
	mkdir -p ${PACKAGEPATH}/bcs-services
	cp -R ./install/conf/bcs-services/bcs-mesh-manager ${PACKAGEPATH}/bcs-services
	cd bcs-services/bcs-mesh-manager && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-mesh-manager/bcs-mesh-manager ./main.go

hpacontroller:pre
	mkdir -p ${PACKAGEPATH}/bcs-mesos-master
	cp -R ./install/conf/bcs-mesos-master/bcs-hpacontroller ${PACKAGEPATH}/bcs-mesos-master
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-mesos-master/bcs-hpacontroller/bcs-hpacontroller ./bcs-mesos/bcs-hpacontroller

sd-prometheus:pre
	mkdir -p ${PACKAGEPATH}/bcs-services
	mkdir -p ${PACKAGEPATH}/bcs-mesos-master
	cp -R ./install/conf/bcs-services/bcs-service-prometheus-service ${PACKAGEPATH}/bcs-services
	cp -R ./install/conf/bcs-mesos-master/bcs-service-prometheus ${PACKAGEPATH}/bcs-mesos-master
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-services/bcs-service-prometheus-service/bcs-service-prometheus-service ./bcs-services/bcs-service-prometheus/main.go
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-mesos-master/bcs-service-prometheus/bcs-service-prometheus ./bcs-services/bcs-service-prometheus/main.go

k8s-driver:pre
	cd ./bcs-k8s/bcs-k8s-driver && make k8s-driver

k8s-watch:pre
	mkdir -p ${PACKAGEPATH}/bcs-k8s-master/bcs-k8s-watch
	cp -R ./install/conf/bcs-k8s-master/bcs-k8s-watch/* ${PACKAGEPATH}/bcs-k8s-master/bcs-k8s-watch
	cd bcs-k8s/bcs-k8s-watch && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-k8s-master/bcs-k8s-watch/bcs-k8s-watch ./main.go

gamestatefulset:pre
	mkdir -p ${PACKAGEPATH}/bcs-k8s-master
	cp -R ./install/conf/bcs-k8s-master/bcs-gamestatefulset-operator ${PACKAGEPATH}/bcs-k8s-master
	cd bcs-k8s/bcs-gamestatefulset-operator && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-k8s-master/bcs-gamestatefulset-operator/bcs-gamestatefulset-operator ./cmd/gamestatefulset-operator/main.go

gamedeployment:pre
	mkdir -p ${PACKAGEPATH}/bcs-k8s-master
	cp -R ./install/conf/bcs-k8s-master/bcs-gamedeployment-operator ${PACKAGEPATH}/bcs-k8s-master
	cd bcs-k8s/bcs-gamedeployment-operator && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-k8s-master/bcs-gamedeployment-operator/bcs-gamedeployment-operator ./cmd/gamedeployment-operator/main.go

hook-operator:pre
	mkdir -p ${PACKAGEPATH}/bcs-k8s-master
	cp -R ./install/conf/bcs-k8s-master/bcs-hook-operator ${PACKAGEPATH}/bcs-k8s-master
	cd bcs-k8s/bcs-hook-operator && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-k8s-master/bcs-hook-operator/bcs-hook-operator ./cmd/hook-operator/main.go

federated-apiserver:pre
	mkdir -p ${PACKAGEPATH}/bcs-k8s-master
	cp -R ./install/conf/bcs-k8s-master/bcs-federated-apiserver ${PACKAGEPATH}/bcs-k8s-master
	cd bcs-k8s/bcs-federated-apiserver && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-k8s-master/bcs-federated-apiserver/bcs-federated-apiserver ./cmd/apiserver/main.go

federated-apiserver-kubectl-agg:pre
	mkdir -p ${PACKAGEPATH}/bcs-k8s-master
	cd bcs-k8s/bcs-federated-apiserver && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-k8s-master/bcs-federated-apiserver/kubectl-agg ./cmd/kubectl-agg/main.go

egress-controller:pre
	mkdir -p ${PACKAGEPATH}/bcs-k8s-master
	cp -R ./install/conf/bcs-k8s-master/bcs-egress-controller ${PACKAGEPATH}/bcs-k8s-master
	#copy nginx template for egress controller
	cp -R bcs-k8s/bcs-egress/deploy/config ${PACKAGEPATH}/bcs-k8s-master/bcs-egress-controller
	cd bcs-k8s/bcs-egress && go build -o ${WORKSPACE}/${PACKAGEPATH}/bcs-k8s-master/bcs-egress-controller/bcs-egress-controller ./cmd/bcs-egress-controller/main.go

consoleproxy:pre
	mkdir -p ${PACKAGEPATH}/bcs-mesos-node/bcs-consoleproxy
	cp -R ./install/conf/bcs-mesos-node/bcs-consoleproxy ${PACKAGEPATH}/bcs-mesos-node
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-mesos-node/bcs-consoleproxy/bcs-consoleproxy ./bcs-mesos/bcs-consoleproxy/main.go

bmsf-mesos-adapter:pre
	mkdir -p ${PACKAGEPATH}/bcs-mesos-master
	cp -R ./install/conf/bcs-mesos-master/bmsf-mesos-adapter ${PACKAGEPATH}/bcs-mesos-master
	cd ./bmsf-mesh && go build ${LDFLAG} -o ../${PACKAGEPATH}/bcs-mesos-master/bmsf-mesos-adapter/bmsf-mesos-adapter ./bmsf-mesos-adapter/main.go

cpuset:pre
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-cpuset-device
	cp -R ./install/conf/bcs-mesos-node/bcs-cpuset-device ${PACKAGEPATH}/bcs-services
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-services/bcs-cpuset-device/bcs-cpuset-device ./bcs-services/bcs-cpuset-device/main.go

gw-controller:pre
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-gw-controller
	cp -R ./install/conf/bcs-services/bcs-gw-controller ${PACKAGEPATH}/bcs-services
	cd bcs-services/bcs-clb-controller && go build ${LDFLAG} -o ../../${PACKAGEPATH}/bcs-services/bcs-gw-controller/bcs-gw-controller ./bcs-gw-controller/main.go

webhook-server:pre
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-webhook-server
	cp -R ./install/conf/bcs-services/bcs-webhook-server/* ${PACKAGEPATH}/bcs-services/bcs-webhook-server
	cd ./bcs-services/bcs-webhook-server && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-webhook-server/bcs-webhook-server ./cmd/server.go

detection:pre
	cp -R ./install/conf/bcs-services/bcs-network-detection ${PACKAGEPATH}/bcs-services
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-services/bcs-network-detection/bcs-network-detection ./bcs-services/bcs-network-detection/main.go

tools:
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-services/cryptools ./install/cryptool/main.go
	
user-manager:pre
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-user-manager
	cp -R ./install/conf/bcs-services/bcs-user-manager ${PACKAGEPATH}/bcs-services
	cd bcs-services/bcs-user-manager/ && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-user-manager/bcs-user-manager ./main.go

cc-agent:pre
	mkdir -p ${PACKAGEPATH}/bcs-k8s-master
	cp -R ./install/conf/bcs-k8s-master/bcs-cc-agent ${PACKAGEPATH}/bcs-k8s-master
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-k8s-master/bcs-cc-agent/bcs-cc-agent ./bcs-k8s/bcs-cc-agent/main.go

# network plugins section
networkpolicy:pre
	cd ./bcs-network && make networkpolicy

cloud-network-agent:pre
	cd ./bcs-network && make bcs-cloud-network-agent

bkcmdb-synchronizer:pre
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-bkcmdb-synchronizer
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-services/bcs-bkcmdb-synchronizer/bcs-bkcmdb-synchronizer ./bcs-services/bcs-bkcmdb-synchronizer/main.go

cloud-netservice:pre
	cd ./bcs-network && make cloud-netservice

cloud-netcontroller:pre
	cd ./bcs-network && make cloud-netcontroller

cloud-netagent:pre
	cd ./bcs-network && make cloud-netagent

ingress-controller:pre
	cd ./bcs-network && make ingress-controller

network:pre
	cd ./bcs-network && make network

clb-controller:pre
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-clb-controller
	cp -R ./install/conf/bcs-services/bcs-clb-controller ${PACKAGEPATH}/bcs-services
	cp ./bcs-services/bcs-clb-controller/docker/Dockerfile ${PACKAGEPATH}/bcs-services/bcs-clb-controller/Dockerfile.old
	cd ./bcs-services/bcs-clb-controller && go build ${LDFLAG} -o ../../${PACKAGEPATH}/bcs-services/bcs-clb-controller/bcs-clb-controller ./main.go
	cp ${PACKAGEPATH}/bcs-services/bcs-clb-controller/bcs-clb-controller ${PACKAGEPATH}/bcs-services/bcs-clb-controller/clb-controller
#end of network plugins

# bcs-service section
cluster-manager:pre
	cd ./bcs-services/bcs-cluster-manager && make clustermanager

alert-manager:pre
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-alert-manager/swagger
	cp -R ./install/conf/bcs-services/bcs-alert-manager/*  ${PACKAGEPATH}/bcs-services/bcs-alert-manager
	cp -R ./bcs-services/bcs-alert-manager/pkg/third_party/swagger-ui ${PACKAGEPATH}/bcs-services/bcs-alert-manager/swagger/swagger-ui
	cp ./bcs-services/bcs-alert-manager/pkg/proto/alertmanager/alertmanager.swagger.json ${PACKAGEPATH}/bcs-services/bcs-alert-manager/swagger/alertmanager.swagger.json
	cd ./bcs-services/bcs-alert-manager/ && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-alert-manager/bcs-alert-manager ./main.go

# end of bcs-service section
