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

# k8s-components port iptables rule config

# apiserver	tcp/6443	secure-port
# controller	tcp/10257	secure-port
# scheduler	tcp/10259	secure-port
# etcd	tcp/2379, tcp/2380	advertise_port, peer-port
# kubelet	tcp/10250	metric-server need
# flannel-vxlan	linux	udp/8472
# flannel-vxlan	windows	udp/4789
# flannel-host-gw	linux	udp/51820, udp/51821 前者为ipv4，后者为ipv6
# flannell-udp		8285	仅当内核/网络不支持vxlan/host-gw

set -euo pipefail

PROGRAM="$(basename "$0")"
VERSION="1.0.0"
SELF_DIR=$(dirname "$(readlink -f "$0")")
ROOT_DIR="${SELF_DIR}/.."

readonly SELF_DIR ROOT_DIR PROGRAM VERSION

CHAIN="BCS-OPS-K8S"
COMMENT_PREFIX=${COMMENT_PREFIX:-"bcs_ops_k8s"}

readonly CHAIN COMMENT_PREFIX

usage_and_exit() {
  cat <<EOF
Usage:
    $PROGRAM
      [ help    show usage ]
      [ version show script version]
      [ add  <src_cidr4> <src_cidr6>   add k8s-components iptables rules]
      [ del  <src_cidr4> <src_cidr6>   del k8s-components iptables rules]
      [ clean   clean k8s-components iptables rules]
      [ list    list k8s-components iptables rules]
EOF
  exit "$1"
}

version() {
  echo "$PROGRAM version $VERSION"
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
}

check_IPv4Chain() {
  if ! iptables -L "$CHAIN" &>/dev/null; then
    utils::log "WARN" "iptables chain $CHAIN does not exist!"
    return 1
  fi
  return 0
}

check_IPv6Chain() {
  if ! ip6tables -L "$CHAIN" &>/dev/null; then
    utils::log "WARN" "ip6tables chain $CHAIN does not exist!"
    return 1
  fi
  return 0
}

add_chain() {
  if ! check_IPv4Chain; then
    iptables -N "$CHAIN"
  fi
  local rule_tpl=(-j "$CHAIN" -m comment --comment "${COMMENT_PREFIX}")
  if ! iptables -C INPUT "${rule_tpl[@]}" &>/dev/null; then
    iptables -A INPUT "${rule_tpl[@]}"
  fi

  if ! check_IPv6Chain; then
    ip6tables -N "$CHAIN"
  fi
  if ! ip6tables -C INPUT "${rule_tpl[@]}" &>/dev/null; then
    ip6tables -A INPUT "${rule_tpl[@]}"
  fi
}

list_rules() {
  if check_IPv4Chain; then
    utils::log "INFO" "ipv4 iptables chain ${CHAIN} rule"
    iptables -L "${CHAIN}" -n
  fi

  if check_IPv6Chain; then
    utils::log "INFO" "ipv6 iptables chain ${CHAIN} rule"
    ip6tables -L "${CHAIN}" -n
  fi
}

add_rules() {
  local protocol port src_cidr4 src_cidr6
  if (($# >= 3)); then
    protocol=$1
    port=$2
    comment="${COMMENT_PREFIX}:$3"
    src_cidr4=${src_cidr4:-}
    src_cidr6=${src_cidr6:-}
    if (($# >= 4)); then
      src_cidr4=$4
    fi
    if (($# >= 5)); then
      src_cidr6=$5
    fi
  else
    utils::log "ERROR" "least 3 para, now is $#"
    return 1
  fi

  # 检查参数是否为空
  if [[ -z "$protocol" ]] || [[ -z "$port" ]]; then
    utils::log "ERROR" "missing protocol or port"
    return 1
  fi

  local rule_tpl
  if [[ -n ${src_cidr4:-} ]]; then
    rule_tpl=(-s "$src_cidr4" -p "$protocol" --dport "$port" -j ACCEPT -m comment --comment "$comment")
  else
    rule_tpl=(-p "$protocol" --dport "$port" -j ACCEPT -m comment --comment "$comment")
  fi

  if iptables -C $CHAIN "${rule_tpl[@]}" &>/dev/null; then
    utils::log "WARN" "Accept IPv4 $protocol $port ${src_cidr4} rule already exists"
  else
    iptables -A ${CHAIN} "${rule_tpl[@]}"
    iptables-save >/etc/sysconfig/iptables
    utils::log "OK" "Accept IPv4 $protocol $port ${src_cidr4} rule added"
  fi

  if [[ -n ${src_cidr6:-} ]]; then
    rule_tpl=(-s "$src_cidr6" -p "$protocol" --dport "$port" -j ACCEPT -m comment --comment "$comment")
  else
    rule_tpl=(-p "$protocol" --dport "$port" -j ACCEPT -m comment --comment "$comment")
  fi

  if ip6tables -C $CHAIN "${rule_tpl[@]}" &>/dev/null; then
    utils::log "WARN" "Accept IPv6 $protocol $port ${src_cidr6} rule already exists"
  else
    ip6tables -A $CHAIN "${rule_tpl[@]}"
    ip6tables-save >/etc/sysconfig/ip6tables
    utils::log "OK" "Accept IPv6 $protocol $port ${src_cidr6} rule added"
  fi
}

del_rules() {
  local protocol port src_cidr4 src_cidr6
  if (($# >= 3)); then
    protocol=$1
    port=$2
    comment="${COMMENT_PREFIX}:$3"
    src_cidr4=${src_cidr4:-}
    src_cidr6=${src_cidr6:-}
    if (($# >= 4)); then
      src_cidr4=$4
    fi
    if (($# >= 5)); then
      src_cidr6=$5
    fi
  else
    utils::log "ERROR" "least 3 para, now is $#"
    return 1
  fi

  # 检查参数是否为空
  if [[ -z "$protocol" ]] || [[ -z "$port" ]]; then
    utils::log "ERROR" "missing protocol or port"
    return 1
  fi

  local rule_tpl
  if [[ -n ${src_cidr4:-} ]]; then
    rule_tpl=(-s "$src_cidr4" -p "$protocol" --dport "$port" -j ACCEPT -m comment --comment "$comment")
  else
    rule_tpl=(-p "$protocol" --dport "$port" -j ACCEPT -m comment --comment "$comment")
  fi

  if ! iptables -C $CHAIN "${rule_tpl[@]}" &>/dev/null; then
    utils::log "WARN" "Accept IPv4 $protocol $port ${src_cidr4} rule already delete"
  else
    iptables -D ${CHAIN} "${rule_tpl[@]}"
    iptables-save >/etc/sysconfig/iptables
    utils::log "OK" "Accept IPv4 $protocol $port ${src_cidr4} rule deleted"
  fi

  if [[ -n ${src_cidr6:-} ]]; then
    rule_tpl=(-s "$src_cidr6" -p "$protocol" --dport "$port" -j ACCEPT -m comment --comment "$comment")
  else
    rule_tpl=(-p "$protocol" --dport "$port" -j ACCEPT -m comment --comment "$comment")
  fi

  if ! ip6tables -C $CHAIN "${rule_tpl[@]}" &>/dev/null; then
    utils::log "WARN" "Accept IPv6 $protocol $port ${src_cidr6} rule already delete"
  else
    ip6tables -D $CHAIN "${rule_tpl[@]}"
    ip6tables-save >/etc/sysconfig/ip6tables
    utils::log "OK" "Accept IPv6 $protocol $port ${src_cidr6} rule deleted"
  fi
}

clean_rules() {
  if check_IPv4Chain; then
    utils::log "INFO" "iptables clean chain $CHAIN rules"
    iptables -F "$CHAIN" &>/dev/null
    iptables -D INPUT -j "$CHAIN" -m comment --comment "${COMMENT_PREFIX}"
    iptables -X "$CHAIN"
    utils::log "OK" "clean iptables chain $CHAIN"
  fi

  if check_IPv6Chain; then
    utils::log "INFO" "iptables clean chain $CHAIN rules"
    ip6tables -F "$CHAIN" &>/dev/null
    ip6tables -D INPUT -j "$CHAIN" -m comment --comment "${COMMENT_PREFIX}"
    ip6tables -X $CHAIN
    utils::log "OK" "clean ip6tables chain"
  fi
}

main() {
  local targets protocol port comment

  safe_source "${ROOT_DIR}/functions/utils.sh"
  targets=(tcp/6443/apiserver tcp/10257/controller tcp/10259/scheduler
    tcp/2379/etcd-adver tcp/2380/etcd-peer tcp/10250/kubelet udp/8472/flannel-vxlan)

  while (($# > 0)); do
    case "$1" in
      "help")
        usage_and_exit 0
        ;;
      "version")
        version
        exit 0
        ;;
      "add")
        add_chain
        shift
        for target in "${targets[@]}"; do
          IFS='/' read -r protocol port comment <<<"$target"
          add_rules "$protocol" "$port" "$comment" "$@"
        done
        break
        ;;
      "del")
        add_chain
        shift
        for target in "${targets[@]}"; do
          IFS='/' read -r protocol port comment <<<"$target"
          del_rules "$protocol" "$port" "$comment" "$@"
        done
        break
        ;;
      "list")
        list_rules
        break
        ;;
      "clean")
        clean_rules
        break
        ;;
      *)
        utils::log "WARN" "unkown param $1"
        break
        ;;
    esac
  done
  exit 0
}

main "$@"
