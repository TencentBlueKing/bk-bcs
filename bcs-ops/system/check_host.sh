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

# 通用脚本框架变量
VERSION="1.0.0"
PROGRAM="$(basename "$0")"

# 全局默认变量
SELF_DIR=$(dirname "$(readlink -f "$0")")
ROOT_DIR="${SELF_DIR}/.."
KERNEL_VERSION="3.10.0"
KERNEL_VERSION_IPv6="4.19.1"
LIMIT_VALUE="204800"
RPM_LIST="zip unzip curl lsof wget expect lsof socat procps-ng conntrack-tools \
openssl-devel readline-devel libcurl-devel libxml2-devel glibc-devel \
zlib-devel bind-utils bash-completion"
CHECK_LIST=(check_kernel check_swap check_selinux check_firewalld
  check_yum_proxy check_http_proxy check_openssl check_hostname check_tools)

# common
_version_ge() {
  test "$(echo "$@" | tr " " "\n" | sort -rV | head -n 1)" == "$1"
}

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

usage() {
  cat <<EOF
用法:
    $PROGRAM [ -h --help -?  查看帮助 ]
            [ -c, --check     [必选] 检查单个/所有all，根据-l选项的输出 ]
            [ -l, --list     [可选] 查看脚本支持选项 ]
            [ -v, --version     [可选] 查看脚本版本号 ]
EOF
}

usage_and_exit() {
  usage
  exit "$1"
}

version() {
  echo "$PROGRAM version $VERSION"
}

# 检查主机内核 check_kernel
check_kernel() {
  local currfmt kenerl_version
  currfmt=$(uname -r | cut -d '-' -f1)
  if [[ -n ${K8S_IPv6_STATUS} ]] && [[ ${K8S_IPv6_STATUS,,} != "disable" ]]; then
    kenerl_version=$KERNEL_VERSION_IPv6
  else
    kenerl_version=$KERNEL_VERSION
  fi

  if _version_ge "$currfmt" "$kenerl_version"; then
    utils::log "OK" "$1 : 当前配置($currfmt).内核版本大于或等于$kenerl_version"
  else
    utils::log "FATAL" "$1 : 当前配置($currfmt).内核版本小于$kenerl_version.\
k8s ipv4 内核版本要求不低于 $KERNEL_VERSION, \
k8s ipv6 内核版本要求不低于 $KERNEL_VERSION_IPv6"
  fi
}

# 检查关闭swap check_swap
check_swap() {
  local curr
  curr=$(free -h | grep Swap | awk '{print $2}')
  if [[ "$curr" == "0B" ]]; then
    utils::log "OK" "$1 : 当前值($curr)."
  else
    utils::log "ERROR" "$1 : 当前值($curr)."
    utils::log "WARN" "部署k8s建议是关闭swap分区，可以使用 swapoff -a 进行关闭"
  fi
}
# 关闭selinux check_selinux
check_selinux() {
  local curr
  curr=$(getenforce)
  if [[ $curr == "Enforcing" ]]; then
    utils::log "ERROR" "$1 : 当前配置($curr)."
    utils::log "WARN" "部署k8s建议是关闭selinux，可以使用 setenforce 0 进行关闭"
  else
    utils::log "OK" "$1 : 当前配置($curr)."
  fi
}
# 关闭防火墙 check_firewalld
check_firewalld() {
  local curr
  curr=$(systemctl is-active firewalld)
  if [[ $curr == "active" ]]; then
    utils::log "ERROR" "$1 : 当前状态(RUNNING)."
    utils::log "WARN" "部署k8s建议是关闭firewalld，\
可以使用 systemctl stop firewalld && systemctl disable firewalld 进行关闭"
  else
    utils::log "OK" "$1 : 当前状态(STOPPED)."
  fi
}
# 设置ulimit204800 check_ulimit
check_ulimit() {
  local curr
  curr=$(ulimit -n)
  if ((curr >= LIMIT_VALUE)); then
    utils::log "OK" "$1 : 当前值($curr)."
  else
    utils::log "WARN" "$1 : 当前值($curr)."
  fi
}
# 检查是否存在YUM代理 check_yum_proxy
check_yum_proxy() {
  curr=$(grep -i proxy /etc/yum.conf)
  if [ "$curr" != "" ]; then
    utils::log "WARN" "$1 : 当前配置($curr)."
  else
    utils::log "OK" "$1 : 当前配置(无代理)."
  fi
}
# 检查是否存在HTTP代理 check_http_proxy
check_http_proxy() {
  local curr currfmt
  curr=$(
    source /etc/profile
    env | grep -i 'http.*_proxy'
  )
  currfmt=${curr// /;}
  if [ "$currfmt" != "" ]; then
    utils::log "WARN" "$1 : 当前配置($currfmt)."
  else
    utils::log "OK" "$1 : 当前配置(无代理)."
  fi
}
# 检查openssl check_openssl
check_openssl() {
  local curr currformatch
  curr=$(openssl version | awk '{print $2}')
  currformatch=$(openssl version | awk '{print $2}' | awk -F'.' '{print $1$2}')
  if [ "$currformatch" == "11" ]; then
    utils::log "WARN" "$1 : 当前配置($curr)."
  else
    utils::log "OK" "$1 : 当前配置($curr)."
  fi
}
# 检查主机名 check_hostname
check_hostname() {
  local curr
  curr=$(hostname)
  if [[ "$curr" =~ "_" ]]; then
    utils::log "ERROR" "$1 : 当前主机名($curr).包含下划线"
    utils::log "WARN" "部署k8s建议是主机名不包含下划线，可以使用 hostnamectl set-hostname 主机名 进行修改"
  else
    utils::log "OK" "$1 : 当前主机名($curr)."
  fi
}
# 安装检查通用工具，所有主机 check_tools
check_tools() {
  local rpm_list_fmt rpm_list_array curr currfmt diff diff_array
  rpm_list_fmt=${RPM_LIST// /|}
  IFS=" " read -ra rpm_list_array <<<"$RPM_LIST"

  curr=$(rpm -qa --queryformat '%{NAME}\n' | grep -E "^($rpm_list_fmt)")
  currfmt=$(tr "\n" " " <<<"$curr" | sed -e 's/,$/\n/')
  utils::log "OK" "$1 : 目前主机已安装($currfmt)."
  IFS=" " read -ra currfmt <<<"${currfmt}"
  diff=$(echo "${currfmt[@]}" "${rpm_list_array[@]}" \
    | tr ' ' '\n' \
    | sort \
    | uniq -u)
  read -r -d '' -a diff_array <<<"$diff"
  if ((${#diff_array[@]} > 0)); then
    if [[ -z "${BCS_OFFLINE:-}" ]]; then
      utils::log "WARN" "$1 : 目前主机未安装(${diff_array[*]})."
    else
      utils::log "ERROR" "$1 : 目前主机未安装(${diff_array[*]})."
    fi
  fi
}

# 解析命令行参数，长短混合模式
(($# == 0)) && usage_and_exit 1
while (($# > 0)); do
  case "$1" in
    -c | --check)
      shift
      CHECK_MODULE="$1"
      ;;
    -l | --list)
      shift
      echo "${CHECK_LIST[@]}" | xargs -n 1
      exit 0
      ;;
    --help | -h | '-?')
      usage_and_exit 0
      ;;
    --version | -v | -V)
      version
      exit 0
      ;;
    -*)
      utils::log "ERROR" "unkown para: $1"
      ;;
    *)
      break
      ;;
  esac
  shift
done
safe_source "${ROOT_DIR}/functions/utils.sh"
rerun=$CHECK_MODULE
if [ "$rerun" == "" ]; then
  utils::log "FATAL" "当前值为($curr).请输入"
  exit 1
else
  for index in "${!CHECK_LIST[@]}"; do
    index2=$((index + 1))
    utils::log "INFO" "├──[STEP ${index2}/${#CHECK_LIST[@]}] => \
[${CHECK_LIST[$index]}] - [$(date +'%H:%M:%S')]"
    ${CHECK_LIST[$index]} "${CHECK_LIST[$index]}"
  done
fi
