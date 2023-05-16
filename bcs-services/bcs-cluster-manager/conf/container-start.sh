#!/bin/bash

module="bcs-cluster-manager"

cd /data/bcs/${module}
chmod +x ${module}

#check configuration render
if [ "x$BCS_CONFIG_TYPE" == "xrender" ]; then
  cat ${module}.json.template | envsubst | tee ${module}.json
fi

#ready to start
/data/bcs/${module}/${module} $@
