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

set -euo pipefail
trap "utils::on_ERR;" ERR

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
        echo "[ERROR]: FAIL to source, missing ${source_file}" >&2
        exit 1
    fi
    return 0
}

safe_source "${ROOT_DIR}/functions/utils.sh"

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

# 循环前先判断job是否有创建出来

if ! kubectl get job -n ${CHECK_NAMESAPCE} | grep ${JOB_NAME} >/dev/null 2>&1; then
    utils::log "INFO" "Failed to create Job resource,check cluster health fail $JOB_NAME"
    utils::log "ERROR" "use command: kubectl describe job/${JOB_NAME} -n ${CHECK_NAMESAPCE}"
    kubectl describe job/${JOB_NAME} -n ${CHECK_NAMESAPCE} | grep -A 50 Events
    exit 1
fi

while [[ $(($(date +%s) - $start_time)) -lt $duration ]]; do
    # 每隔1秒检测一次log文件的内容
    pod_name=$(kubectl get po -l job-name="$JOB_NAME" -n "$CHECK_NAMESAPCE" -o=jsonpath='{.items[*].metadata.name}')
    pod_status=$(kubectl get po -l job-name="$JOB_NAME" -n "$CHECK_NAMESAPCE" -o=jsonpath='{.items[*].status.phase}')
    job_status=$(kubectl get jobs ${JOB_NAME} -n ${CHECK_NAMESAPCE} -o=jsonpath='{.status.conditions[?(@.type=="Complete")].status}')
    # 判断pod是否创建出来
    if [ -n "$pod_name" ]; then
        utils::log "INFO" "create pod success"
    fi

    # Pod和Job状态都已完成，退出循环
    if [[ "$pod_status" == "Succeeded" && "$job_status" == "True" ]]; then
        utils::log "OK" "Pod and job run completed,Complete inspection"
        break
    fi
    utils::log "INFO" "cluster healthz checking, Please wait"
    case "${pod_status}" in
    Running)
        utils::log "INFO" "The status of $pod_name pod is $pod_status"
        ;;
    Pending)
        utils::log "INFO" "The status of $pod_name pod is $pod_status"
        ;;
    Failed)
        utils::log "FATAL" "The status of $pod_name pod is $pod_status"
        ;;
    Unknown)
        utils::log "INFO" "The status of $pod_name pod is $pod_status"
        ;;
    *)
        utils::log "INFO" "The status of $pod_name pod is $pod_status"
        ;;
    esac
    kubectl get event -n $CHECK_NAMESAPCE --field-selector involvedObject.name="$pod_name"
    sleep 3
    clear
done

# 判断pod是否创建了出来

pod_name=$(kubectl get po -l job-name="$JOB_NAME" -n "$CHECK_NAMESAPCE" -o=jsonpath='{.items[*].metadata.name}')
pod_status=$(kubectl get po -l job-name="$JOB_NAME" -n "$CHECK_NAMESAPCE" -o=jsonpath='{.items[*].status.phase}')
job_status=$(kubectl get jobs ${JOB_NAME} -n ${CHECK_NAMESAPCE} -o=jsonpath='{.status.conditions[?(@.type=="Complete")].status}')
if [ -z "$pod_name" ]; then
    utils::log "ERROR" "check cluster health fail $JOB_NAME timeout"
fi

utils::log "INFO" "The current status of $pod_name pod is $pod_status."
utils::log "INFO" "The current status of $JOB_NAME job is $job_status."

# 记录开始时间
start_time=$(date +%s)

if [[ "$job_status" != "True" ]]; then
    while [[ $(($(date +%s) - $start_time)) -lt $duration ]]; do
        utils::log "INFO" "Job ${JOB_NAME} status is ready to go. Please wait."
        kubectl describe job/${JOB_NAME} -n ${CHECK_NAMESAPCE}
        job_status=$(kubectl get jobs ${JOB_NAME} -n ${CHECK_NAMESAPCE} -o=jsonpath='{.status.conditions[?(@.type=="Complete")].status}')
        if [[ "$job_status" == "True" ]]; then
            utils::log "OK" "job run completed,Complete inspection"
            break
        fi
    done
elif [[ "$pod_status" == "Succeeded" && "$job_status" == "True" ]]; then
    # success
    utils::log "OK" "check cluster health success"
else
    # fail
    utils::log "INFO" "use command: kubectl describe po/${pod_name} -n ${CHECK_NAMESAPCE}"
    utils::log "FATAL" "check cluster health fail $JOB_NAME"
fi
kubectl delete job ${JOB_NAME} -n ${CHECK_NAMESAPCE} 2>/dev/null