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

BCS_UI_PATH=${WORKSPACE}/bcs-ui
BCS_SERVICES_PATH=${WORKSPACE}/bcs-services
BCS_INSTALL_PATH=${WORKSPACE}/install
BCS_NETWORK_PATH=${WORKSPACE}/bcs-runtime/bcs-k8s/bcs-network
BCS_COMPONENT_PATH=${WORKSPACE}/bcs-runtime/bcs-k8s/bcs-component
BCS_MESOS_PATH=${WORKSPACE}/bcs-runtime/bcs-mesos
BCS_CONF_UI_PATH=${WORKSPACE}/install/conf/bcs-ui
BCS_CONF_COMPONENT_PATH=${WORKSPACE}/install/conf/bcs-runtime/bcs-k8s/bcs-component
BCS_CONF_NETWORK_PATH=${WORKSPACE}/install/conf/bcs-runtime/bcs-k8s/bcs-network
BCS_CONF_MESOS_PATH=${WORKSPACE}/install/conf/bcs-runtime/bcs-mesos
BCS_CONF_SERVICES_PATH=${WORKSPACE}/install/conf/bcs-services

export LDFLAG=-ldflags "-X github.com/Tencent/bk-bcs/bcs-common/common/static.ZookeeperClientUser=${bcs_zk_client_user} \
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
export PACKAGEPATH=./build/bcs.${VERSION}
export SCENARIOSPACKAGE=${WORKSPACE}/${PACKAGEPATH}/bcs-scenarios

# bscp 应用自定义
export BSCP_LDFLAG=-ldflags "-X bscp.io/pkg/version.BUILDTIME=${BUILDTIME} -X bscp.io/pkg/version.GITHASH=${GITHASH}"

# tongsuo related environment variables
export TONGSUO_PATH?=$(WORKSPACE)/build/bcs.${VERSION}/tongsuo
export IS_STATIC?=true

ifeq ($(IS_STATIC),true)
        CGO_BUILD_FLAGS= CGO_ENABLED=1 CGO_CFLAGS="-I${TONGSUO_PATH}/include -Wno-deprecated-declarations" \
        CGO_LDFLAGS="-L${TONGSUO_PATH}/lib -lssl -lcrypto -ldl -lpthread -static-libgcc -static-libstdc++"
else
        CGO_BUILD_FLAGS= CGO_ENABLED=1 CGO_CFLAGS="-I${TONGSUO_PATH}/include -Wno-deprecated-declarations" \
        CGO_LDFLAGS="-L${TONGSUO_PATH}/lib -lssl -lcrypto"
endif

# options
default:bcs-runtime bcs-scenarios bcs-services #TODO: bcs-resources

bcs-runtime: bcs-k8s bcs-mesos

bcs-k8s: bcs-component bcs-network

bcs-component:k8s-driver \
	cc-agent kube-sche apiserver-proxy \
	apiserver-proxy-tools logbeat-sidecar webhook-server \
	general-pod-autoscaler cluster-autoscaler

bcs-network:network networkpolicy cloud-netservice cloud-netcontroller cloud-netagent

bcs-mesos:executor mesos-driver mesos-watch scheduler loadbalance netservice hpacontroller \
	consoleproxy process-executor process-daemon bmsf-mesos-adapter detection clb-controller gw-controller

bcs-services:api client bkcmdb-synchronizer cpuset gateway log-manager \
	netservice sd-prometheus storage \
	user-manager cluster-manager tools alert-manager k8s-watch kube-agent data-manager \
	helm-manager project-manager nodegroup-manager

bcs-scenarios: kourse gitops

kourse: gamedeployment gamestatefulset hook-operator

gitops: gitops-proxy gitops-manager 

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
	go mod tidy
	go fmt ./...
	cd ./scripts && chmod +x vet.sh && ./vet.sh
	cd ./scripts && chmod +x tongsuo.sh && ./tongsuo.sh

api:pre
	mkdir -p ${PACKAGEPATH}/bcs-services
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-api ${PACKAGEPATH}/bcs-services
	cd bcs-services/bcs-api && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-api/bcs-api ./main.go

gateway:pre
	mkdir -p ${PACKAGEPATH}/bcs-services
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-gateway-discovery ${PACKAGEPATH}/bcs-services
	cp -R ./bcs-services/bcs-gateway-discovery/plugins/apisix ${PACKAGEPATH}/bcs-services/bcs-gateway-discovery/
	cd bcs-services/bcs-gateway-discovery && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-gateway-discovery/bcs-gateway-discovery ./main.go

micro-gateway:pre
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-micro-gateway
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-gateway-discovery/* ${PACKAGEPATH}/bcs-services/bcs-micro-gateway/
	cp -R ./bcs-services/bcs-gateway-discovery/plugins/apisix ${PACKAGEPATH}/bcs-services/bcs-micro-gateway/

client:pre
	mkdir -p ${PACKAGEPATH}/bcs-services
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-client ${PACKAGEPATH}/bcs-services
	cd ./bcs-services/bcs-client && go mod tidy -go=1.16 && go mod tidy -go=1.17 && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-client/bcs-client ./cmd/main.go

dns:
	mkdir -p ${PACKAGEPATH}/bcs-services
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	cp -R ${BCS_CONF_MESOS_PATH}/bcs-mesos-master/bcs-dns ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-dns-service ${PACKAGEPATH}/bcs-services
	cd ../coredns && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-dns-service/bcs-dns-service coredns.go
	cd ../coredns && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master/bcs-dns/bcs-dns coredns.go

storage:pre
	mkdir -p ${PACKAGEPATH}/bcs-services
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-storage ${PACKAGEPATH}/bcs-services
	cd ./bcs-services/bcs-storage && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-storage/bcs-storage ./storage.go

loadbalance:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master/bcs-loadbalance/bcs-loadbalance ${BCS_MESOS_PATH}/bcs-loadbalance/main.go
	cp -r ${BCS_MESOS_PATH}/bcs-loadbalance/image/* ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master/bcs-loadbalance/

executor:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-node
	go mod tidy && go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-node/bcs-container-executor/bcs-container-executor ${BCS_MESOS_PATH}/bcs-container-executor/main.go

process-executor:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-node
	go mod tidy && go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-node/bcs-process-executor/bcs-process-executor ${BCS_MESOS_PATH}/bcs-process-executor/main.go

process-daemon:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-node
	cp -R ${BCS_CONF_MESOS_PATH}/bcs-mesos-node/bcs-process-daemon ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-node
	go mod tidy && go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-node/bcs-process-daemon/bcs-process-daemon ${BCS_MESOS_PATH}/bcs-process-daemon/main.go

netservice:pre
	mkdir -p ${PACKAGEPATH}/bcs-services
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-netservice ${PACKAGEPATH}/bcs-services
	cd ./bcs-services/bcs-netservice && go mod tidy  && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-netservice/bcs-netservice ./main.go

	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	cp -R ${BCS_CONF_MESOS_PATH}/bcs-mesos-master/bcs-netservice ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	cd ./bcs-services/bcs-netservice && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master/bcs-netservice/bcs-netservice ./main.go

	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-node/bcs-cni/bin/conf
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-node/bcs-cni/conf
	cp -R ${BCS_CONF_MESOS_PATH}/bcs-mesos-node/bcs-ipam/bcs.conf.template ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-node/bcs-cni/bin/conf
	cd ./bcs-services/bcs-netservice/bcs-ipam && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-node/bcs-cni/bin/bcs-ipam ./main.go

mesos-driver:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	cp -R ${BCS_CONF_MESOS_PATH}/bcs-mesos-master/bcs-mesos-driver ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	go mod tidy && go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master/bcs-mesos-driver/bcs-mesos-driver ${BCS_MESOS_PATH}/bcs-mesos-driver/main.go

mesos-watch:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	cp -R ${BCS_CONF_MESOS_PATH}/bcs-mesos-master/bcs-mesos-watch ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	go mod tidy && go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master/bcs-mesos-watch/bcs-mesos-watch ${BCS_MESOS_PATH}/bcs-mesos-watch/main.go

kube-sche:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cp -R ${BCS_CONF_COMPONENT_PATH}/bcs-k8s-custom-scheduler ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cd ${BCS_COMPONENT_PATH}/bcs-k8s-custom-scheduler && go mod tidy -go=1.16 && go mod tidy -go=1.17 && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-k8s-custom-scheduler/bcs-k8s-custom-scheduler ./main.go

scheduler:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	cp -R ${BCS_CONF_MESOS_PATH}/bcs-mesos-master/bcs-scheduler ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	cd ${BCS_MESOS_PATH}/bcs-scheduler && go mod tidy -go=1.16 && go mod tidy -go=1.17 && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master/bcs-scheduler/bcs-scheduler ./main.go && cd -
	cd ${BCS_MESOS_PATH}/bcs-scheduler && go mod tidy -go=1.16 && go mod tidy -go=1.17 && go build -buildmode=plugin -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master/bcs-scheduler/plugin/bin/ip-resources/ip-resources.so ./src/plugin/bin/ip-resources/ipResource.go && cd -
	cd ${BCS_MESOS_PATH}/bcs-scheduler && go mod tidy -go=1.16 && go mod tidy -go=1.17 && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master/bcs-scheduler/bcs-migrate-data ./bcs-migrate-data/main.go && cd -

logbeat-sidecar:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cp -R ${BCS_CONF_COMPONENT_PATH}/bcs-logbeat-sidecar ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cd ./bcs-runtime/bcs-k8s/bcs-component/bcs-logbeat-sidecar && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-logbeat-sidecar/bcs-logbeat-sidecar ./main.go

multi-ns-proxy:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cp -R ${BCS_CONF_COMPONENT_PATH}/bcs-multi-ns-proxy ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cd ./bcs-runtime/bcs-k8s/bcs-component/bcs-multi-ns-proxy  && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-multi-ns-proxy/bcs-multi-ns-proxy ./main.go

log-manager:pre
	mkdir -p ${PACKAGEPATH}/bcs-services
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-log-manager ${PACKAGEPATH}/bcs-services
	cd ./bcs-services/bcs-log-manager && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-log-manager/bcs-log-manager ./main.go

hpacontroller:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	cp -R ${BCS_CONF_MESOS_PATH}/bcs-mesos-master/bcs-hpacontroller ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	go mod tidy && go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master/bcs-hpacontroller/bcs-hpacontroller ${BCS_MESOS_PATH}/bcs-hpacontroller

sd-prometheus:pre
	mkdir -p ${PACKAGEPATH}/bcs-services
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-service-prometheus-service ${PACKAGEPATH}/bcs-services
	cp -R ${BCS_CONF_MESOS_PATH}/bcs-mesos-master/bcs-service-prometheus ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	go mod tidy && go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-services/bcs-service-prometheus-service/bcs-service-prometheus-service ./bcs-services/bcs-service-prometheus/main.go
	go mod tidy && go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master/bcs-service-prometheus/bcs-service-prometheus ./bcs-services/bcs-service-prometheus/main.go

k8s-driver:pre
	cd ${BCS_COMPONENT_PATH}/bcs-k8s-driver && go mod tidy -go=1.16 && go mod tidy -go=1.17 && make k8s-driver

egress-controller:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cp -R ${BCS_CONF_COMPONENT_PATH}/bcs-egress-controller ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	#copy nginx template for egress controller
	cp -R ${BCS_COMPONENT_PATH}/bcs-egress/deploy/config ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-egress-controller
	cd ${BCS_COMPONENT_PATH}/bcs-egress && go mod tidy && go build -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-egress-controller/bcs-egress-controller ./cmd/bcs-egress-controller/main.go

webhook-server:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cp -R ${BCS_CONF_COMPONENT_PATH}/bcs-webhook-server ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cd ${BCS_COMPONENT_PATH}/bcs-webhook-server && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/bcs-webhook-server ./cmd/server.go


consoleproxy:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-node
	cp -R ${BCS_CONF_MESOS_PATH}/bcs-mesos-node/bcs-consoleproxy ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-node
	go mod tidy && go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-node/bcs-consoleproxy/bcs-consoleproxy ${BCS_MESOS_PATH}/bcs-consoleproxy/main.go

bmsf-mesos-adapter:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	cp -R ${BCS_CONF_MESOS_PATH}/bcs-mesos-master/bmsf-mesos-adapter ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	cd ${BCS_MESOS_PATH}/bmsf-mesh && go mod tidy && go build ${LDFLAG} -o ../${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master/bmsf-mesos-adapter/bmsf-mesos-adapter ./bmsf-mesos-adapter/main.go

cpuset:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cp -R ${BCS_CONF_COMPONENT_PATH}/bcs-cpuset-device ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	go mod tidy && go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-cpuset-device/bcs-cpuset-device ${BCS_COMPONENT_PATH}/bcs-cpuset-device/main.go

detection:pre
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-network-detection ${PACKAGEPATH}/bcs-services
	go mod tidy && go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-services/bcs-network-detection/bcs-network-detection ./bcs-services/bcs-network-detection/main.go

tools:pre
	mkdir -p ${PACKAGEPATH}/bcs-services
	cd ${BCS_INSTALL_PATH}/cryptool && go mod tidy && $(CGO_BUILD_FLAGS) go build ${LDFLAG} -o  ${WORKSPACE}/${PACKAGEPATH}/bcs-services/cryptools main.go

ui:pre
	mkdir -p ${PACKAGEPATH}/bcs-ui
	cp -R ${BCS_CONF_UI_PATH} ${PACKAGEPATH}
	cd ${BCS_UI_PATH} && ls -la && cd frontend && npm install && npm run build && cd ../ && go mod tidy -compat=1.17 && CGO_ENABLED=0 go build -trimpath ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-ui/bcs-ui ./cmd/bcs-ui

user-manager:pre
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-user-manager
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-user-manager ${PACKAGEPATH}/bcs-services
	cd bcs-services/bcs-user-manager/ && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-user-manager/bcs-user-manager ./main.go

webconsole:pre
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-webconsole
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-webconsole ${PACKAGEPATH}/bcs-services
	cd bcs-services/bcs-webconsole/ && go mod tidy -compat=1.17 && CGO_ENABLED=0 go build -trimpath ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-webconsole/bcs-webconsole ./main.go

monitor:pre
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-monitor
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-monitor ${PACKAGEPATH}/bcs-services
	cd bcs-services/bcs-monitor/ && go mod tidy -compat=1.17 && CGO_ENABLED=0 go build -trimpath ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-monitor/bcs-monitor ./cmd/bcs-monitor

bscp:pre
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-bscp
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-bscp ${PACKAGEPATH}/bcs-services
	cd bcs-services/bcs-bscp && cd ui && npm install && npm run build && cd ../ && go mod tidy -compat=1.20 && CGO_ENABLED=0 go build -trimpath ${BSCP_LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-bscp/hyper/bscp-ui ./cmd/ui
	cd bcs-services/bcs-bscp && go mod tidy -compat=1.20 && CGO_ENABLED=0 go build -trimpath ${BSCP_LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-bscp/hyper/bk-bscp-apiserver ./cmd/api-server
	cd bcs-services/bcs-bscp && go mod tidy -compat=1.20 && CGO_ENABLED=0 go build -trimpath ${BSCP_LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-bscp/hyper/bk-bscp-authserver ./cmd/auth-server
	cd bcs-services/bcs-bscp && go mod tidy -compat=1.20 && CGO_ENABLED=0 go build -trimpath ${BSCP_LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-bscp/hyper/bk-bscp-configserver ./cmd/config-server
	cd bcs-services/bcs-bscp && go mod tidy -compat=1.20 && CGO_ENABLED=0 go build -trimpath ${BSCP_LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-bscp/hyper/bk-bscp-dataservice ./cmd/data-service
	cd bcs-services/bcs-bscp && go mod tidy -compat=1.20 && CGO_ENABLED=0 go build -trimpath ${BSCP_LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-bscp/hyper/bk-bscp-feedserver ./cmd/feed-server
	cd bcs-services/bcs-bscp && go mod tidy -compat=1.20 && CGO_ENABLED=0 go build -trimpath ${BSCP_LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-bscp/hyper/bk-bscp-cacheservice ./cmd/cache-service
	# alias docker image name to bk-bscp-hyper
	touch ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-bscp/bk-bscp-hyper && chmod a+x ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-bscp/bk-bscp-hyper && ls -la ${PACKAGEPATH}/bcs-services/bcs-bscp/hyper

k8s-watch:pre
	mkdir -p ${PACKAGEPATH}/bcs-services
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-k8s-watch ${PACKAGEPATH}/bcs-services
	cd ./bcs-services/bcs-k8s-watch && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-k8s-watch/bcs-k8s-watch  ./main.go

kube-agent:pre
	mkdir -p ${PACKAGEPATH}/bcs-services
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-kube-agent ${PACKAGEPATH}/bcs-services
	cd ./bcs-services/bcs-kube-agent && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-kube-agent/bcs-kube-agent  ./main.go

cc-agent:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cp -R ${BCS_CONF_COMPONENT_PATH}/bcs-cc-agent ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	go mod tidy && go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-cc-agent/bcs-cc-agent ${BCS_COMPONENT_PATH}/bcs-cc-agent/main.go

general-pod-autoscaler:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cp -R ${BCS_CONF_COMPONENT_PATH}/bcs-general-pod-autoscaler ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cd ${BCS_COMPONENT_PATH}/bcs-general-pod-autoscaler && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-general-pod-autoscaler/bcs-general-pod-autoscaler ./cmd/gpa/main.go

cluster-autoscaler:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/
	cp -R ${BCS_CONF_COMPONENT_PATH}/bcs-cluster-autoscaler ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cd ${BCS_COMPONENT_PATH}/bcs-cluster-autoscaler/bcs-cluster-autoscaler-1.16 && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/hyper/bcs-cluster-autoscaler-1.16 ./main.go
	cd ${BCS_COMPONENT_PATH}/bcs-cluster-autoscaler/bcs-cluster-autoscaler-1.22 && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/hyper/bcs-cluster-autoscaler-1.22 ./main.go
	touch ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/bcs-cluster-autoscaler && chmod a+x ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/bcs-cluster-autoscaler && ls -la ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/hyper

# network plugins section
networkpolicy:pre
	cd ${BCS_NETWORK_PATH} && go mod tidy && make networkpolicy

cloud-network-agent:pre
	cd ${BCS_NETWORK_PATH} && go mod tidy && make bcs-cloud-network-agent

bkcmdb-synchronizer:pre
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-bkcmdb-synchronizer
	go mod tidy && go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-services/bcs-bkcmdb-synchronizer/bcs-bkcmdb-synchronizer ./bcs-services/bcs-bkcmdb-synchronizer/main.go

cloud-netservice:pre
	cd ${BCS_NETWORK_PATH} && go mod tidy && make cloud-netservice

cloud-netcontroller:pre
	cd ${BCS_NETWORK_PATH} && go mod tidy && make cloud-netcontroller

cloud-netagent:pre
	cd ${BCS_NETWORK_PATH} && go mod tidy && make cloud-netagent

ingress-controller:pre
	cd ${BCS_NETWORK_PATH} && go mod tidy && make ingress-controller

ipmasq-cidrsync:pre
	cd ${BCS_NETWORK_PATH} && go mod tidy && make ipmasq-cidrsync

ipres-webhook:pre
	cd ${BCS_NETWORK_PATH} && go mod tidy && make ipres-webhook

network:pre
	cd ${BCS_NETWORK_PATH} && go mod tidy && make network

clb-controller:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	cp -R ${BCS_CONF_MESOS_PATH}/bcs-mesos-master/bcs-clb-controller ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	cp ${BCS_MESOS_PATH}/bcs-clb-controller/docker/Dockerfile ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master/bcs-clb-controller/Dockerfile.old
	cd ${BCS_MESOS_PATH}/bcs-clb-controller && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master/bcs-clb-controller/bcs-clb-controller ./main.go && cd -
	cp ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master/bcs-clb-controller/bcs-clb-controller  ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master/bcs-clb-controller/clb-controller

gw-controller:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	cp -R ${BCS_CONF_MESOS_PATH}/bcs-mesos-master/bcs-gw-controller ${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master
	cd ${BCS_MESOS_PATH}/bcs-clb-controller && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-mesos/bcs-mesos-master/bcs-gw-controller/bcs-gw-controller ./bcs-gw-controller/main.go

#end of network plugins

# bcs-service section
cluster-manager:pre
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-cluster-manager
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-cluster-manager/* ${PACKAGEPATH}/bcs-services/bcs-cluster-manager/
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-cluster-manager/swagger
	cp -R ${BCS_SERVICES_PATH}/bcs-cluster-manager/third_party/swagger-ui ${PACKAGEPATH}/bcs-services/bcs-cluster-manager/swagger/
	cp ${BCS_SERVICES_PATH}/bcs-cluster-manager/api/clustermanager/clustermanager.swagger.json ${PACKAGEPATH}/bcs-services/bcs-cluster-manager/swagger/swagger-ui/clustermanager.swagger.json
	cd ${BCS_SERVICES_PATH}/bcs-cluster-manager && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-cluster-manager/bcs-cluster-manager ./main.go

alert-manager:pre
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-alert-manager/swagger
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-alert-manager/*  ${PACKAGEPATH}/bcs-services/bcs-alert-manager
	cp -R ${BCS_SERVICES_PATH}/bcs-alert-manager/pkg/third_party/swagger-ui ${PACKAGEPATH}/bcs-services/bcs-alert-manager/swagger/swagger-ui
	cp ${BCS_SERVICES_PATH}/bcs-alert-manager/pkg/proto/alertmanager/alertmanager.swagger.json ${PACKAGEPATH}/bcs-services/bcs-alert-manager/swagger/alertmanager.swagger.json
	cd ${BCS_SERVICES_PATH}/bcs-alert-manager/ && go mod tidy -compat=1.17 && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-alert-manager/bcs-alert-manager ./main.go

project-manager:pre
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-project-manager/swagger
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-project-manager/* ${PACKAGEPATH}/bcs-services/bcs-project-manager
	cp -R ${BCS_SERVICES_PATH}/bcs-project-manager/third_party/swagger-ui ${PACKAGEPATH}/bcs-services/bcs-project-manager/swagger/swagger-ui
	cp ${BCS_SERVICES_PATH}/bcs-project-manager/proto/bcsproject/bcsproject.swagger.json ${PACKAGEPATH}/bcs-services/bcs-project-manager/swagger/bcsproject.swagger.json
	cd ${BCS_SERVICES_PATH}/bcs-project-manager && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-project-manager/bcs-project-manager ./main.go
	cd ${BCS_SERVICES_PATH}/bcs-project-manager && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-project-manager/bcs-project-migration ./script/migrations/project/migrate.go
	cd ${BCS_SERVICES_PATH}/bcs-project-manager && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-project-manager/bcs-variable-migration ./script/migrations/variable/migrate.go

CR_LDFLAG_EXT=" -X github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/version.Version=${VERSION} \
 -X github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/version.GitCommit=${GITHASH} \
 -X github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/version.BuildTime=${BUILDTIME}"

cluster-resources:pre
	mkdir -p ${PACKAGEPATH}/bcs-services/cluster-resources
	cp -R ${BCS_CONF_SERVICES_PATH}/cluster-resources/* ${PACKAGEPATH}/bcs-services/cluster-resources
	# etc config files
	mkdir -p ${PACKAGEPATH}/bcs-services/cluster-resources/etc
	cp -R ${BCS_SERVICES_PATH}/cluster-resources/etc/ ${PACKAGEPATH}/bcs-services/cluster-resources/etc/
	# example files
	mkdir -p ${PACKAGEPATH}/bcs-services/cluster-resources/example/
	cp -R ${BCS_SERVICES_PATH}/cluster-resources/pkg/resource/example/config/ ${PACKAGEPATH}/bcs-services/cluster-resources/example/config/
	cp -R ${BCS_SERVICES_PATH}/cluster-resources/pkg/resource/example/manifest/ ${PACKAGEPATH}/bcs-services/cluster-resources/example/manifest/
	cp -R ${BCS_SERVICES_PATH}/cluster-resources/pkg/resource/example/reference/ ${PACKAGEPATH}/bcs-services/cluster-resources/example/reference/
	# form tmpl & schema files
	cp -R ${BCS_SERVICES_PATH}/cluster-resources/pkg/resource/form/tmpl/ ${PACKAGEPATH}/bcs-services/cluster-resources/tmpl/
	# i18n files
	cp ${BCS_SERVICES_PATH}/cluster-resources/pkg/i18n/locale/lc_msgs.yaml ${PACKAGEPATH}/bcs-services/cluster-resources/lc_msgs.yaml
	# go build
	cd ${BCS_SERVICES_PATH}/cluster-resources && go mod tidy -compat=1.17 && CGO_ENABLED=0 go build ${LDFLAG}${CR_LDFLAG_EXT} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/cluster-resources/bcs-cluster-resources *.go

# end of bcs-service section

apiserver-proxy:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-apiserver-proxy
	cp -R ${BCS_CONF_COMPONENT_PATH}/bcs-apiserver-proxy/* ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-apiserver-proxy
	cd ${BCS_COMPONENT_PATH}/bcs-apiserver-proxy && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-apiserver-proxy/bcs-apiserver-proxy ./main.go
	cd ${BCS_COMPONENT_PATH}/bcs-apiserver-proxy/ipvs_tools && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-apiserver-proxy/apiserver-proxy-tools .

apiserver-proxy-tools:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-apiserver-proxy
	cd ${BCS_COMPONENT_PATH}/bcs-apiserver-proxy/ipvs_tools && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-apiserver-proxy/apiserver-proxy-tools .


data-manager:pre
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-data-manager
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-data-manager ${PACKAGEPATH}/bcs-services
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-data-manager/swagger
	cp -R ${BCS_SERVICES_PATH}/bcs-data-manager/third_party/swagger-ui/* ${PACKAGEPATH}/bcs-services/bcs-data-manager/swagger/
	cp ${BCS_SERVICES_PATH}/bcs-data-manager/proto/bcs-data-manager/bcs-data-manager.swagger.json  ${PACKAGEPATH}/bcs-services/bcs-data-manager/swagger/bcs-data-manager.swagger.json
	cd bcs-services/bcs-data-manager/ && go mod tidy -go=1.16 && go mod tidy -go=1.17 && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-data-manager/bcs-data-manager ./main.go

helm-manager:pre
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-helm-manager
	cp -R ${BCS_SERVICES_PATH}/bcs-helm-manager/images/bcs-helm-manager/* ${PACKAGEPATH}/bcs-services/bcs-helm-manager/
	# swagger
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-helm-manager/swagger
	cp -R ${BCS_SERVICES_PATH}/bcs-helm-manager/third_party/swagger-ui ${PACKAGEPATH}/bcs-services/bcs-helm-manager/swagger/
	cp ${BCS_SERVICES_PATH}/bcs-helm-manager/proto/bcs-helm-manager/bcs-helm-manager.swagger.json ${PACKAGEPATH}/bcs-services/bcs-helm-manager/swagger/swagger-ui/bcs-helm-manager.swagger.json
	# i18n files
	cp ${BCS_SERVICES_PATH}/bcs-helm-manager/internal/i18n/locale/lc_msgs.yaml ${PACKAGEPATH}/bcs-services/bcs-helm-manager/lc_msgs.yaml
	# build
	cd ${BCS_SERVICES_PATH}/bcs-helm-manager && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-helm-manager/bcs-helm-manager ./main.go
	cd ${BCS_SERVICES_PATH}/bcs-helm-manager && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-helm-manager/bcs-helm-manager-migrator ./cmd/bcs-helm-manager-migrator/main.go


nodegroup-manager:pre
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-nodegroup-manager
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-nodegroup-manager ${PACKAGEPATH}/bcs-services
	cd bcs-services/bcs-nodegroup-manager/ && go mod tidy -go=1.16 && go mod tidy -go=1.17 && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-nodegroup-manager/bcs-nodegroup-manager ./main.go

gitops-proxy:
	mkdir -p ${SCENARIOSPACKAGE}/bcs-gitops-proxy
	cd bcs-scenarios/bcs-gitops-manager && make proxy && cd -

gitops-manager:
	mkdir -p ${SCENARIOSPACKAGE}/bcs-gitops-manager
	cd bcs-scenarios/bcs-gitops-manager && make manager && cd -

gitops-webhook:
	mkdir -p ${SCENARIOSPACKAGE}/bcs-gitops-webhook
	cd bcs-scenarios/bcs-gitops-manager && make webhook && cd -

gitops-vaultplugin-server:
	mkdir -p ${SCENARIOSPACKAGE}/bcs-gitops-vaultplugin-server
	cd bcs-scenarios/bcs-gitops-manager && make vaultplugin && cd -

test: test-bcs-runtime

test-bcs-runtime: test-bcs-k8s

test-bcs-k8s: test-bcs-service

test-bcs-service: test-user-manager

test-user-manager:
	@./scripts/test.sh ${BCS_SERVICES_PATH}/bcs-user-manager

gamedeployment:
	make gamedeployment -f bcs-scenarios/kourse/Makefile

gamestatefulset:
	make gamestatefulset -f bcs-scenarios/kourse/Makefile

hook-operator:
	make hook-operator -f bcs-scenarios/kourse/Makefile
