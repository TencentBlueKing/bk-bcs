GOPATH:=$(shell go env GOPATH)
VERSION=$(shell git describe --tags || echo "unknown version")
BUILDTIME=$(shell date -u)
GOBUILD=CGO_ENABLED=0 go build -trimpath

# init the build information
ifdef HASTAG
	GITTAG=${HASTAG}
else
	GITTAG=$(shell git describe --always)
endif

BUILDTIME=$(shell date +%Y-%m-%dT%T%z)
GITHASH=$(shell git rev-parse HEAD)
VERSION=${GITTAG}-$(shell date +%y.%m.%d)
bcs_edition?=inner_edition

LDFLAG=-ldflags "-X github.com/Tencent/bk-bcs/bcs-common/common/version.BcsVersion=${VERSION} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/version.BcsBuildTime=${BUILDTIME} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/version.BcsGitHash=${GITHASH} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/version.BcsTag=${GITTAG} \
 -X github.com/Tencent/bk-bcs/bcs-common/common/version.BcsEdition=${bcs_edition}"

.PHONY: init
init:
	@go get -u google.golang.org/protobuf/proto
	@go install github.com/golang/protobuf/protoc-gen-go@latest
	@go install github.com/asim/go-micro/cmd/protoc-gen-micro/v4@latest

.PHONY: lint
lint:
	@golangci-lint run --issues-exit-code=0 --fix

.PHONY: proto
proto:
	@protoc --proto_path=. --micro_out=. --go_out=:. proto/bcs-webconsole.proto

.PHONY: update
update:
	@go get -u

.PHONY: tidy
tidy:
	@go mod tidy

.PHONY: build
build:
	${GOBUILD} ${LDFLAG} -o bcs-webconsole *.go

.PHONY: test
test:
	@go test -v ./... -cover

.PHONY: docker
docker:
	@docker build -t bcs-webconsole:latest .

.PHONY: kubectld
kubectld:
	cd kubectld && docker build . -t kubectld:latest

# @go generate ./console/i18n/translations/translations.go
.PHONY: i18n
i18n:
	@cd ./console/i18n/translations && gotext -srclang=zh update -out=catalog.go -lang=en,zh ../../../
	@cd ./console/i18n/translations && cp ./locales/en/out.gotext.json ./locales/en/messages.gotext.json
