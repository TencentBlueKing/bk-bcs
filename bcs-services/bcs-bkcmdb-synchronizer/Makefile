BIN = bcs-bkcmdb-synchronizer
IMAGE_REGISTRY = mirrors.tencent.com
IMAGE_REPO = bcs-bkcmdb-synchronizer
IMAGE_NAMESPACE = bcs
COMPILE_DATE := $(shell date +"%Y-%m-%d")
GOOS ?= linux
GOARCH ?= amd64


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

tidy:
	@go mod tidy

synchronizer: tidy
	@export CGO_ENABLED=1; GOOS=${GOOS} GOARCH=${GOARCH} go build ${LDFLAG} -o bin/${BIN} cmd/synchronizer/*.go

.PHONY: build
build: clean synchronizer

.PHONY: test
test:
	go test -v ./... -cover

docker-build: build
	@docker build --platform linux/amd64 . -t ${IMAGE_REPO}:${COMPILE_DATE}
	@echo ${IMAGE_REPO}:${COMPILE_DATE}

docker-push: docker-build
	@docker tag ${IMAGE_REPO}:${COMPILE_DATE} ${IMAGE_REGISTRY}/${IMAGE_NAMESPACE}/${IMAGE_REPO}:${COMPILE_DATE}
	@docker push ${IMAGE_REGISTRY}/${IMAGE_NAMESPACE}/${IMAGE_REPO}:${COMPILE_DATE}
	@echo ${IMAGE_REGISTRY}/${IMAGE_NAMESPACE}/${IMAGE_REPO}:${COMPILE_DATE}

.PHONY: clean
clean:
	@rm -rf bin

run:
	@./bin/${BIN} --conf ./config/${BIN}.json
