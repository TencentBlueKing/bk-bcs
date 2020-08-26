#!/usr/bin/env bash

# Copyright 2017 The Kubernetes Authors.
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

set -o errexit
set -o nounset
set -o pipefail

# generate the code with:
bash ../vendor/k8s.io/code-generator/generate-groups.sh "deepcopy,client,informer,lister" \
  github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/client github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/apis \
  tkex:v1alpha1 \
  --go-header-file $(pwd)/boilerplate.go.txt
