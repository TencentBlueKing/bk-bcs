all: build

GIT_COMMIT=$(shell git rev-parse "HEAD^{commit}")
VERSION=$(shell git describe --tags --abbrev=14 "${GIT_COMMIT}^{commit}" --always)
BUILD_TIME=$(shell TZ=Asia/Shanghai date +%FT%T%z)

TAG?=dev
FLAGS=
LDFLAGS?=-X 'github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/cloudprovider/bcs.EncryptionKey=${bcs_encryption_key}'
ENVVAR=CGO_ENABLED=0 GO111MODULE=on
GOOS?=linux
REGISTRY?=
REPOSITORY?=bcs
DOCKER_NETWORK?=default
ifdef BUILD_TAGS
  TAGS_FLAG=--tags ${BUILD_TAGS}
  PROVIDER=-${BUILD_TAGS}
  FOR_PROVIDER=" for ${BUILD_TAGS}"
else
  TAGS_FLAG=
  PROVIDER=
  FOR_PROVIDER=
endif
ifdef LDFLAGS
  LDFLAGS_FLAG=-ldflags "${LDFLAGS}"
else
  LDFLAGS_FLAG=
endif
ifdef DOCKER_RM
  RM_FLAG=--rm
else
  RM_FLAG=
endif

build: clean deps
	$(ENVVAR) GOOS=$(GOOS) go build ${LDFLAGS_FLAG} ${TAGS_FLAG} ./...
	$(ENVVAR) GOOS=$(GOOS) go build -o bcs-cluster-autoscaler ${LDFLAGS_FLAG} ${TAGS_FLAG}

build-binary: clean deps
	$(ENVVAR) GOOS=$(GOOS) go build -o bcs-cluster-autoscaler ${LDFLAGS_FLAG} ${TAGS_FLAG}

test-unit: clean deps build
	GO111MODULE=on go test --test.short -race ./... ${TAGS_FLAG}

dev-release: build-binary execute-release
	@echo "Release ${TAG}${FOR_PROVIDER} completed"

make-image:
ifdef BASEIMAGE
	docker build --pull --build-arg BASEIMAGE=${BASEIMAGE} \
	    -t ${REGISTRY}/${REPOSITORY}/bcs-cluster-autoscaler${PROVIDER}:${TAG} .
else
	docker build --pull -t ${REGISTRY}/${REPOSITORY}/bcs-cluster-autoscaler${PROVIDER}:${TAG} .
endif

push-image:
	./push_image.sh ${REGISTRY}/${REPOSITORY}/bcs-cluster-autoscaler${PROVIDER}:${TAG}

execute-release: make-image push-image

clean:
	rm -f bcs-cluster-autoscaler

generate:
	go generate ./cloudprovider/aws

format:
	test -z "$$(find . -path ./vendor -prune -type f -o -name '*.go' -exec gofmt -s -d {} + | tee /dev/stderr)" || \
    test -z "$$(find . -path ./vendor -prune -type f -o -name '*.go' -exec gofmt -s -w {} + | tee /dev/stderr)"

docker-builder:
	docker build --network=${DOCKER_NETWORK} -t autoscaling-builder ./builder

build-in-docker: clean docker-builder
	docker run ${RM_FLAG} -v `pwd`:/gopath/src/github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/:Z autoscaling-builder:latest bash -c 'cd /gopath/src/github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler && BUILD_TAGS=${BUILD_TAGS} LDFLAGS="${LDFLAGS}" make build-binary'

release: build-in-docker execute-release
	@echo "Full in-docker release ${TAG}${FOR_PROVIDER} completed"

container: build-in-docker make-image
	@echo "Created in-docker image ${TAG}${FOR_PROVIDER}"

test-in-docker: clean docker-builder
	docker run ${RM_FLAG} -v `pwd`:/gopath/src/github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/:Z autoscaling-builder:latest bash -c 'cd /gopath/src/github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler && GO111MODULE=on go test -race ./... ${TAGS_FLAG}'

.PHONY: all deps build test-unit clean format execute-release dev-release docker-builder build-in-docker release generate

push: build
	docker build . -t  ${REGISTRY}/${REPOSITORY}/bcs-cluster-autoscaler:$(VERSION)
	docker push  ${REGISTRY}/${REPOSITORY}/bcs-cluster-autoscaler:$(VERSION)

build-e2e:
	$(ENVVAR) GOOS=$(GOOS) go build -o e2e ./test

container-e2e: build-e2e
	docker build . -t  ${REGISTRY}/${REPOSITORY}/bcs-cluster-autoscaler-e2e:$(VERSION) -f $(shell if [ -e ./test/Dockerfile ]; then echo ./test/Dockerfile; \
	else echo Dockerfile; fi) --label revision=$(REV)

push-e2e: container-e2e
	docker push ${REGISTRY}/${REPOSITORY}/bcs-cluster-autoscaler-e2e:$(VERSION)