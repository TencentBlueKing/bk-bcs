GOPATH:=$(shell go env GOPATH)

OUTPUT=build/bcs.${VERSION}

#LDFLAG=-ldflags "-X github.com/Tencent/bk-bcs/bcs-common/common/static.ZookeeperClientUser=${bcs_zk_client_user} \
# -X github.com/Tencent/bk-bcs/bcs-common/common/static.ZookeeperClientPwd=${bcs_zk_client_pwd} \
# -X github.com/Tencent/bk-bcs/bcs-common/common/static.EncryptionKey=${bcs_encryption_key} \
# -X github.com/Tencent/bk-bcs/bcs-common/common/static.ServerCertPwd=${bcs_server_cert_pwd} \
# -X github.com/Tencent/bk-bcs/bcs-common/common/static.ClientCertPwd=${bcs_client_cert_pwd} \
# -X github.com/Tencent/bk-bcs/bcs-common/common/static.LicenseServerClientCertPwd=${bcs_license_server_client_cert_pwd} \
# -X github.com/Tencent/bk-bcs/bcs-common/common/static.BcsDefaultUser=${bcs_registry_default_user} \
# -X github.com/Tencent/bk-bcs/bcs-common/common/static.BcsDefaultPasswd=${bcs_registry_default_pwd} \
# -X github.com/Tencent/bk-bcs/bcs-common/common/version.BcsVersion=${VERSION} \
# -X github.com/Tencent/bk-bcs/bcs-common/common/version.BcsBuildTime=${BUILDTIME} \
# -X github.com/Tencent/bk-bcs/bcs-common/common/version.BcsGitHash=${GITHASH} \
# -X github.com/Tencent/bk-bcs/bcs-common/common/version.BcsTag=${GITTAG} \
# -X github.com/Tencent/bk-bcs/bcs-common/common/version.BcsEdition=${bcs_edition}"

.PHONY: init
init:
	go get -u google.golang.org/protobuf/proto
	go install github.com/golang/protobuf/protoc-gen-go@latest
	go install github.com/asim/go-micro/cmd/protoc-gen-micro/v4@v4.6.0
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.7.3
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.7.3

.PHONY: proto
proto:
	protoc --proto_path=. \
    		--proto_path=../../bcs-services/bcs-cluster-manager/third_party \
    		--go_out=. \
    		--micro_out=. \
    		--go-grpc_out=. \
    		--validate_out=paths=source_relative,lang=go:. \
    		--grpc-gateway_out=logtostderr=true,register_func_suffix=Gw:. --swagger_out=logtostderr=true:. \
    		proto/bcs-terraform-bkprovider.proto

.PHONY: update
update:
	go get -u

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: build
build: tidy
	mkdir -p ${OUTPUT}/bcs-terraform-bkprovider
	cp -r images/* ${OUTPUT}/bcs-terraform-bkprovider/
	go build ${LDFLAG} -o ${OUTPUT}/bcs-terraform-bkprovider/bcs-terraform-bkprovider main.go

.PHONY: test
test:
	go test -v ./... -cover

.PHONY: docker
docker:
	docker build -t bcs-terraform-bkprovider:latest .