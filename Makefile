# BlueKing Container System Makefile
# default config
MAKE:=make
bcs_edition?=inner_edition

# init the build information
ifdef HASTAG
	GITTAG=$(shell git describe --tags)
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
default:api client storage executor mesos-driver mesos-watch scheduler \
	loadbalance metricservice metriccollector k8s-watch kube-agent k8s-driver \
	netservice sd-prometheus process-executor process-daemon bmsf-mesos-adapter \
	hpacontroller kube-sche consoleproxy clb-controller gw-controller logbeat-sidecar \
	csi-cbs bcs-webhook-server gamestatefulset network detection cpuset bcs-networkpolicy \
	tools gateway user-manager cc-agent bkcmdb-synchronizer bcs-cloud-netservice bcs-cloud-netcontroller \
	bcs-cloud-netagent mesh-manager bcs-ingress-controller log-manager gamedeployment
k8s:api client storage k8s-watch kube-agent k8s-driver csi-cbs kube-sche gamestatefulset gamedeployment
mesos:api client storage dns mesos-driver mesos-watch scheduler loadbalance netservice hpacontroller \
	consoleproxy clb-controller

allpack: svcpack k8spack mmpack mnpack
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

pre:
	@echo "git tag: ${GITTAG}"
	mkdir -p ${PACKAGEPATH}
	mkdir -p ${EXPORTPATH}
	go fmt ./...
	cd ./scripts && chmod +x vet.sh && ./vet.sh

api:pre
	mkdir -p ${PACKAGEPATH}/bcs-services
	cp -R ./install/conf/bcs-services/bcs-api ${PACKAGEPATH}/bcs-services
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-services/bcs-api/bcs-api ./bcs-services/bcs-api/main.go

gateway:pre
	mkdir -p ${PACKAGEPATH}/bcs-services
	cp -R ./install/conf/bcs-services/bcs-gateway-discovery ${PACKAGEPATH}/bcs-services
	cp -R ./bcs-services/bcs-gateway-discovery/bkbcs-auth ${PACKAGEPATH}/bcs-services/bcs-gateway-discovery/
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-services/bcs-gateway-discovery/bcs-gateway-discovery ./bcs-services/bcs-gateway-discovery/main.go

kube-agent:pre
	mkdir -p ${PACKAGEPATH}/bcs-k8s-master
	cp -R ./install/conf/bcs-k8s-master/bcs-kube-agent ${PACKAGEPATH}/bcs-k8s-master
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-k8s-master/bcs-kube-agent/bcs-kube-agent ./bcs-k8s/bcs-kube-agent/main.go

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

metricservice:pre
	mkdir -p ${PACKAGEPATH}/bcs-services
	cp -R ./install/conf/bcs-services/bcs-metricservice ${PACKAGEPATH}/bcs-services
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-services/bcs-metricservice/bcs-metricservice ./bcs-services/bcs-metricservice/main.go

metriccollector:pre
	mkdir -p ${PACKAGEPATH}/bcs-mesos-node
	cp -R ./install/conf/bcs-mesos-node/bcs-metriccollector ${PACKAGEPATH}/bcs-mesos-node
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-mesos-node/bcs-metriccollector/bcs-metriccollector ./bcs-services/bcs-metriccollector/main.go

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
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-mesos-master/bcs-scheduler/bcs-scheduler ./bcs-mesos/bcs-scheduler
	go build -buildmode=plugin -o ${PACKAGEPATH}/bcs-mesos-master/bcs-scheduler/plugin/bin/ip-resources/ip-resources.so ./bcs-mesos/bcs-scheduler/src/plugin/bin/ip-resources/ipResource.go

logbeat-sidecar:pre
	mkdir -p ${PACKAGEPATH}/bcs-services
	cp -R ./install/conf/bcs-services/bcs-logbeat-sidecar ${PACKAGEPATH}/bcs-services
	cd ./bcs-services/bcs-logbeat-sidecar && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-logbeat-sidecar/bcs-logbeat-sidecar ./main.go && cd -

log-manager:pre
	mkdir -p ${PACKAGEPATH}/bcs-services
	cp -R ./install/conf/bcs-services/bcs-log-manager ${PACKAGEPATH}/bcs-services
	cd ./bcs-services/bcs-log-manager && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-log-manager/bcs-log-manager ./main.go && cd -

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
	mkdir -p ${PACKAGEPATH}/bcs-k8s-master
	cp -R ./install/conf/bcs-k8s-master/bcs-k8s-driver ${PACKAGEPATH}/bcs-k8s-master
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-k8s-master/bcs-k8s-driver/bcs-k8s-driver ./bcs-k8s/bcs-k8s-driver/main.go

k8s-watch:pre
	mkdir -p ${PACKAGEPATH}/bcs-k8s-master
	cp -R ./install/conf/bcs-k8s-master/bcs-k8s-watch ${PACKAGEPATH}/bcs-k8s-master
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-k8s-master/bcs-k8s-watch/bcs-k8s-watch ./bcs-k8s/bcs-k8s-watch/main.go

gamestatefulset:pre
	mkdir -p ${PACKAGEPATH}/bcs-k8s-master
	cp -R ./install/conf/bcs-k8s-master/bcs-gamestatefulset-operator ${PACKAGEPATH}/bcs-k8s-master
	cd bcs-k8s/bcs-gamestatefulset-operator && go build -o ${WORKSPACE}/${PACKAGEPATH}/bcs-k8s-master/bcs-gamestatefulset-operator/bcs-gamestatefulset-operator ./cmd/gamestatefulset-operator/main.go

gamedeployment:pre
	mkdir -p ${PACKAGEPATH}/bcs-k8s-master
	cp -R ./install/conf/bcs-k8s-master/bcs-gamedeployment-operator ${PACKAGEPATH}/bcs-k8s-master
	cd bcs-k8s/bcs-gamedeployment-operator && go build -o ${WORKSPACE}/${PACKAGEPATH}/bcs-k8s-master/bcs-gamedeployment-operator/bcs-gamedeployment-operator ./cmd/gamedeployment-operator/main.go

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
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-mesos-master/bmsf-mesos-adapter/bmsf-mesos-adapter ./bmsf-mesh/bmsf-mesos-adapter/main.go

network:pre
	cd ./bcs-network && make network && cd -

clb-controller:pre
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-clb-controller
	cp -R ./install/conf/bcs-services/bcs-clb-controller ${PACKAGEPATH}/bcs-services
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-services/bcs-clb-controller/bcs-clb-controller ./bcs-services/bcs-clb-controller/main.go

cpuset:pre
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-cpuset-device
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-services/bcs-cpuset-device/bcs-cpuset-device ./bcs-services/bcs-cpuset-device/main.go

gw-controller:pre
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-gw-controller
	cp -R ./install/conf/bcs-services/bcs-gw-controller ${PACKAGEPATH}/bcs-services
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-services/bcs-gw-controller/bcs-gw-controller ./bcs-services/bcs-gw-controller/main.go

bcs-webhook-server:pre
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-webhook-server
	cp ./install/conf/bcs-services/bcs-webhook-server/* ${PACKAGEPATH}/bcs-services/bcs-webhook-server
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-services/bcs-webhook-server/bcs-webhook-server ./bcs-services/bcs-webhook-server/main.go

detection:pre
	cp -R ./install/conf/bcs-services/bcs-network-detection ${PACKAGEPATH}/bcs-services
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-services/bcs-network-detection/bcs-network-detection ./bcs-services/bcs-network-detection/main.go

tools:
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-services/cryptools ./install/cryptool/main.go
	
bcs-networkpolicy:pre
	cd ./bcs-network && make networkpolicy && cd -

bcs-cloud-network-agent:pre
	cd ./bcs-network && make bcs-cloud-network-agent && cd -
	
user-manager:pre
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-user-manager
	cp -R ./install/conf/bcs-services/bcs-user-manager ${PACKAGEPATH}/bcs-services
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-services/bcs-user-manager/bcs-user-manager ./bcs-services/bcs-user-manager/main.go

cc-agent:pre
	mkdir -p ${PACKAGEPATH}/bcs-k8s-master
	cp -R ./install/conf/bcs-k8s-master/bcs-cc-agent ${PACKAGEPATH}/bcs-k8s-master
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-k8s-master/bcs-cc-agent/bcs-cc-agent ./bcs-k8s/bcs-cc-agent/main.go

bkcmdb-synchronizer:pre
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-bkcmdb-synchronizer
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-services/bcs-bkcmdb-synchronizer/bcs-bkcmdb-synchronizer ./bcs-services/bcs-bkcmdb-synchronizer/main.go

bcs-cloud-netservice:pre
	cd ./bcs-network && make cloud-netservice && cd -

bcs-cloud-netcontroller:pre
	cd ./bcs-network && make cloud-netcontroller && cd -

bcs-cloud-netagent:pre
	cd ./bcs-network && make cloud-netagent && cd -

bcs-ingress-controller:pre
	cd ./bcs-network && make ingress-controller && cd -
