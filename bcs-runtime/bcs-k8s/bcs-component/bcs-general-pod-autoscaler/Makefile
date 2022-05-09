# Copyright 2019 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# you need to fill in the image url yourself
REGISTRY_NAME=
GIT_COMMIT=$(shell git rev-parse "HEAD^{commit}")
VERSION=$(shell git describe --tags --abbrev=14 "${GIT_COMMIT}^{commit}" --always)
BUILD_TIME=$(shell TZ=Asia/Shanghai date +%FT%T%z)
VERSION_KEY=github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-general-pod-autoscaler/pkg/version.Version
COMMIT_KEY=github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-general-pod-autoscaler/pkg/version.Commit

CMDS=build
all: test build

build: vet fmt build-gpa

build-gpa:
	CGO_ENABLED=0 GOOS=linux go build -ldflags "-X '$(VERSION_KEY)=$(VERSION)' -X '$(COMMIT_KEY)=$(GIT_COMMIT)'" -o ./bin/gpa ./cmd/gpa

container: build
	docker build -t $(REGISTRY_NAME)/bcs-general-pod-autoscaler:$(VERSION) -f $(shell if [ -e ./cmd/gpa/Dockerfile ]; then echo ./cmd/gpa/Dockerfile; else echo Dockerfile; fi) --label revision=$(REV) .

push: container
	docker push $(REGISTRY_NAME)/bcs-general-pod-autoscaler:$(VERSION)

test:
	go test ./pkg/...

autogen:
	go mod vendor
	bash hack/update-codegen.sh

vet:
	go vet ./pkg/...

fmt:
	go fmt ./pkg/...

clean:
	rm -r ./bin

generate-code:
	hack/generate-groups.sh  "defaulter,deepcopy,client,informer,lister" \
	github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-general-pod-autoscaler/pkg/client \
	github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-general-pod-autoscaler/pkg/apis \
	"autoscaling:v1alpha1" \
	--go-header-file ./hack/boilerplate.go.txt