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
BCS_CONF_UI_PATH=${WORKSPACE}/install/conf/bcs-ui
BCS_CONF_COMPONENT_PATH=${WORKSPACE}/install/conf/bcs-runtime/bcs-k8s/bcs-component
BCS_CONF_NETWORK_PATH=${WORKSPACE}/install/conf/bcs-runtime/bcs-k8s/bcs-network
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

export GATEWAYERRORLDFLAG=-ldflags "-X github.com/Tencent/bk-bcs/bcs-common/common/static.ZookeeperClientUser=${bcs_zk_client_user} \
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
 -X github.com/Tencent/bk-bcs/bcs-common/common/version.BcsEdition=${bcs_edition} \
 -X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=warn"

# build path config
export PACKAGEPATH=./build/bcs.${VERSION}
export SCENARIOSPACKAGE=${WORKSPACE}/${PACKAGEPATH}/bcs-scenarios

# bscp 应用自定义
export BSCP_LDFLAG=-ldflags "-X github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/version.BUILDTIME=${BUILDTIME} \
	-X github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/version.GITHASH=${GITHASH}"

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

bcs-runtime: bcs-k8s

bcs-k8s: bcs-component bcs-network

bcs-component:kube-sche apiserver-proxy \
	webhook-server \
	general-pod-autoscaler cluster-autoscaler \
	netservice-controller external-privilege \
	image-loader

bcs-network:ingress-controller

bcs-services:bkcmdb-synchronizer gateway \
	storage user-manager cluster-manager cluster-reporter nodeagent tools k8s-watch kube-agent data-manager \
	helm-manager project-manager nodegroup-manager federation-manager powertrading

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

netpack:
	cd ./build/bcs.${VERSION}/bcs-runtime/bcs-k8s/bcs-network && find . -type f ! -name MD5 | xargs -L1 md5sum > MD5

pre:
	@echo "git tag: ${GITTAG}"
	mkdir -p ${PACKAGEPATH}
	go mod tidy
	go fmt ./...
	cd ./scripts && chmod +x vet.sh && ./vet.sh

tongsuo:
	cd ./scripts && chmod +x tongsuo.sh && ./tongsuo.sh

gateway:pre
	mkdir -p ${PACKAGEPATH}/bcs-services
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-gateway-discovery ${PACKAGEPATH}/bcs-services
	cp -R ./bcs-services/bcs-gateway-discovery/plugins/apisix ${PACKAGEPATH}/bcs-services/bcs-gateway-discovery/
	cd bcs-services/bcs-gateway-discovery && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-gateway-discovery/bcs-gateway-discovery ./main.go

micro-gateway:pre
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-micro-gateway
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-gateway-discovery/* ${PACKAGEPATH}/bcs-services/bcs-micro-gateway/
	cp -R ./bcs-services/bcs-gateway-discovery/plugins/apisix ${PACKAGEPATH}/bcs-services/bcs-micro-gateway/

api-gateway-syncing:
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-api-gateway-syncing
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-api-gateway-syncing/* ${PACKAGEPATH}/bcs-services/bcs-api-gateway-syncing/

storage:pre
	mkdir -p ${PACKAGEPATH}/bcs-services
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-storage ${PACKAGEPATH}/bcs-services
	cd ./bcs-services/bcs-storage && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-storage/bcs-storage ./storage.go

kube-sche:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cp -R ${BCS_CONF_COMPONENT_PATH}/bcs-k8s-custom-scheduler ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cd ${BCS_COMPONENT_PATH}/bcs-k8s-custom-scheduler && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-k8s-custom-scheduler/bcs-k8s-custom-scheduler ./main.go

multi-ns-proxy:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cp -R ${BCS_CONF_COMPONENT_PATH}/bcs-multi-ns-proxy ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cd ./bcs-runtime/bcs-k8s/bcs-component/bcs-multi-ns-proxy  && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-multi-ns-proxy/bcs-multi-ns-proxy ./main.go

log-manager:pre
	mkdir -p ${PACKAGEPATH}/bcs-services
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-log-manager ${PACKAGEPATH}/bcs-services
	cd ./bcs-services/bcs-log-manager && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-log-manager/bcs-log-manager ./main.go

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

tools:pre tongsuo
	mkdir -p ${PACKAGEPATH}/bcs-services
	cd ${BCS_INSTALL_PATH}/cryptool && go mod tidy && $(CGO_BUILD_FLAGS) go build ${LDFLAG} -o  ${WORKSPACE}/${PACKAGEPATH}/bcs-services/cryptools main.go

ui:pre
	mkdir -p ${PACKAGEPATH}/bcs-ui
	cp -R ${BCS_CONF_UI_PATH} ${PACKAGEPATH}
	cd ${BCS_UI_PATH} && ls -la && cd frontend && npm install && npm run build && cd ../ && go mod tidy && CGO_ENABLED=0 go build -trimpath ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-ui/bcs-ui ./cmd/bcs-ui

user-manager:pre tongsuo
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-user-manager
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-user-manager ${PACKAGEPATH}/bcs-services
	cd bcs-services/bcs-user-manager/ && go mod tidy && $(CGO_BUILD_FLAGS) go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-user-manager/bcs-user-manager ./main.go

webconsole:pre
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-webconsole
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-webconsole ${PACKAGEPATH}/bcs-services
	cd bcs-services/bcs-webconsole/ && go mod tidy && CGO_ENABLED=0 go build -trimpath ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-webconsole/bcs-webconsole ./main.go

monitor:pre
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-monitor
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-monitor ${PACKAGEPATH}/bcs-services
	cd bcs-services/bcs-monitor/ && go mod tidy && CGO_ENABLED=0 go build -trimpath ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-monitor/bcs-monitor ./cmd/bcs-monitor

bscp:pre
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-bscp
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-bscp ${PACKAGEPATH}/bcs-services
	cd bcs-services/bcs-bscp && cd ui && npm install --legacy-peer-deps && npm run build && cd ../ && go mod tidy -compat=1.20 && CGO_ENABLED=0 go build -trimpath ${BSCP_LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-bscp/hyper/bscp-ui ./cmd/ui
	cd bcs-services/bcs-bscp && go mod tidy -compat=1.20 && CGO_ENABLED=0 go build -trimpath ${BSCP_LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-bscp/hyper/bk-bscp-apiserver ./cmd/api-server
	cd bcs-services/bcs-bscp && go mod tidy -compat=1.20 && CGO_ENABLED=0 go build -trimpath ${BSCP_LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-bscp/hyper/bk-bscp-authserver ./cmd/auth-server
	cd bcs-services/bcs-bscp && go mod tidy -compat=1.20 && CGO_ENABLED=0 go build -trimpath ${BSCP_LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-bscp/hyper/bk-bscp-configserver ./cmd/config-server
	cd bcs-services/bcs-bscp && go mod tidy -compat=1.20 && CGO_ENABLED=0 go build -trimpath ${BSCP_LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-bscp/hyper/bk-bscp-dataservice ./cmd/data-service
	cd bcs-services/bcs-bscp && go mod tidy -compat=1.20 && CGO_ENABLED=0 go build -trimpath ${BSCP_LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-bscp/hyper/bk-bscp-feedserver ./cmd/feed-server
	cd bcs-services/bcs-bscp && go mod tidy -compat=1.20 && CGO_ENABLED=0 go build -trimpath ${BSCP_LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-bscp/hyper/bk-bscp-cacheservice ./cmd/cache-service
	cd bcs-services/bcs-bscp && go mod tidy -compat=1.20 && CGO_ENABLED=0 go build -trimpath ${BSCP_LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-bscp/hyper/bk-bscp-vaultserver ./cmd/vault-server
	cd bcs-services/bcs-bscp/cmd/vault-server/vault && go mod tidy -compat=1.20 && CGO_ENABLED=0 go build -trimpath ${BSCP_LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-bscp/hyper/vault ./main.go
	cd bcs-services/bcs-bscp/cmd/vault-server/vault-sidecar && go mod tidy -compat=1.20 && CGO_ENABLED=0 go build -trimpath ${BSCP_LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-bscp/hyper/vault-sidecar *.go
	cd bcs-services/bcs-bscp/cmd/vault-server/vault-plugins && go mod tidy -compat=1.20 && CGO_ENABLED=0 go build -trimpath ${BSCP_LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-bscp/hyper/bk-bscp-secret *.go
	# alias docker image name to bk-bscp-hyper
	touch ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-bscp/bk-bscp-hyper && chmod a+x ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-bscp/bk-bscp-hyper && ls -la ${PACKAGEPATH}/bcs-services/bcs-bscp/hyper

busybox:pre
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-busybox
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-busybox ${PACKAGEPATH}/bcs-services
	# alias docker image name to bcs-busybox
	touch ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-busybox/bcs-busybox && chmod a+x ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-busybox/bcs-busybox

k8s-watch:pre
	mkdir -p ${PACKAGEPATH}/bcs-services
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-k8s-watch ${PACKAGEPATH}/bcs-services
	cd ./bcs-services/bcs-k8s-watch && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-k8s-watch/bcs-k8s-watch  ./main.go

kube-agent:pre
	mkdir -p ${PACKAGEPATH}/bcs-services
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-kube-agent ${PACKAGEPATH}/bcs-services
	cd ./bcs-services/bcs-kube-agent && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-kube-agent/bcs-kube-agent  ./main.go

general-pod-autoscaler:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cp -R ${BCS_CONF_COMPONENT_PATH}/bcs-general-pod-autoscaler ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cd ${BCS_COMPONENT_PATH}/bcs-general-pod-autoscaler && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-general-pod-autoscaler/bcs-general-pod-autoscaler ./cmd/gpa/main.go

image-loader:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cp -R ${BCS_CONF_COMPONENT_PATH}/bcs-image-loader ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cd ${BCS_COMPONENT_PATH}/bcs-image-loader && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-image-loader/bcs-image-loader ./cmd/main.go

cluster-autoscaler:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/
	cp -R ${BCS_CONF_COMPONENT_PATH}/bcs-cluster-autoscaler ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cd ${BCS_COMPONENT_PATH}/bcs-cluster-autoscaler/bcs-cluster-autoscaler-1.16 && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/hyper/bcs-cluster-autoscaler-1.16 ./main.go
	cd ${BCS_COMPONENT_PATH}/bcs-cluster-autoscaler/bcs-cluster-autoscaler-1.22 && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/hyper/bcs-cluster-autoscaler-1.22 ./main.go
	touch ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/bcs-cluster-autoscaler && chmod a+x ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/bcs-cluster-autoscaler && ls -la ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/hyper

netservice-controller:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/
	cp -R ${BCS_CONF_COMPONENT_PATH}/bcs-netservice-controller ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cd ${BCS_COMPONENT_PATH}/bcs-netservice-controller && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-netservice-controller/bcs-netservice-controller ./main.go
	cd ${BCS_COMPONENT_PATH}/bcs-netservice-controller && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-netservice-controller/bcs-netservice-ipam ./ipam/main.go
	cd ${BCS_COMPONENT_PATH}/bcs-netservice-controller/cni && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-netservice-controller/bcs-underlay-cni ./cni.go

external-privilege:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/
	cp -R ${BCS_CONF_COMPONENT_PATH}/bcs-external-privilege ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component
	cd ${BCS_COMPONENT_PATH}/bcs-external-privilege && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-external-privilege/bcs-external-privilege ./main.go

bkcmdb-synchronizer:
	mkdir -p ${PACKAGEPATH}/bcs-services
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-bkcmdb-synchronizer ${PACKAGEPATH}/bcs-services
	cd bcs-services/bcs-bkcmdb-synchronizer && make synchronizer
	cp bcs-services/bcs-bkcmdb-synchronizer/bin/bcs-bkcmdb-synchronizer ${PACKAGEPATH}/bcs-services/bcs-bkcmdb-synchronizer/

# network plugins section

ingress-controller:pre
	cd ${BCS_NETWORK_PATH} && go mod tidy && make ingress-controller

#end of network plugins

# bcs-service section
cluster-manager:pre tongsuo
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-cluster-manager
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-cluster-manager/* ${PACKAGEPATH}/bcs-services/bcs-cluster-manager/
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-cluster-manager/swagger
	cp -R ${BCS_SERVICES_PATH}/bcs-cluster-manager/third_party/swagger-ui ${PACKAGEPATH}/bcs-services/bcs-cluster-manager/swagger/
	cp ${BCS_SERVICES_PATH}/bcs-cluster-manager/api/clustermanager/clustermanager.swagger.json ${PACKAGEPATH}/bcs-services/bcs-cluster-manager/swagger/swagger-ui/clustermanager.swagger.json
	cd ${BCS_SERVICES_PATH}/bcs-cluster-manager && go mod tidy && ${CGO_BUILD_FLAGS} go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-cluster-manager/bcs-cluster-manager ./main.go

cluster-reporter:
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-cluster-reporter
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-cluster-reporter/* ${PACKAGEPATH}/bcs-services/bcs-cluster-reporter/
	cd ${BCS_SERVICES_PATH}/bcs-cluster-reporter/cmd/reporter && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-cluster-reporter/bcs-cluster-reporter ./main.go

nodeagent:
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-nodeagent
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-nodeagent/* ${PACKAGEPATH}/bcs-services/bcs-nodeagent/
	cd ${BCS_SERVICES_PATH}/bcs-cluster-reporter/cmd/nodeagent && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-nodeagent/bcs-nodeagent ./main.go

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
	cd ${BCS_SERVICES_PATH}/cluster-resources && go mod tidy && CGO_ENABLED=0 go build ${LDFLAG}${CR_LDFLAG_EXT} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/cluster-resources/bcs-cluster-resources *.go

# end of bcs-service section

apiserver-proxy:pre
	mkdir -p ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-apiserver-proxy
	cp -R ${BCS_CONF_COMPONENT_PATH}/bcs-apiserver-proxy/* ${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-apiserver-proxy
	cd ${BCS_COMPONENT_PATH}/bcs-apiserver-proxy/ipvs_tools && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-apiserver-proxy/bcs-apiserver-proxy-tools .
	cd ${BCS_COMPONENT_PATH}/bcs-apiserver-proxy && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-runtime/bcs-k8s/bcs-component/bcs-apiserver-proxy/bcs-apiserver-proxy ./main.go

data-manager:pre
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-data-manager
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-data-manager ${PACKAGEPATH}/bcs-services
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-data-manager/swagger
	cp -R ${BCS_SERVICES_PATH}/bcs-data-manager/third_party/swagger-ui/* ${PACKAGEPATH}/bcs-services/bcs-data-manager/swagger/
	cp ${BCS_SERVICES_PATH}/bcs-data-manager/proto/bcs-data-manager/bcs-data-manager.swagger.json  ${PACKAGEPATH}/bcs-services/bcs-data-manager/swagger/bcs-data-manager.swagger.json
	cd bcs-services/bcs-data-manager/ && go mod tidy && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-data-manager/bcs-data-manager ./main.go

federation-manager:pre
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-federation-manager
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-federation-manager ${PACKAGEPATH}/bcs-services
	cd bcs-services/bcs-federation-manager/ && go mod tidy && go build ${GATEWAYERRORLDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-federation-manager/bcs-federation-manager ./main.go

helm-manager:pre tongsuo
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-helm-manager
	cp -R ${BCS_SERVICES_PATH}/bcs-helm-manager/images/bcs-helm-manager/* ${PACKAGEPATH}/bcs-services/bcs-helm-manager/
	# swagger
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-helm-manager/swagger
	cp -R ${BCS_SERVICES_PATH}/bcs-helm-manager/third_party/swagger-ui ${PACKAGEPATH}/bcs-services/bcs-helm-manager/swagger/
	cp ${BCS_SERVICES_PATH}/bcs-helm-manager/proto/bcs-helm-manager/bcs-helm-manager.swagger.json ${PACKAGEPATH}/bcs-services/bcs-helm-manager/swagger/swagger-ui/bcs-helm-manager.swagger.json
	# i18n files
	cp ${BCS_SERVICES_PATH}/bcs-helm-manager/internal/i18n/locale/lc_msgs.yaml ${PACKAGEPATH}/bcs-services/bcs-helm-manager/lc_msgs.yaml
	# build
	cd ${BCS_SERVICES_PATH}/bcs-helm-manager && go mod tidy && $(CGO_BUILD_FLAGS) go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-helm-manager/bcs-helm-manager ./main.go
	cd ${BCS_SERVICES_PATH}/bcs-helm-manager && go mod tidy && $(CGO_BUILD_FLAGS) go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-helm-manager/bcs-helm-manager-migrator ./cmd/bcs-helm-manager-migrator/main.go


nodegroup-manager:pre
	mkdir -p ${PACKAGEPATH}/bcs-services/bcs-nodegroup-manager
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-nodegroup-manager ${PACKAGEPATH}/bcs-services
	cd bcs-services/bcs-nodegroup-manager/ && go mod tidy && go build ${LDFLAG} -o ${WORKSPACE}/${PACKAGEPATH}/bcs-services/bcs-nodegroup-manager/bcs-nodegroup-manager ./main.go

gitops-proxy:
	mkdir -p ${SCENARIOSPACKAGE}/bcs-gitops-proxy
	cd bcs-scenarios/bcs-gitops-manager && make proxy && cd -

gitops-manager:
	mkdir -p ${SCENARIOSPACKAGE}/bcs-gitops-manager
	cd bcs-scenarios/bcs-gitops-manager && make manager && cd -

gitops-analysis:
	mkdir -p ${SCENARIOSPACKAGE}/bcs-gitops-analysis
	cd bcs-scenarios/bcs-gitops-analysis && make analysis && cd -

gitops-webhook:
	mkdir -p ${SCENARIOSPACKAGE}/bcs-gitops-webhook
	cd bcs-scenarios/bcs-gitops-manager && make webhook && cd -

gitops-vaultplugin-server:
	mkdir -p ${SCENARIOSPACKAGE}/bcs-gitops-vaultplugin-server
	cd bcs-scenarios/bcs-gitops-vaultplugin-server && make vaultplugin && cd -

gitops-gitgenerator-webhook:
	mkdir -p ${SCENARIOSPACKAGE}/bcs-gitops-gitgenerator-webhook
	cd bcs-scenarios/bcs-gitops-manager && make gitgenerator-webhook && cd -

gitops-workflow:
	mkdir -p ${SCENARIOSPACKAGE}/bcs-gitops-workflow
	cd bcs-scenarios/bcs-gitops-workflow && make gitops-workflow && cd -

terraform-controller:
	mkdir -p ${SCENARIOSPACKAGE}/bcs-terraform-controller
	cd bcs-scenarios/bcs-terraform-controller && make terraform-controller && cd -

terraform-bkprovider:
	mkdir -p ${SCENARIOSPACKAGE}/bcs-terraform-bkprovider
	cd bcs-scenarios/bcs-terraform-bkprovider && make build && cd -

monitor-controller:
	mkdir -p ${SCENARIOSPACKAGE}/bcs-monitor-controller
	cd bcs-scenarios/bcs-monitor-controller && make manager && cd -

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

powertrading:
	mkdir -p ${PACKAGEPATH}/bcs-scenarios/bcs-powertrading
	cp -R ${BCS_CONF_SERVICES_PATH}/bcs-powertrading ${PACKAGEPATH}/bcs-scenarios
	cd bcs-scenarios/bcs-powertrading/ && go mod tidy && make build && cd -
