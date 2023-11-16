#!/bin/bash

#######################################
# Tencent is pleased to support the open source community by making Blueking Container Service available.
# Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
# Licensed under the MIT License (the "License"); you may not use this file except
# in compliance with the License. You may obtain a copy of the License at
# http://opensource.org/licenses/MIT
# Unless required by applicable law or agreed to in writing, software distributed under
# the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
# either express or implied. See the License for the specific language governing permissions and
# limitations under the License.
#######################################

SELF_DIR=$(dirname "$(readlink -f "$0")")
ROOT_DIR=${SELF_DIR}
readonly SELF_DIR ROOT_DIR

# only 1.2[0-1] to run
kubeadm reset phase update-cluster-status --v=5 || true
kubeadm reset phase remove-etcd-member --v=5

"${ROOT_DIR}"/clean_node.sh
