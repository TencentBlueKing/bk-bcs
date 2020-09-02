#!/bin/bash

module="bcs-dns-service"

cd /data/bcs/${module}
chmod +x ${module}

#check configuration render
if [ $BCS_CONFIG_TYPE == "render" ]; then
  cat ${module}.config.template | envsubst | tee ${module}.config
fi

#ready to start
/data/bcs/${module}/${module} $@
