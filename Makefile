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

LDFLAG=-ldflags "-X bk-bcs/bcs-common/common/static.ZookeeperClientUser=${bcs_zk_client_user} \
 -X bk-bcs/bcs-common/common/static.ZookeeperClientPwd=${bcs_zk_client_pwd} \
 -X bk-bcs/bcs-common/common/static.EncryptionKey=${bcs_encryption_key} \
 -X bk-bcs/bcs-common/common/static.ServerCertPwd=${bcs_server_cert_pwd} \
 -X bk-bcs/bcs-common/common/static.ClientCertPwd=${bcs_client_cert_pwd} \
 -X bk-bcs/bcs-common/common/static.LicenseServerClientCertPwd=${bcs_license_server_client_cert_pwd} \
 -X bk-bcs/bcs-common/common/static.BcsDefaultUser=${bcs_registry_default_user} \
 -X bk-bcs/bcs-common/common/static.BcsDefaultPasswd=${bcs_registry_default_pwd} \
 -X bk-bcs/bcs-common/common/version.BcsVersion=${VERSION} \
 -X bk-bcs/bcs-common/common/version.BcsBuildTime=${BUILDTIME} \
 -X bk-bcs/bcs-common/common/version.BcsGitHash=${GITHASH} \
 -X bk-bcs/bcs-common/common/version.BcsTag=${GITTAG} \
 -X bk-bcs/bcs-common/common/version.BcsEdition=${bcs_edition}"

# build path config
PACKAGEPATH=./build/bcs.${VERSION}/
EXPORTPATH=./build/api_export

# options
default:api dns health client storage check executor driver mesos_watch scheduler loadbalance metricservice metriccollector exporter k8s_watch kube_agent api_export netservice sd_prometheus process_executor process_daemon bmsf-mesos-adapter hpacontroller logbeat-sidecar
specific:api dns health client storage check executor driver mesos_watch scheduler loadbalance metricservice metriccollector exporter k8s_watch kube_agent api_export netservice hpacontroller bmsf-mesos-adapter logbeat-sidecar

# tag for different edition compiling
inner:
	$(MAKE) specific bcs_edition=inner_edition
ce:
	$(MAKE) specific bcs_edition=communication_edition
ee:
	$(MAKE) specific bcs_edition=enterprise_edition

clean:
	rm -rf ./build

pre:
	@echo "git tag: ${GITTAG}"
	mkdir -p ${PACKAGEPATH}
	mkdir -p ${EXPORTPATH}
	cp -R ./install/cmd/conf/* ${PACKAGEPATH}/
	if [ ! -d "./vendor/github.com/sirupsen" ]; then cd ./vendor/github.com && ln -sf Sirupsen sirupsen; fi
	if [ ! -d "./vendor/github.com/Sirupsen" ]; then cd ./vendor/github.com && ln -sf sirupsen Sirupsen; fi
	go fmt ./...
	cd ./scripts && chmod +x vet.sh && ./vet.sh

api:pre
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-api/bcs-api ./bcs-services/bcs-api/main.go

kube_agent:pre
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-kube-agent/bcs-kube-agent ./bcs-k8s/bcs-kube-agent/main.go
	cp ./bcs-k8s/bcs-kube-agent/Dockerfile_new ${PACKAGEPATH}/bcs-kube-agent/Dockerfile

client:pre
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-client/bcs-client ./bcs-services/bcs-client/cmd/main.go

dns:pre
	cp bcs-services/bcs-dns/plugin.cfg vendor/github.com/coredns/coredns/
	cd vendor/github.com/coredns/coredns && make gen && cd -
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-dns/bcs-dns bk-bcs/vendor/github.com/coredns/coredns

health:pre
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-health-master/bcs-health-master ./bcs-services/bcs-health/master/main.go
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-health-slave/bcs-health-slave ./bcs-services/bcs-health/slave/main.go

metricservice:pre
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-metricservice/bcs-metricservice ./bcs-services/bcs-metricservice/main.go

metriccollector:pre
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-metriccollector/bcs-metriccollector ./bcs-services/bcs-metriccollector/main.go
	cp ./bcs-services/bcs-metriccollector/conf/config_file_docker.json ${PACKAGEPATH}/bcs-metriccollector/config_file_docker.json
	cp ./bcs-services/bcs-metriccollector/conf/start_docker.sh ${PACKAGEPATH}/bcs-metriccollector/start_docker.sh
	cp ./bcs-services/bcs-metriccollector/conf/Dockerfile ${PACKAGEPATH}/bcs-metriccollector/Dockerfile

exporter:pre
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-exporter/bcs-exporter ./bcs-services/bcs-exporter/main.go
	go build ${LDFLAG} -buildmode=plugin -o ${PACKAGEPATH}/bcs-exporter/default_exporter.so ./bcs-services/bcs-exporter/pkg/output/plugins/default_exporter/default_exporter.go
	go build ${LDFLAG} -buildmode=plugin -o ${PACKAGEPATH}/bcs-exporter/bkdata_exporter.so ./bcs-services/bcs-exporter/pkg/output/plugins/bkdata_exporter/

storage:pre
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-storage/bcs-storage ./bcs-services/bcs-storage/storage.go

loadbalance:pre
	@mkdir -p ${PACKAGEPATH}/bcs-loadbalance
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-loadbalance/bcs-loadbalance ./bcs-services/bcs-loadbalance/main.go

check:pre
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-check/bcs-check ./bcs-mesos/bcs-check/bcs-check.go

executor:pre
	@mkdir -p ${PACKAGEPATH}/bcs-container-executor
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-container-executor/bcs-container-executor ./bcs-mesos/bcs-container-executor/main.go

process_executor:pre
	mkdir -p ${PACKAGEPATH}/bcs-process-executor
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-process-executor/bcs-process-executor ./bcs-mesos/bcs-process-executor/main.go

process_daemon:pre
	@mkdir -p ${PACKAGEPATH}/bcs-process-daemon
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-process-daemon/bcs-process-daemon ./bcs-mesos/bcs-process-daemon/main.go

netservice:pre
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-netservice/bcs-netservice ./bcs-services/bcs-netservice/main.go
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-ipam/bcs-ipam ./bcs-services/bcs-netservice/bcs-ipam/main.go

driver:pre
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-mesos-driver/bcs-mesos-driver ./bcs-mesos/bcs-mesos-driver/main.go

mesos_watch:pre
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-mesos-watch/bcs-mesos-watch ./bcs-mesos/bcs-mesos-watch/main.go

scheduler:pre
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-scheduler/bcs-scheduler ./bcs-mesos/bcs-scheduler
	go build -buildmode=plugin -o ${PACKAGEPATH}/bcs-scheduler/plugin/bin/ip-resources/ip-resources.so ./bcs-mesos/bcs-scheduler/src/plugin/bin/ip-resources/ipResource.go

hpacontroller:pre
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-hpacontroller/bcs-hpacontroller ./bcs-mesos/bcs-hpacontroller

sd_prometheus:pre
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-service-prometheus/bcs-service-prometheus ./bcs-services/bcs-sd-prometheus/main.go

logbeat-sidecar:pre
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-logbeat-sidecar/bcs-logbeat-sidecar ./bcs-services/bcs-logbeat-sidecar/main.go

k8s_watch:pre
	mkdir -p ${PACKAGEPATH}/bcs-k8s-watch
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-k8s-watch/bcs-k8s-watch ./bcs-k8s/bcs-k8s-watch/main.go
	cp ./bcs-k8s/bcs-k8s-watch/Dockerfile_new ${PACKAGEPATH}/bcs-k8s-watch/Dockerfile

api_export:pre
	mkdir -p ${EXPORTPATH}
	cp ./bcs-common/common/types/meta.go ${EXPORTPATH}
	cp ./bcs-common/common/types/status.go ${EXPORTPATH}
	cp ./bcs-common/common/types/secret.go ${EXPORTPATH}
	cp ./bcs-common/common/types/configmap.go ${EXPORTPATH}

consoleproxy:pre
	go build ${LDFLAG} -o ${PACKAGEPATH}/bcs-consoleproxy/bcs-consoleproxy ./bcs-mesos/bcs-consoleproxy/main.go

bmsf-mesos-adapter:pre
	go build ${LDFLAG} -o ${PACKAGEPATH}/bmsf-mesos-adapter/bmsf-mesos-adapter ./bmsf-mesh/bmsf-mesos-adapter/main.go

network:
	go build ${LDFLAG} -o ${PACKAGEPATH}/qcloud-eip/qcloud-eip ./bcs-services/bcs-network/qcloud-eip/main.go
