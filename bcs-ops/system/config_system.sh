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

# configure host for k8s-api.bcs.local and sysctl optimization
VERSION=0.1.0
PROGRAM="$(basename "$0")"

SELF_DIR=$(dirname "$(readlink -f "$0")")
ROOT_DIR="${SELF_DIR}/.."

# PROJECTS=()

readonly PROGRAM SELF_DIR ROOT_DIR

usage_and_exit() {
  cat <<EOF
Usage:
    $PROGRAM
      [ -h --help -?  show usage ]
      [ -v -V --version show script version]
      [ -c --config
         dns: config bcs dns
         systcl: config bcs sysctl ]
      [ -d --del
	     dns: del bcs dns
		 systcl: del bcs sysctl ]
EOF
  exit "$1"
}

version() {
  echo "$PROGRAM version $VERSION"
}

# ToDo: Kernel optimization Split
add_kernel_para() {
  source /etc/os-release
  if [[ $VERSION_ID != "2.2" ]]; then
    echo br_netfilter ip_vs ip_vs_rr ip_vs_wrr ip_vs_sh nf_conntrack | xargs -n1 modprobe
    echo "br_netfilter" >> /etc/modules-load.d/k8s.conf
  fi

  [[ ${BCS_SYSCTL} == "1" ]] || return 0
  utils::log "INFO" "Adding kernel parameter"
  local total_mem page_size thread_size ipv6_status
  ipv6_status=${K8S_IPv6_STATUS:-"Disable"}
  total_mem=$(free -b | awk 'NR==2{print $2}')
  total_mem=${total_mem:-$((16 * 1024 * 1024 * 1024))}
  page_size=$(getconf PAGE_SIZE)
  page_size=${page_size:-4096}
  thread_size=$((page_size << 2))
  utils::log "INFO" "backup /etc/sysctl.conf"
  sed -ri.bcs-"$(date +%s)".bak '/bcs config begin/,/bcs config end/d' /etc/sysctl.conf
  cat >>"/etc/sysctl.conf" <<EOF
# bcs config begin
# 系统中每一个端口最大的监听队列的长度,这是个全局的参数,默认值128太小，32768跟友商一致
net.core.somaxconn=32768
# 大量短连接时，开启TIME-WAIT端口复用
net.ipv4.tcp_tw_reuse=1
# TCP半连接队列长度。值太小的话容易造成高并发时客户端连接请求被拒绝
net.ipv4.tcp_max_syn_backlog=8096 # RPS是将内核网络rx方向报文处理的软中断分配到合适CPU核，以提升网络应用整体性能的技术。这个参数设置RPS flow table大小
fs.inotify.max_user_instances=8192
# inotify watch总数量限制。调大该参数避免"Too many open files"错误
fs.inotify.max_user_watches=524288
# 使用bpf需要开启
net.core.bpf_jit_enable=1
# 使用bpf需要开启
net.core.bpf_jit_harden=1
# 使用bpf需要开启
net.core.bpf_jit_kallsyms=1
# 用于调节rx软中断周期中内核可以从驱动队列获取的最大报文数，以每CPU为基础有效，计算公式(dev_weight * dev_weight_tx_bias)。主要用于调节网络栈和CPU在tx上的不对称
net.core.dev_weight_tx_bias=1
# socket receive buffer大小
net.core.rmem_max=16777216
# RPS是将内核网络rx方向报文处理的软中断分配到合适CPU核，以提升网络应用整体性能的技术。这个参数设置RPS flow table大小
net.core.rps_sock_flow_entries=8192
# socket send buffer大小
net.core.wmem_max=16777216
# 避免"neighbor table overflow"错误(发生过真实客户案例，触发场景为节点数量超过1024，并且某应用需要跟所有节点通信)
net.ipv4.neigh.default.gc_thresh1=2048
# 同上
net.ipv4.neigh.default.gc_thresh2=8192
# 同上
net.ipv4.neigh.default.gc_thresh3=16384
$(
    [[ ${ipv6_status,,} != "disable" ]] && cat <<IPv6_EOF
net.ipv6.neigh.default.gc_thresh1=2048
net.ipv6.neigh.default.gc_thresh2=8192
net.ipv6.neigh.default.gc_thresh3=16384
IPv6_EOF
  )
# orphan socket是应用以及close但TCP栈还没有释放的socket（不包含TIME_WAIT和CLOSE_WAIT）。 适当调大此参数避免负载高时报'Out of socket memory'错误。32768跟友商一致。
net.ipv4.tcp_max_orphans=32768
# 代理程序(如nginx)容易产生大量TIME_WAIT状态的socket。适当调大这个参数避免"TCP: time wait bucket table overflow"错误。
net.ipv4.tcp_max_tw_buckets=16384
# TCP socket receive buffer大小。 太小会造成TCP连接throughput降低
net.ipv4.tcp_rmem=4096 12582912 16777216
# TCP socket send buffer大小。 太小会造成TCP连接throughput降低
net.ipv4.tcp_wmem=4096 12582912 16777216
# 控制每个进程的内存地址空间中 virtual memory area的数量
vm.max_map_count=262144
# 为了支持k8s service, 必须开启
net.ipv4.ip_forward=1
$(
    [[ ${ipv6_status,,} != "disable" ]] && cat <<IPv6_EOF
net.ipv6.conf.all.forwarding=1
IPv6_EOF
  )
# ubuntu系统上这个参数缺省为"/usr/share/apport/apport %p %s %c %P"。在容器中会造成无法生成core文件
kernel.core_pattern=core
# 内核在发生死锁或者死循环的时候可以触发panic,默认值是0.
kernel.softlockup_panic=0
# 使得iptable可以作用在网桥上
net.bridge.bridge-nf-call-ip6tables=1
net.bridge.bridge-nf-call-iptables=1
# 系统全局PID号数值的限制。
kernel.pid_max=$((4 * 1024 * 1024))
# 系统进程描述符总数量限制，根据内存大小动态计算得出，TOTAL_MEM为系统的内存总量，单位是字节，THREAD_SIZE默认为16，单位是kb。
kernel.threads-max=$((total_mem / (8 * thread_size)))
# 整个系统fd（包括socket）的总数量限制。根据内存大小动态计算得出，TOTAL_MEM为系统的内存总量，单位是字节，调大该参数避免"Too many open files"错误。
fs.file-max=$((total_mem / 10240))
fs.may_detach_mounts=1
# bcs config end
EOF
  sysctl --system
}

#######################################
# ulimit concurrent processes
# ulimit concurrently open file descriptors
#######################################
add_limits() {
  [[ ${BCS_SYSCTL} == "1" ]] || return 0
  utils::log "INFO" "Adding limits config"
  cat >/etc/security/limits.d/99-bcs.conf <<EOF
# bcs config begin
*   soft  nproc    1028546
*   hard  nproc    1028546
*   soft  nofile    204800
*   hard  nofile    204800
# bcs config end
EOF
}

add_master_hosts() {
  local master_iplist
  read -r -a master_iplist <<<"${K8S_CTRL_IP//,/ }"
  if [[ -z ${master_iplist[0]} ]]; then
    utils::log ERROR "BCS_K8S_CTRL_IP is null"
  fi
  "${ROOT_DIR}"/system/config_bcs_dns -u "${master_iplist[0]}" k8s-api.bcs.local
  return 0
}

add_mirror_hosts() {
  if [[ -n ${MIRROR_IP} ]]; then
    "${ROOT_DIR}"/system/config_bcs_dns -u "${MIRROR_IP}" mirrors.tencentyun.com
  fi
  return 0
}

set_hostname() {
    curr=$(hostname)
    if [[ "$curr" =~ "_" ]]; then
        utils::log "INFO" "Set hostname"
        hostnamectl set-hostname "$(hostname | sed 's/_/-/g')"
    fi

}

set_selinux() {
    selinux_status=$(getenforce)
    if [[ $selinux_status == "Enforcing" ]]; then
        utils::log "INFO" "Set selinux"
        setenforce 0 2>/dev/null
        sed -i 's/^SELINUX=.*/SELINUX=disabled/g' /etc/selinux/config
    fi
}

close_swap() {
    curr=$(free -h | grep Swap | awk '{print $2}')
    if [[ "$curr" != "0B" ]]; then
        utils::log "INFO" "Close swap"
        sed -ri 's/.*swap.*/#&/' /etc/fstab
        swapoff -a
    fi
}

stop_firewalld() {
	if systemctl is-active firewalld; then
        utils::log "INFO" "Stop firewalld"
        systemctl stop firewalld
        systemctl disable firewalld
    fi
}

config_dns() {
  add_master_hosts
  add_mirror_hosts
}

config_sysctl() {
  add_kernel_para
  add_limits
  set_hostname
  set_selinux
  close_swap
  stop_firewalld
}

del_dns() {
  "${ROOT_DIR}"/system/config_bcs_dns -d k8s-api.bcs.local
}

del_sysctl() {
  sed -ri.bcs-"$(date +%s)".bak '/bcs config begin/,/bcs config end/d' /etc/sysctl.conf
  sed -ri.bcs-"$(date +%s)".bak '/bcs config begin/,/bcs config end/d' /etc/security/limits.d/99-bcs.conf
}

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
}

main() {
  local source_files
  source_files=("${ROOT_DIR}/functions/utils.sh" "${ROOT_DIR}/env/bcs.env")
  for file in "${source_files[@]}"; do
    safe_source "$file"
  done

  local op_type project

  (($# == 0)) && usage_and_exit 1
  while (($# > 0)); do
    case "$1" in
      --help | -h | '-?')
        usage_and_exit 0
        ;;
      --version | -v | -V)
        version
        exit 0
        ;;
      --config | -c)
        op_type=config
        shift
        while (($# > 0)); do
          project=$1
          shift
          utils::check_op "${op_type}" "${project}"
          "${op_type}_${project}"
        done
        break
        ;;
      --del | -d)
        op_type=del
        shift
        while (($# > 0)); do
          project=$1
          shift
          utils::check_op "${op_type}" "${project}"
          "${op_type}_${project}"
        done
        break
        ;;
      -*)
        # ToDo: Unified standard error code
        export ERR_CODE=1
        utils::log "ERROR" "unkown para: $1"
        ;;
      *)
        usage_and_exit 0
        ;;
    esac
    shift
  done
  return 0
}

main "$@"
