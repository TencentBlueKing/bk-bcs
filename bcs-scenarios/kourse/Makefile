
DIRECTORY := $(PWD)
PROJECT_PACKAGE := github.com/Tencent/bk-bcs/bcs-scenarios/kourse

pre:
	go mod tidy
	go fmt ./...

clean-generate:
	echo "Cleaning generated files..."
	rm -rf ./pkg/client
	rm -rf ./pkg/apis/tkex/v1alpha1/zz_generated.deepcopy.go

all: clean-generate generate-code

generate-code: pre
	hack/generate-groups.sh  "deepcopy,client,informer,lister" \
	$(PROJECT_PACKAGE)/pkg/client \
	$(PROJECT_PACKAGE)/pkg/apis \
	"tkex:v1alpha1" \
	--go-header-file ./hack/boilerplate.go.txt

generate-default: pre
	hack/generate-groups.sh  "defaulter" \
	$(PROJECT_PACKAGE)/pkg/client \
	$(PROJECT_PACKAGE)/pkg/apis \
	"tkex:v1alpha1" \
	--go-header-file ./hack/boilerplate.go.txt