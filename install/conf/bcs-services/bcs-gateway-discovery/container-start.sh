#!/bin/bash

module="bcs-gateway-discovery"

cd /data/bcs/${module}
chmod +x ${module}

#check configuration render
if [ $BCS_CONFIG_TYPE == "render" ]; then
  cd /data/bcs/${module}
  cat ${module}.json.template | envsubst | tee ${module}.json
fi

#starting module
cd /data/bcs/${module}
/data/bcs/${module}/${module} $@
