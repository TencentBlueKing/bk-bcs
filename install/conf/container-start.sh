#!/bin/bash

#script use for starting necessary module for
# for test, demonstration

echo "begin to start modules $@"
# test modules
for module in $@
do
  echo "checking module $module binary"
  if [ ! -e /data/bcs/$module/$module ]; then 
    echo "lost module $module in /data/bcs"
    exit 1
  fi
  echo "checking module $module binary successfully"
done

#init kong conf & bkbcs-auth
cd /data/bcs/bcs-gateway-discovery
cp -r ./bkbcs-auth /usr/local/share/lua/5.1/kong/plugins/
#kong configuration
cp kong.conf.template /etc/kong/

#ready to start all specified module by 
# using container-start script
for module in $@
do
  echo "starting module $module ... "
  if [ $module == "bcs-kube-agent" ]; then
    cd /data/bcs/bcs-kube-agent
    ./bcs-kube-agent --bke-address=https://127.0.0.1:8443 \
      --cluster-id=BCS-K8S-00000 --insecureSkipVerify=true &
    continue
  fi
  if [ $module == "bcs-k8s-watch" ]; then
    cd /data/bcs/bcs-k8s-watch
    chmod +x container-start.sh
    /data/bcs/bcs-k8s-watch/container-start.sh --config \
      /data/bcs/bcs-k8s-watch/bcs-k8s-watch.json &
    continue
  fi
  cd /data/bcs/$module/
  chmod + container-start.sh
  ./container-start.sh -f $module.json &
done

echo "waiting for signal to exit..."
trap "exit 1" HUP INT PIPE QUIT TERM
