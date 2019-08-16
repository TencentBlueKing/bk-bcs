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
BINARYPATH=./build/bcs.${VERSION}/bin
CONFPATH=./build/bcs.${VERSION}/conf
COMMONPATH=./build/bcs.${VERSION}/common
EXPORTPATH=./build/api_export

# options
default:api dns health client storage check executor driver mesos_watch scheduler loadbalance metricservice metriccollector exporter k8s_watch kube_agent api_export netservice sd_prometheus
specific:api dns health client storage check executor driver mesos_watch scheduler loadbalance metricservice metriccollector exporter k8s_watch kube_agent api_export netservice hpacontroller

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
	mkdir -p ${BINARYPATH}
	mkdir -p ${CONFPATH}
	mkdir -p ${EXPORTPATH}
	cp -R ./install/cmd/conf/* ${CONFPATH}/
	if [ ! -d "./vendor/github.com/sirupsen" ]; then cd ./vendor/github.com && ln -sf Sirupsen sirupsen; fi
	if [ ! -d "./vendor/github.com/Sirupsen" ]; then cd ./vendor/github.com && ln -sf sirupsen Sirupsen; fi
	go fmt ./...
	cd ./scripts && chmod +x vet.sh && ./vet.sh

api:pre
	go build ${LDFLAG} -o ${BINARYPATH}/bcs-api ./bcs-services/bcs-api/main.go

kube_agent:pre
	go build ${LDFLAG} -o ${BINARYPATH}/bcs-kube-agent/bcs-kube-agent ./bcs-k8s/bcs-kube-agent/main.go
	cp ./bcs-k8s/bcs-kube-agent/Dockerfile_new ${BINARYPATH}/bcs-kube-agent/Dockerfile

client:pre
	go build ${LDFLAG} -o ${BINARYPATH}/bcs-client ./bcs-services/bcs-client/cmd/main.go

dns:pre
	cp bcs-services/bcs-dns/plugin.cfg vendor/github.com/coredns/coredns/
	cd vendor/github.com/coredns/coredns && make gen && cd -
	go build ${LDFLAG} -o ${BINARYPATH}/bcs-dns bk-bcs/vendor/github.com/coredns/coredns

health:pre
	go build ${LDFLAG} -o ${BINARYPATH}/bcs-health-master ./bcs-services/bcs-health/master/main.go
	go build ${LDFLAG} -o ${BINARYPATH}/bcs-health-slave ./bcs-services/bcs-health/slave/main.go

metricservice:pre
	go build ${LDFLAG} -o ${BINARYPATH}/bcs-metricservice ./bcs-services/bcs-metricservice/main.go

metriccollector:pre
	go build ${LDFLAG} -o ${BINARYPATH}/bcs-metriccollector/bcs-metriccollector ./bcs-services/bcs-metriccollector/main.go
	cp ./bcs-services/bcs-metriccollector/conf/config_file_docker.json ${BINARYPATH}/bcs-metriccollector/config_file_docker.json
	cp ./bcs-services/bcs-metriccollector/conf/start_docker.sh ${BINARYPATH}/bcs-metriccollector/start_docker.sh
	cp ./bcs-services/bcs-metriccollector/conf/Dockerfile ${BINARYPATH}/bcs-metriccollector/Dockerfile

exporter:pre
	go build ${LDFLAG} -o ${BINARYPATH}/bcs-exporter ./bcs-services/bcs-exporter/main.go
	go build ${LDFLAG} -buildmode=plugin -o ${BINARYPATH}/default_exporter.so ./bcs-services/bcs-exporter/pkg/output/plugins/default_exporter/default_exporter.go
	go build ${LDFLAG} -buildmode=plugin -o ${BINARYPATH}/bkdata_exporter.so ./bcs-services/bcs-exporter/pkg/output/plugins/bkdata_exporter/

storage:pre
	go build ${LDFLAG} -o ${BINARYPATH}/bcs-storage ./bcs-services/bcs-storage/storage.go

loadbalance:pre
	go build ${LDFLAG} -o ${BINARYPATH}/bcs-loadbalance ./bcs-services/bcs-loadbalance/main.go

check:pre
	go build ${LDFLAG} -o ${BINARYPATH}/bcs-check ./bcs-mesos/bcs-check/bcs-check.go

executor:pre
	go build ${LDFLAG} -o ${BINARYPATH}/bcs-container-executor ./bcs-mesos/bcs-container-executor/main.go

processexecutor:pre
	go build ${LDFLAG} -o ${BINARYPATH}/bcs-process-executor ./bcs-mesos/bcs-process-executor/main.go

daemon:pre
	go build ${LDFLAG} -o ${BINARYPATH}/process.so -buildmode=c-shared ./bcs-mesos/bcs-process-daemon/app.go
	mv ${BINARYPATH}/process.so ${BINARYPATH}/libprocess.so
	gcc -o ${BINARYPATH}/bcs-daemon ./bcs-mesos/bcs-process-daemon/daemon.c -I${BINARYPATH} -L${BINARYPATH} -lprocess
	rm -f ${BINARYPATH}/process.h

netservice:pre
	go build ${LDFLAG} -o ${BINARYPATH}/bcs-netservice ./bcs-services/bcs-netservice/main.go
	go build ${LDFLAG} -o ${BINARYPATH}/bcs-ipam ./bcs-services/bcs-netservice/bcs-ipam/main.go

driver:pre
	go build ${LDFLAG} -o ${BINARYPATH}/bcs-mesos-driver ./bcs-mesos/bcs-mesos-driver/main.go

mesos_watch:pre
	go build ${LDFLAG} -o ${BINARYPATH}/bcs-mesos-watch ./bcs-mesos/bcs-mesos-watch/main.go

scheduler:pre
	go build ${LDFLAG} -o ${BINARYPATH}/bcs-scheduler ./bcs-mesos/bcs-scheduler
	go build -buildmode=plugin -o ${BINARYPATH}/ip-resources.so ./bcs-mesos/bcs-scheduler/src/plugin/bin/ip-resources/ipResource.go

hpacontroller:pre
	go build ${LDFLAG} -o ${BINARYPATH}/bcs-hpacontroller ./bcs-mesos/bcs-hpacontroller

sd_prometheus:pre
	go build ${LDFLAG} -o ${BINARYPATH}/bcs-sd-prometheus ./bcs-services/bcs-sd-prometheus/main.go

k8s_watch:pre
	mkdir -p ${BINARYPATH}/bcs-k8s-watch
	go build ${LDFLAG} -o ${BINARYPATH}/bcs-k8s-watch/bcs-k8s-watch ./bcs-k8s/bcs-k8s-watch/main.go
	cp ./bcs-k8s/bcs-k8s-watch/Dockerfile_new ${BINARYPATH}/bcs-k8s-watch/Dockerfile

api_export:pre
	mkdir -p ${EXPORTPATH}
	cp ./bcs-common/common/types/meta.go ${EXPORTPATH}
	cp ./bcs-common/common/types/status.go ${EXPORTPATH}
	cp ./bcs-common/common/types/secret.go ${EXPORTPATH}
	cp ./bcs-common/common/types/configmap.go ${EXPORTPATH}

consoleproxy:pre
	go build ${LDFLAG} -o ${BINARYPATH}/bcs-consoleproxy ./bcs-mesos/bcs-consoleproxy/main.go
