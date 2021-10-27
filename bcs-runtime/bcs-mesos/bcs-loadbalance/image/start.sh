#!/bin/bash

# set kernel parameters
sysctl -w net.ipv4.tcp_syncookies=1
sysctl -w net.ipv4.tcp_tw_reuse=1
sysctl -w net.ipv4.tcp_tw_recycle=1
sysctl -w net.ipv4.tcp_fin_timeout=30

proxy=$1
localIp=`hostname -I`

#start proxy module and shift cli arg
if [ $proxy == "haproxy" ]; then
  if [[ -z "$LB_PROXY_BINPATH" || -z "$LB_PROXY_CFGPATH" ]]; then
    /usr/sbin/haproxy -f /etc/haproxy/haproxy.cfg
  else
    "$LB_PROXY_BINPATH" -f "$LB_PROXY_BINPATH"
  fi
fi

if [ $proxy == "nginx" ]; then
  if [[ -z "$LB_PROXY_BINPATH" || -z "$LB_PROXY_CFGPATH" ]]; then
    /usr/local/nginx/sbin/nginx -c /usr/local/nginx/conf/nginx.conf
  else
    "$LB_PROXY_BINPATH" -c "$LB_PROXY_BINPATH"
  fi
fi
shift 1

if [ "$LB_HAPROXY_ENABLE_LOG" ]; then
  echo "start rsyslog daemon"
  rsyslogd
fi

#set haproxy && nginx template http timeout time from env

if [ $proxy == "haproxy" ] && [ -z "$LB_SHUNDOWNKILL" ] ;then
    /bcs-lb/kill_ha.sh &
fi

#monitor loadbalance
lastTime=$(date +%s)
while true
do
  num=`ps -ef | grep bcs-loadbalance | grep -v grep | wc -l`
  if [ $num == 0 ]; then
    #starting loadbalance watch configuration and reload haproxy
    cd /bcs-lb
    /bcs-lb/bcs-loadbalance $@ --proxy $proxy --address $localIp >> ./logs/bcs-loadbalance.log 2>&1 &
  fi
  currentHour=$(date +%H)
  currentMin=$(date +%M)
  currentTime=$(date +%s)
  let intervalTime=$currentTime-$lastTime
  if [ $currentHour -eq 03 ] && [ $currentMin -eq 00 ] && [ $intervalTime -gt 60 ]
  then
          lastTime=$(date +%s)
          sysctl vm.drop_caches=1
  fi
  sleep 15
done