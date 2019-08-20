#!/bin/bash

# set kernel parameters
sysctl -w net.ipv4.tcp_syncookies=1
sysctl -w net.ipv4.tcp_tw_reuse=1
sysctl -w net.ipv4.tcp_tw_recycle=1
sysctl -w net.ipv4.tcp_fin_timeout=30

proxy=$1
lb_interface_name=${LB_NETWORKCARD-"eth1"}
localIp=`ifconfig $lb_interface_name | /bin/grep 'inet 10\.\|inet 172\.\|inet 192\.\|inet 100\.\|inet 9\.' | awk '{print $2}'`

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

echo "wait network interface $lb_interface_name" >> result.txt
while true; do
  grep -q '^1$' "/sys/class/net/$lb_interface_name/carrier" && break
  ip link ls dev "$lb_interface_name" && break
  sleep 1
done > /dev/null 2>&1

#wait for NIC ready
sleep 10

#set haproxy && nginx template http timeout time from env
echo $LB_SESSION_TIMEOUT >> result.txt
sed -i "s/http-keep-alive 60/http-keep-alive $LB_SESSION_TIMEOUT/g" /bcs-lb/template/haproxy.cfg.template
sed -i "s/keepalive_timeout   65/keepalive_timeout $LB_SESSION_TIMEOUT/g" /bcs-lb/template/nginx.conf.template
if [ -z "$LB_IPFORWARD" ]; then 
    echo "LB_IPFORWARD is empty" >> result.txt
    sed -i "/forwardfor/d" /bcs-lb/template/haproxy.cfg.template
fi

if [ -z "$LB_HAPROXYMONITORPORT" ];then
    echo "haproxy monitor port is 8080" >> result.txt
else
    sed -i "s/8080/$LB_HAPROXYMONITORPORT/g" /bcs-lb/template/haproxy.cfg.template
fi

if [ -z "$LB_HAPROXYTHREADNUM" ];then
    echo "haproxy thread num is 4" >> result.txt
else
    sed -i "s/nbthread 4/nbthread $LB_HAPROXYTHREADNUM/g" /bcs-lb/template/haproxy.cfg.template
fi

#monitor loadbalance
lastTime=$(date +%s)
while true
do
  num=`ps -ef | grep bcs-loadbalance | grep -v grep | wc -l`
  if [ $num == 0 ]; then
    #starting loadbalance watch configuration and reload haproxy
    cd /bcs-lb
    /bcs-lb/bcs-loadbalance $@ --proxy $proxy --address $localIp > ./logs/bcs-loadbalance.log 2>&1 &
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