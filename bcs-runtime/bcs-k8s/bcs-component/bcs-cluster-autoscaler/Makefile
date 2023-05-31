all: build

GIT_COMMIT=$(shell git rev-parse "HEAD^{commit}")
VERSION=$(shell git describe --tags --abbrev=14 "${GIT_COMMIT}^{commit}" --always)
BUILD_TIME=$(shell TZ=Asia/Shanghai date +%FT%T%z)
WORKSPACE=$(shell pwd)

TAG?=dev
FLAGS=
LDFLAGS?=-X 'github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/cloudprovider/bcs.EncryptionKey=${bcs_encryption_key}'
ENVVAR=CGO_ENABLED=0 GO111MODULE=on
GOOS?=linux
REGISTRY?=
REPOSITORY?=bcs
DOCKER_NETWORK?=default
ifdef LDFLAGS
  LDFLAGS_FLAG=-ldflags "${LDFLAGS}"
else
  LDFLAGS_FLAG=
endif


build: clean
	cd ${WORKSPACE}/bcs-cluster-autoscaler-1.16 && go mod tidy && go build ${LDFLAGS_FLAG} ${TAGS_FLAG} -o ${WORKSPACE}/bcs-ca-1.16
	cd ${WORKSPACE}/bcs-cluster-autoscaler-1.22 && go mod tidy && go build ${LDFLAGS_FLAG} ${TAGS_FLAG} -o ${WORKSPACE}/bcs-ca-1.22

clean:
	rm -f ${WORKSPACE}/bcs-ca-1.16 ${WORKSPACE}/bcs-ca-1.22

make-image:
ifdef BASEIMAGE
	docker build --pull --build-arg BASEIMAGE=${BASEIMAGE} \
	    -t ${REGISTRY}/${REPOSITORY}/bcs-cluster-autoscaler${PROVIDER}:${TAG} .
else
	docker build --pull -t ${REGISTRY}/${REPOSITORY}/bcs-cluster-autoscaler${PROVIDER}:${TAG} .
endif
