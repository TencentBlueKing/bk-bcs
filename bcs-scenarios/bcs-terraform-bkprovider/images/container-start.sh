#!/bin/bash

module="bcs-terraform-bkprovider"

cd /data/bcs/${module}
chmod +x ${module}

#check configuration render
if [ "$BCS_CONFIG_TYPE" == "render" ]; then
  cat ${module}.json.template | envsubst | tee ${module}.json
fi

#ready to start
exec /data/bcs/${module}/${module} $@