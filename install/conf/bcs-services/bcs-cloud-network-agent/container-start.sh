#!/bin/bash

# copy cni files
cp /data/bcs/bcs-cni/bcs-eni /bcs/cni/bin/
cp /data/bcs/bcs-cni/bcs-eni.conf /bcs/cni/conf/

#check
module="bcs-cloud-network-agent"

cd /data/bcs/${module}
chmod +x ${module}
#check configuration render
if [ $BCS_CONFIG_TYPE == "render" ]; then
  cat ${module}.conf.template | envsubst | tee ${module}.json
fi
#ready to start
/data/bcs/${module}/${module} $@

