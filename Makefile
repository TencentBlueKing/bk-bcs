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

BCS_NETWORK_PATH=${WORKSPACE}/bcs-runtime/bcs-k8s/bcs-network
BCS_COMPONENT_PATH=${WORKSPACE}/bcs-runtime/bcs-k8s/bcs-component
BCS_MESOS_PATH=${WORKSPACE}/bcs-runtime/bcs-mesos
BCS_CONF_COMPONENT_PATH=${WORKSPACE}/install/conf/bcs-runtime/bcs-k8s/bcs-component
BCS_CONF_NETWORK_PATH=${WORKSPACE}/install/conf/bcs-runtime/bcs-k8s/bcs-network
BCS_CONF_MESOS_PATH=${WORKSPACE}/install/conf/bcs-runtime/bcs-mesos
BCS_CONF_SERVICES_PATH=${WORKSPACE}/install/conf/bcs-services

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
default:bcs-runtime bcs-scenarios bcs-services #TODO: bcs-resources

bcs-runtime: bcs-k8s bcs-mesos

bcs-k8s: bcs-component bcs-network

bcs-component:k8s-driver gamestatefulset gamedeployment hook-operator \
	cc-agent csi-cbs kube-sche federated-apiserver apiserver-proxy \
	apiserver-proxy-tools logbeat-sidecar webhook-server clusternet-controller

bcs-network:network networkpolicy ingress-controller cloud-netservice cloud-netcontroller cloud-netagent

bcs-mesos:executor mesos-driver mesos-watch scheduler loadbalance netservice hpacontroller \
	consoleproxy process-executor process-daemon bmsf-mesos-adapter detection clb-controller gw-controller

bcs-services:api client bkcmdb-synchronizer cpuset gateway log-manager \
	mesh-manager netservice sd-prometheus storage \
	user-manager cluster-manager tools alert-manager k8s-watch kube-agent

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
	cd ./build/bcs.${VERSION}/bcs-runtime/bcs-k8s/bcs-component && find . -type f ! -name MD5 | xargs -L1 md5sum > MD5 

mmpack:
	cd ./build/bcs.${VERSION}/bcs-runtime/bcs-mesos/bcs-mesos-master && find . -type f ! -name MD5 | xargs -L1 md5sum > MD5

mnpack:
	cd ./build/bcs.${VERSION}/bcs-runtime/bcs-mesos/bcs-mesos-node && find . -type f ! -name MD5 | xargs -L1 md5sum > MD5

netpack:
	cd ./build/bcs.${VERSION}/bcs-runtime/bcs-k8s/bcs-network && find . -type f ! -name MD5 | xargs -L1 md5sum > MD5

pre:
	@echo "git tag: ${GITTAG}"
	mkdir -p ${PACKAGEPATH}
	mkdir -p ${EXPORTPATH}
	go fmt ./...
	cd ./scripts && chmod +x vet.sh && ./vet.sh

api:pre
	mkdir -p ${PACKAGEPATH}/bcs-services
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-api ${PACKAGEPATH}/bcs-services
	cd bcs-services/bcs-api && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-api/bcs-api ./main.go

gateway:pre
	mkdir -p ${PACKAGEPATH}/bcs-services
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-gateway-discovery ${PACKAGEPATH}/bcs-services
	cp -R ./bcs-services/bcs-gateway-discovery/plugins/apisix ${PACKAGEPATH}/bcs-services/bcs-gateway-discovery/
	cd bcs-services/bcs-gateway-discovery && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-gateway-discovery/bcs-gateway-discovery ./main.go

gateway-container: gateway
	cd ${PACKAGEPATH}/bcs-services/bcs-gateway-discovery/ && docker build -t bcs/apisix:${GITTAG} -f Dockerfile.apisix .
	cd ${PACKAGEPATH}/bcs-services/bcs-gateway-discovery/ && docker build -t bcs/bcs-gateway-discovery:${GITTAG} -f Dockerfile.gateway .


client:pre
	mkdir -p ${PACKAGEPATH}/bcs-services
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-client ${PACKAGEPATH}/bcs-services
	cd ./bcs-services/bcs-client && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-client/bcs-client ./cmd/main.go

dns:
	mkdir -p ${PACKAGEPATH}/bcs-services
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	cp -R ${BCS_CONF_MESOS_PATH}/bcs-mesos-master/bcs-dns ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-dns-service ${PACKAGEPATH}/bcs-services
	cd ../coredns && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-dns-service/bcs-dns-service coredns.go
	cd ../coredns && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master/bcs-dns/bcs-dns coredns.go

storage:pre
	mkdir -p ${PACKAGEPATH}/bcs-services
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-storage ${PACKAGEPATH}/bcs-services
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-services/bcs-storage/bcs-storage ./bcs-services/bcs-storage/storage.go

loadbalance:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master/bcs-loadbalance/bcs-loadbalance ${BCS_MESOS_PATH}/bcs-loadbalance/main.go
	cp -r ${BCS_MESOS_PATH}/bcs-loadbalance/image/* ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master/bcs-loadbalance/

executor:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-node
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-node/bcs-container-executor/bcs-container-executor ${BCS_MESOS_PATH}/bcs-container-executor/main.go

process-executor:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-node
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-node/bcs-process-executor/bcs-process-executor ${BCS_MESOS_PATH}/bcs-process-executor/main.go

process-daemon:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-node
	cp -R ${BCS_CONF_MESOS_PATH}/bcs-mesos-node/bcs-process-daemon ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-node
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-node/bcs-process-daemon/bcs-process-daemon ${BCS_MESOS_PATH}/bcs-process-daemon/main.go

netservice:pre
	mkdir -p ${PACKAGEPATH}/bcs-services
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-netservice ${PACKAGEPATH}/bcs-services
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-services/bcs-netservice/bcs-netservice ./bcs-services/bcs-netservice/main.go
	
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	cp -R ${BCS_CONF_MESOS_PATH}/bcs-mesos-master/bcs-netservice ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master/bcs-netservice/bcs-netservice ./bcs-services/bcs-netservice/main.go

	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-node/bcs-cni/bin/conf
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-node/bcs-cni/conf
	cp -R ${BCS_CONF_MESOS_PATH}/bcs-mesos-node/bcs-ipam/bcs.conf.template ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-node/bcs-cni/bin/conf
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-node/bcs-cni/bin/bcs-ipam ./bcs-services/bcs-netservice/bcs-ipam/main.go

mesos-driver:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	cp -R ${BCS_CONF_MESOS_PATH}/bcs-mesos-master/bcs-mesos-driver ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master/bcs-mesos-driver/bcs-mesos-driver ${BCS_MESOS_PATH}/bcs-mesos-driver/main.go

mesos-watch:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	cp -R ${BCS_CONF_MESOS_PATH}/bcs-mesos-master/bcs-mesos-watch ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master/bcs-mesos-watch/bcs-mesos-watch ${BCS_MESOS_PATH}/bcs-mesos-watch/main.go

kube-sche:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cp -R ${BCS_CONF_COMPONENT_PATH}/bcs-k8s-custom-scheduler ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cd ${BCS_COMPONENT_PATH}/bcs-k8s-custom-scheduler && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-k8s-custom-scheduler/bcs-k8s-custom-scheduler ./main.go

csi-cbs:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cp -R ${BCS_CONF_COMPONENT_PATH}/bcs-k8s-csi-tencentcloud ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-k8s-csi-tencentcloud/bcs-k8s-csi-tencentcloud ${BCS_COMPONENT_PATH}/bcs-k8s-csi-tencentcloud/cmd/cbs/main.go

scheduler:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	cp -R ${BCS_CONF_MESOS_PATH}/bcs-mesos-master/bcs-scheduler ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	cd ${BCS_MESOS_PATH}/bcs-scheduler && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master/bcs-scheduler/bcs-scheduler ./main.go && cd -
	cd ${BCS_MESOS_PATH}/bcs-scheduler && go build -buildmode=plugin -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master/bcs-scheduler/plugin/bin/ip-resources/ip-resources.so ./src/plugin/bin/ip-resources/ipResource.go && cd -
	cd ${BCS_MESOS_PATH}/bcs-scheduler && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master/bcs-scheduler/bcs-migrate-data ./bcs-migrate-data/main.go && cd -

logbeat-sidecar:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cp -R ${BCS_CONF_COMPONENT_PATH}/bcs-logbeat-sidecar ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cd ./bcs-runtime/bcs-k8s/bcs-component/bcs-logbeat-sidecar && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-logbeat-sidecar/bcs-logbeat-sidecar ./main.go

log-manager:pre
	mkdir -p ${PACKAGEPATH}/bcs-services
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-log-manager ${PACKAGEPATH}/bcs-services
	cd ./bcs-services/bcs-log-manager && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-log-manager/bcs-log-manager ./main.go

mesh-manager:pre
	mkdir -p ${PACKAGEPATH}/bcs-services
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-mesh-manager ${PACKAGEPATH}/bcs-services
	cd bcs-services/bcs-mesh-manager && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-mesh-manager/bcs-mesh-manager ./main.go

hpacontroller:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	cp -R ${BCS_CONF_MESOS_PATH}/bcs-mesos-master/bcs-hpacontroller ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master/bcs-hpacontroller/bcs-hpacontroller ${BCS_MESOS_PATH}/bcs-hpacontroller

sd-prometheus:pre
	mkdir -p ${PACKAGEPATH}/bcs-services
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-service-prometheus-service ${PACKAGEPATH}/bcs-services
	cp -R ${BCS_CONF_MESOS_PATH}/bcs-mesos-master/bcs-service-prometheus ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-services/bcs-service-prometheus-service/bcs-service-prometheus-service ./bcs-services/bcs-service-prometheus/main.go
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master/bcs-service-prometheus/bcs-service-prometheus ./bcs-services/bcs-service-prometheus/main.go

k8s-driver:pre
	cd ${BCS_COMPONENT_PATH}/bcs-k8s-driver && make k8s-driver

gamestatefulset:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cp -R ${BCS_CONF_COMPONENT_PATH}/bcs-gamestatefulset-operator ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cd ${BCS_COMPONENT_PATH}/bcs-gamestatefulset-operator && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-gamestatefulset-operator/bcs-gamestatefulset-operator ./cmd/gamestatefulset-operator/main.go

gamedeployment:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cp -R ${BCS_CONF_COMPONENT_PATH}/bcs-gamedeployment-operator ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cd ${BCS_COMPONENT_PATH}/bcs-gamedeployment-operator && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/bcs-gamedeployment-operator ./cmd/gamedeployment-operator/main.go

hook-operator:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cp -R ${BCS_CONF_COMPONENT_PATH}/bcs-hook-operator ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cd ${BCS_COMPONENT_PATH}/bcs-hook-operator && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-hook-operator/bcs-hook-operator ./cmd/hook-operator/main.go

federated-apiserver:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cp -R ${BCS_CONF_COMPONENT_PATH}/bcs-federated-apiserver ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cd ${BCS_COMPONENT_PATH}/bcs-federated-apiserver && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-federated-apiserver/bcs-federated-apiserver ./cmd/apiserver/main.go

federated-apiserver-kubectl-agg:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cd ${BCS_COMPONENT_PATH}/bcs-federated-apiserver && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-federated-apiserver/kubectl-agg ./cmd/kubectl-agg/main.go

egress-controller:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cp -R ${BCS_CONF_COMPONENT_PATH}/bcs-egress-controller ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	#copy nginx template for egress controller
	cp -R ${BCS_COMPONENT_PATH}/bcs-egress/deploy/config ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-egress-controller
	cd ${BCS_COMPONENT_PATH}/bcs-egress && go build -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-egress-controller/bcs-egress-controller ./cmd/bcs-egress-controller/main.go

webhook-server:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cp -R ${BCS_CONF_COMPONENT_PATH}/bcs-webhook-server ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cd ${BCS_COMPONENT_PATH}/bcs-webhook-server && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/bcs-webhook-server ./cmd/server.go


consoleproxy:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-node
	cp -R ${BCS_CONF_MESOS_PATH}/bcs-mesos-node/bcs-consoleproxy ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-node
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-node/bcs-consoleproxy/bcs-consoleproxy ${BCS_MESOS_PATH}/bcs-consoleproxy/main.go

bmsf-mesos-adapter:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	cp -R ${BCS_CONF_MESOS_PATH}/bcs-mesos-master/bmsf-mesos-adapter ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	cd ${BCS_MESOS_PATH}/bmsf-mesh && go build ${LDFLAG} -o ../${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master/bmsf-mesos-adapter/bmsf-mesos-adapter ./bmsf-mesos-adapter/main.go

cpuset:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cp -R ${BCS_CONF_COMPONENT_PATH}/bcs-cpuset-device ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-cpuset-device/bcs-cpuset-device ${BCS_COMPONENT_PATH}/bcs-cpuset-device/main.go

detection:pre
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-network-detection ${PACKAGEPATH}/bcs-services
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-services/bcs-network-detection/bcs-network-detection ./bcs-services/bcs-network-detection/main.go

tools:
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-services/cryptools ./install/cryptool/main.go
	
user-manager:pre
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-user-manager
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-user-manager ${PACKAGEPATH}/bcs-services
	cd bcs-services/bcs-user-manager/ && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-user-manager/bcs-user-manager ./main.go

k8s-watch:pre
	mkdir -p ${PACKAGEPATH}/bcs-services
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-k8s-watch ${PACKAGEPATH}/bcs-services
	cd ./bcs-services/bcs-k8s-watch && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-k8s-watch/bcs-k8s-watch  ./main.go

kube-agent:pre
	mkdir -p ${PACKAGEPATH}/bcs-services
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-kube-agent ${PACKAGEPATH}/bcs-services
	cd ./bcs-services/bcs-kube-agent && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-kube-agent/bcs-kube-agent  ./main.go

cc-agent:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cp -R ${BCS_CONF_COMPONENT_PATH}/bcs-cc-agent ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-cc-agent/bcs-cc-agent ${BCS_COMPONENT_PATH}/bcs-cc-agent/main.go

clusternet-controller:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cp -R ${BCS_CONF_COMPONENT_PATH}/bcs-clusternet-controller ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cd ${BCS_COMPONENT_PATH}/bcs-clusternet-controller && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-clusternet-controller/bcs-clusternet-controller ./cmd/clusternet-controller/main.go

# network plugins section
networkpolicy:pre
	cd ${BCS_NETWORK_PATH} && make networkpolicy

cloud-network-agent:pre
	cd ${BCS_NETWORK_PATH} && make bcs-cloud-network-agent

bkcmdb-synchronizer:pre
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-bkcmdb-synchronizer
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-services/bcs-bkcmdb-synchronizer/bcs-bkcmdb-synchronizer ./bcs-services/bcs-bkcmdb-synchronizer/main.go

cloud-netservice:pre
	cd ${BCS_NETWORK_PATH} && make cloud-netservice

cloud-netcontroller:pre
	cd ${BCS_NETWORK_PATH} && make cloud-netcontroller

cloud-netagent:pre
	cd ${BCS_NETWORK_PATH} && make cloud-netagent

ingress-controller:pre
	cd ${BCS_NETWORK_PATH} && make ingress-controller

ipmasq-cidrsync:pre
	cd ${BCS_NETWORK_PATH} && make ipmasq-cidrsync

ipres-webhook:pre
	cd ${BCS_NETWORK_PATH} && make ipres-webhook

network:pre
	cd ${BCS_NETWORK_PATH} && make network

clb-controller:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	cp -R ${BCS_CONF_MESOS_PATH}/bcs-mesos-master/bcs-clb-controller ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	cp ${BCS_MESOS_PATH}/bcs-clb-controller/docker/Dockerfile ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master/bcs-clb-controller/Dockerfile.old
	cd ${BCS_MESOS_PATH}/bcs-clb-controller && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master/bcs-clb-controller/bcs-clb-controller ./main.go && cd -
	cp ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master/bcs-clb-controller/bcs-clb-controller  ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master/bcs-clb-controller/clb-controller

gw-controller:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	cp -R ${BCS_CONF_MESOS_PATH}/bcs-mesos-master/bcs-gw-controller ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	cd ${BCS_MESOS_PATH}/bcs-clb-controller && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master/bcs-gw-controller/bcs-gw-controller ./bcs-gw-controller/main.go

#end of network plugins

# bcs-service section
cluster-manager:pre
	cd ./bcs-services/bcs-cluster-manager && make clustermanager

alert-manager:pre
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-alert-manager/swagger
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-alert-manager/*  ${PACKAGEPATH}/bcs-services/bcs-alert-manager
	cp -R ./bcs-services/bcs-alert-manager/pkg/third_party/swagger-ui ${PACKAGEPATH}/bcs-services/bcs-alert-manager/swagger/swagger-ui
	cp ./bcs-services/bcs-alert-manager/pkg/proto/alertmanager/alertmanager.swagger.json ${PACKAGEPATH}/bcs-services/bcs-alert-manager/swagger/alertmanager.swagger.json
	cd ./bcs-services/bcs-alert-manager/ && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-alert-manager/bcs-alert-manager ./main.go

# end of bcs-service section

apiserver-proxy:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-apiserver-proxy
	cp -R ${BCS_CONF_COMPONENT_PATH}/bcs-apiserver-proxy/* ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-apiserver-proxy
	cd ${BCS_COMPONENT_PATH}/bcs-apiserver-proxy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-apiserver-proxy/bcs-apiserver-proxy ./main.go

apiserver-proxy-tools:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-apiserver-proxy
	cd ${BCS_COMPONENT_PATH}/bcs-apiserver-proxy/ipvs_tools && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-apiserver-proxy/apiserver-proxy-tools .
