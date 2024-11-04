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
LIMIT_VALUE="204800"
BASE_YUM_LINK="http://mirrors.cloud.tencent.com/repo/centos7_base.repo"
BASE_EPEL_LINK="http://mirrors.cloud.tencent.com/repo/epel-7.repo"
BACKUPTIME=$(date +%Y%m%d_%H%M)
RPM_LIST=(ntpdate chrony screen pssh parallel zip unzip rsync gawk curl lsof tar sed iproute uuid psmisc wget rsync jq expect uuid bash-completion lsof openssl-devel readline-devel libcurl-devel libxml2-devel glibc-devel zlib-devel iproute procps-ng bind-utils)
NTP_SEVER="cn.pool.ntp.org"
if [[ -n ${BCS_OFFLINE:-} ]]; then
    SET_LIST=(set_kernel_params set_ulimit set_hostname set_selinux close_swap stop_firewalld install_tools)

else
    SET_LIST=(set_kernel_params set_ulimit set_hostname set_selinux close_swap stop_firewalld set_yum_repo install_tools set_time_sync)
fi

log() {
    echo "$@"
}

error() {
    echo "$@" 1>&2
}

fail() {
    echo "$@" 1>&2
    exit 1
}

warning() {
    echo "$@" 1>&2
    EXITCODE=$((EXITCODE + 1))
}

version() {
    echo "$PROGRAM version $VERSION"
}

usage() {
    cat <<EOF
用法:
    $PROGRAM [ -h --help -?  查看帮助 ]
            [ -i, --init     [必选] 检查单个/所有all，根据-l选项的输出 ]
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

set_kernel_params() {

    sysctl --system | grep net.ipv6.neigh.default.gc_thresh1 2>&1 >/dev/null
    if [ $? -eq 0 ]; then
        log "│   └──[SUCC] => 内核参数调优 配置成功"
    else
        cp /etc/sysctl.conf /etc/sysctl.conf.$BACKUPTIME
        cat >>"/etc/sysctl.conf" <<EOF
# init config start
#是否开启转发
net.ipv4.ip_forward=1
# 内核网络收发缓冲区大小
net.ipv4.tcp_rmem=4096 12582912 16777216
net.ipv4.tcp_wmem=4096 12582912 16777216
net.core.rmem_max=16777216
net.core.wmem_max=16777216
#TCP半连接队列长度
net.ipv4.tcp_max_syn_backlog=8096
# ipv6相关参数
## 开启ipv6转发
net.ipv6.conf.all.forwarding=1
## iptables 可以作用到网桥
net.bridge.bridge-nf-call-ip6tables=1
## ipv6 邻居条目调优
net.ipv6.neigh.default.gc_thresh1=2048
net.ipv6.neigh.default.gc_thresh2=8192
net.ipv6.neigh.default.gc_thresh3=16384
# init config end
EOF
        sysctl --system
        if [[ $? = 0 ]]; then
            log "│   └──[SUCC] => 内核参数调优 配置成功"
        else
            error "│   └──[FAIL] => 内核参数调优 配置失败"
        fi
    fi
}

set_ulimit() {
    curr=$(ulimit -n)
    if [[ $curr -ge $LIMIT_VALUE ]]; then
        log "│   └──[SUCC] => $1 : 当前值($curr)."
    else
        cp /etc/security/limits.conf /etc/security/limits.conf_$(date +%s)
        echo "root soft nofile $LIMIT_VALUE" >>/etc/security/limits.conf
        echo "root hard nofile $LIMIT_VALUE" >>/etc/security/limits.conf
        source /etc/profile
        if [[ $curr -ge 204800 ]]; then
            log "│   └──[SUCC] => $1 : 当前值($curr)."
        else
            error "│   └──[FAIL] => ulimit 配置失败"
        fi
    fi
}

set_hostname() {
    curr=$(hostname)
    if [[ "$curr" =~ "_" ]]; then
        hostnamectl set-hostname bkce$(hostname -I | awk -F'.' '{print $NF}')
        curr=$(hostname)
        if [[ "$curr" =~ "_" ]]; then
            error "│   └──[FAIL] => $1 : 当前主机名($curr).包含下划线"
        else
            log "│   └──[SUCC] => $1 : 当前主机名($curr)."
        fi

    else
        log "│   └──[SUCC] => $1 : 当前主机名($curr)."
    fi

}

set_selinux() {
    curr=$(getenforce)
    if [[ $selinux_status == "Enforcing" ]]; then
        setenforce 0 2>/dev/null
        sed -i 's/^SELINUX=.*/SELINUX=disabled/g' /etc/selinux/config
        curr=$(getenforce)
        if [[ $selinux_status == "Enforcing" ]]; then
            error "│   └──[FAIL] => $1 : 当前配置($curr)."
        else
            log "│   └──[SUCC] => $1 : 当前配置($curr)."
        fi
    else
        log "│   └──[SUCC] => $1 : 当前配置($curr)."
    fi
}

close_swap() {
    curr=$(free -h | grep Swap | awk '{print $2}')
    if [[ "$curr" == "0B" ]]; then
        log "│   └──[SUCC] => $1 : 当前值($curr)."
    else
        sed -ri 's/.*swap.*/#&/' /etc/fstab
        swapoff -a
        curr=$(free -h | grep Swap | awk '{print $2}')
        if [[ "$curr" == "0B" ]]; then
            log "│   └──[SUCC] => $1 : 当前值($curr)."
        else
            error "│   └──[FAIL] => $1 : 当前值($curr)."
        fi
    fi

}

stop_firewalld() {
    curr=$(systemctl is-active firewalld)
    if [[ $firewall_status == "active" ]]; then
        systemctl stop firewalld
        systemctl disable firewalld
        curr=$(systemctl is-active firewalld)
        if [[ $firewall_status == "active" ]]; then
            error "│   └──[FAIL] => $1 : 当前状态(RUNNING)."
        else
            log "│   └──[SUCC] => $1 : 当前状态(STOPPED)."
        fi
    else
        log "│   └──[SUCC] => $1 : 当前状态(STOPPED)."
    fi

}

set_yum_repo() {
    curr=$(yum repolist | grep EPEL | awk '{print $2,$3,$4,$5}')
    if [ "$curr" == "EPEL for redhat/centos 7" ]; then
        log "│   └──[SUCC] => $i : 当前配置($curr)."
    else
        whichWget=$(which wget 2>/dev/null)
        if [ "$whichWget" == "" ]; then
            yum install wget -y 2>&1 >/dev/null
        fi

        if ! grep -r BaseOS /etc/yum.repos.d/;then

          if [[ -f /etc/tlinux-release ]];then
            if grep -i "TencentOS Server 3.[0-9]*" /etc/tlinux-release;then
              BASE_YUM_LINK="http://mirrors.cloud.tencent.com/repo/centos8_base.repo"
            elif grep -i "TencentOS Server 2.[0-9]*" /etc/tlinux-release;then
              BASE_YUM_LINK="http://mirrors.cloud.tencent.com/repo/centos7_base.repo"
              BASE_EPEL_LINK="http://mirrors.cloud.tencent.com/repo/epel-7.repo"
            elif grep -i "Tencent tlinux release 2.[0-9]*" /etc/tlinux-release;then
              BASE_YUM_LINK="http://mirrors.cloud.tencent.com/repo/centos7_base.repo"
              BASE_EPEL_LINK="http://mirrors.cloud.tencent.com/repo/epel-7.repo"
            elif grep -i "Tencent linux release 2.[0-9]*" /etc/tlinux-release;then
              BASE_YUM_LINK="http://mirrors.cloud.tencent.com/repo/centos7_base.repo"
              BASE_EPEL_LINK="http://mirrors.cloud.tencent.com/repo/epel-7.repo"
            fi
          fi


          mv /etc/yum.repos.d/CentOS-Base.repo /etc/yum.repos.d/CentOS-Base.repo.$BACKUPTIME
          wget -O /etc/yum.repos.d/CentOS-Base.repo $BASE_YUM_LINK
          mv /etc/yum.repos.d/epel.repo /etc/yum.repos.d/epel.repo.$BACKUPTIME
          wget -O /etc/yum.repos.d/epel.repo $BASE_EPEL_LINK

          if [[ -f /etc/tlinux-release ]];then
            if grep -i "TencentOS Server 3.[0-9]*" /etc/tlinux-release;then
              sed -i "s/\$releasever/8/g" /etc/yum.repos.d/epel.repo
              sed -i "s/\$releasever/8/g" /etc/yum.repos.d/CentOS-Base.repo
            elif grep -i "TencentOS Server 2.[0-9]*" /etc/tlinux-release;then
              sed -i "s/\$releasever/7/g" /etc/yum.repos.d/epel.repo
              sed -i "s/\$releasever/7/g" /etc/yum.repos.d/CentOS-Base.repo
            elif grep -i "Tencent tlinux release 2.[0-9]*" /etc/tlinux-release;then
              sed -i "s/\$releasever/7/g" /etc/yum.repos.d/epel.repo
              sed -i "s/\$releasever/7/g" /etc/yum.repos.d/CentOS-Base.repo
            elif grep -i "Tencent linux release 2.[0-9]*" /etc/tlinux-release;then
              sed -i "s/\$releasever/7/g" /etc/yum.repos.d/epel.repo
              sed -i "s/\$releasever/7/g" /etc/yum.repos.d/CentOS-Base.repo
            fi
          fi

          yum clean all
          yum makecache
        fi

        curr=$(yum repolist | grep EPEL | awk '{print $2,$3,$4,$5}')
        if [ "$curr" == "EPEL for redhat/centos 7" ]; then
            log "│   └──[SUCC] => $i : 当前配置($curr)."
        else
            error "│   └──[FAIL] => $i : 当前配置($curr)."
        fi
    fi

}

install_tools() {
    yum -y install ${RPM_LIST[@]} 2>&1 >/dev/null
    local_rpm_rule=$(echo ${RPM_LIST[@]} | sed 's/ /|/g')
    currfmt=$(echo ${RPM_LIST[@]} | tr "\n" " " | sed -e 's/,$/\n/')
    local_rpm_list=($(rpm -qa --queryformat '%{NAME}\n' | grep -E "^($local_rpm_rule)" | tr "\n" " " | sed -e 's/,$/\n/'))
    log "│   └──[SUCC] => $i : 目前主机已安装($currfmt)."
    diff_array=($(echo "${local_rpm_list[@]}" "${RPM_LIST[@]}" | tr ' ' '\n' | sort | uniq -u))
    error "│   └──[FAIL] => $i : 目前主机未安装(${diff_array[@]})."
}

set_time_sync() {
    curr=$(date +'%Y-%m-%d %H:%M')
    OFFSET_TIME=$(ntpdate -q cn.pool.ntp.org | grep ntpdate | awk -F 'offset' '{print $2}' | awk '{print $1}' | cut -d '-' -f2 | awk -F "." '{print $1}')
    if [ $OFFSET_TIME -ge 3 ]; then
        ntpdate $NTP_SEVER >/dev/null
        mv /etc/chrony.conf /etc/chrony.conf.$BACKUPTIME
        cat >>"/etc/chrony.conf" <<EOF
    server $NTP_SEVER iburst
    driftfile /var/lib/chrony/drift
    allow 0.0.0.0/0
    makestep 1.0 3
    rtcsync
    local stratum 10
    logdir /var/log/chrony
    EOF
        systemctl stop chronyd >/dev/null
        systemctl start chronyd >/dev/null
        sleep 5
        systemctl stop chronyd
        systemctl start chronyd
        chronyc -a makestep >/dev/null
        timedatectl set-ntp yes
EOF
        nowtime=$(date +'%Y-%m-%d %H:%M')
        OFFSET_TIME=$(ntpdate -q cn.pool.ntp.org | grep ntpdate | awk -F 'offset' '{print $2}' | awk '{print $1}' | cut -d '-' -f2 | awk -F "." '{print $1}')
        if [ $OFFSET_TIME -ge 3 ]; then
            error "│   └──[FAIL] => $1 : 当前时间($nowtime).主机时间与时间服务器不一致"
        else
            log "│   └──[SUCC] => $1 : 当前时间($nowtime).主机时间与时间服务器基本一致"
        fi
    else
        log "│   └──[SUCC] => $i : 当前时间($curr)."
    fi
}


# 解析命令行参数，长短混合模式
(($# == 0)) && usage_and_exit 1
while (($# > 0)); do
    case "$1" in
    -i | --init)
        shift
        INSTALL_MODULE="$1"
        ;;
    -l | --list)
        shift
        echo ${SET_LIST[@]} | xargs -n 1
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
        error "不可识别的参数: $1"
        ;;
    *)
        break
        ;;
    esac
    shift
done

rerun=$INSTALL_MODULE
if [ "$rerun" == "" ]; then
    error "当前值为($curr).请输入"
    exit 1
fi

if [ "$rerun" != "all" ]; then
    [[ ${SET_LIST[@]/${rerun}/} == ${SET_LIST[@]} ]] && error "不可识别的参数: $rerun" && exit 1
    $rerun $rerun
else
    for index in "${!SET_LIST[@]}"; do
        let index2=$index+1
        log "├──[STEP ${index2}/${#SET_LIST[@]}] => [${SET_LIST[$index]}] - [$(date +'%H:%M:%S')]"
        ${SET_LIST[$index]} ${SET_LIST[$index]}
    done

fi
