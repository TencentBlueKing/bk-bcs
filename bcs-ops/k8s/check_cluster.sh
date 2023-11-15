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

# set -euo pipefail
# trap "utils::on_ERR;" ERR

SELF_DIR=$(dirname "$(readlink -f "$0")")
ROOT_DIR="${SELF_DIR}/.."
JOB_NAME=bcs-blackbox-job
CHECK_NAMESAPCE=default
IMAGE_NAME=hub.bktencent.com/library/hello-world

readonly SELF_DIR ROOT_DIR

#######################################
# check file and source
# Arguments:
# $1: source_file
# Return:
# if file exists, source return 0; else exit 1
#######################################
safe_source() {
    local source_file=$1
    if [[ -f ${source_file} ]]; then
        #shellcheck source=/dev/null
        source "${source_file}"
    else
        echo "[ERROR]: FAIL to source, missing ${source_file}"
        exit 1
    fi
    return 0
}

safe_source "${ROOT_DIR}/functions/utils.sh"

kubectl delete job ${JOB_NAME} -n ${CHECK_NAMESAPCE}

cat <<EOF | kubectl apply -f -
apiVersion: batch/v1
kind: Job
metadata:
  name: ${JOB_NAME}
  namespace: ${CHECK_NAMESAPCE}
spec:
  template:
    metadata:
      labels:
        test-yaml: test-yaml
    spec:
      tolerations:
      - operator: "Exists"
        effect: "NoSchedule"
      containers:
      - name: blackbox
        image: ${IMAGE_NAME}
        imagePullPolicy: Always
        resources:
          limits:
            cpu: "100m"
            memory: "100Mi"
          requests:
            cpu: "100m"
            memory: "100Mi"
      restartPolicy: Never
  backoffLimit: 1
EOF

# 设置监测时间为1分钟
duration=60

# 记录开始时间
start_time=$(date +%s)

kubectl get job -n ${CHECK_NAMESAPCE} | grep ${JOB_NAME} >/dev/null 2>&1

if [ $? -ne 0 ]; then
    utils::log "FATAL" "check cluster health fail $JOB_NAME"
    kubectl describe job ${JOB_NAME} -n ${CHECK_NAMESAPCE} | grep -A 50 Events
    exit
fi

while [[ $(($(date +%s) - $start_time)) -lt $duration ]]; do
    # 每隔1秒检测一次log文件的内容
    pod_name=$(kubectl get po -l job-name="$JOB_NAME" -n "$CHECK_NAMESAPCE" -o=jsonpath='{.items[*].metadata.name}')
    pod_status=$(kubectl get po -l job-name="$JOB_NAME" -n "$CHECK_NAMESAPCE" -o=jsonpath='{.items[*].status.phase}')

    if [ -n "$pod_name" ]; then
        utils::log "INFO" "create pod success"
    fi

    case "${pod_status}" in
    Succeeded)
        utils::log "OK" "check cluster health success"
        break
        ;;
    *)
        utils::log "INFO" "cluster healthz checking, Please wait"
        kubectl get po -l job-name="$JOB_NAME" -n default -o=jsonpath='{.items[*].status.containerStatuses[*].state}' | jq
        kubectl describe po "$pod_name" -n "$CHECK_NAMESAPCE" | tail -n 1
        sleep 3
        clear
        ;;
    esac
    sleep 1
done

curr=$(kubectl get jobs ${JOB_NAME} -n ${CHECK_NAMESAPCE} -o=jsonpath='{.status.conditions[?(@.type=="Complete")].status}')
if [[ $curr != "True" ]]; then
    utils::log "ERROR" "You can use the command == kubectl describe po ${pod_name} -n ${CHECK_NAMESAPCE} == to view more detailed information"
    utils::log "FATAL" "check cluster health fail $JOB_NAME timeout"
fi