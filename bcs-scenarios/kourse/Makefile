
DIRECTORY := $(PWD)
KOURSE_PROJECT_PACKAGE := github.com/Tencent/bk-bcs/bcs-scenarios/kourse

pre:
	go mod tidy
	go fmt ./...

clean-generate:
	echo "Cleaning generated files..."
	rm -rf ./pkg/client
	rm -rf ./pkg/apis/tkex/v1alpha1/zz_generated.deepcopy.go
	rm -rf ./manifests/

generate-code: pre
	hack/generate-groups.sh  "deepcopy,client,informer,lister" \
	$(KOURSE_PROJECT_PACKAGE)/pkg/client \
	$(KOURSE_PROJECT_PACKAGE)/pkg/apis \
	"tkex:v1alpha1" \
	--go-header-file ./hack/boilerplate.go.txt

generate-crd: pre
	mkdir -p manifests
	controller-gen paths=./... \
	crd:crdVersions=v1,maxDescLen=1000,generateEmbeddedObjectMeta=true \
	rbac:roleName=operator \
	output:crd:artifacts:config=./manifests

generate-default: pre
	hack/generate-groups.sh  "defaulter" \
	$(KOURSE_PROJECT_PACKAGE)/pkg/client \
	$(KOURSE_PROJECT_PACKAGE)/pkg/apis \
	"tkex:v1alpha1" \
	--go-header-file ./hack/boilerplate.go.txt

gamedeployment:
	echo "Do not support gamedeployment"

gamestatefulset:
	echo "Do not support gamestatefulset"

hook-operator:
	echo "Do not support hook operator"