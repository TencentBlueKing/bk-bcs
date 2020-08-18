#!/bin/bash

# use kubebuilder generate deepcopy code
make generate
# use kubebuilder generate yaml configs
make manifests

# use code-generate generate clientset lister and informer
chmod +x ./vendor/k8s.io/code-generator/generate-groups.sh

# generate the code with:
# --output-base    because this script should also be able to run inside the vendor dir of
#                  k8s.io/kubernetes. The output-base is needed for the generators to output into the vendor dir
#                  instead of the $GOPATH directly. For normal projects this can be dropped.

for dir_group in $(ls ./apis/); do
    for dir_version in $(ls "./apis/$dir_group/"); do
        echo "generate client,informer,lister for $dir_group/$dir_version"
        ./vendor/k8s.io/code-generator/generate-groups.sh \
            "client,informer,lister" \
            github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/generated \
            github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis \
            $dir_group:$dir_version \
            --go-header-file $(pwd)/hack/boilerplate.go.txt 
    done
done
