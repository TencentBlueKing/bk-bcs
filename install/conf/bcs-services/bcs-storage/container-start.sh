#!/bin/bash

module="bcs-storage"

cd /data/bcs/${module}
chmod +x ${module}

#check configuration render
if [ $BCS_CONFIG_TYPE == "render" ]; then
  cat ${module}.json.template | envsubst | tee ${module}.json
  cat storage-database.conf.template | envsubst | tee storage-database.conf
fi

#ready to start
/data/bcs/${module}/${module} $@
