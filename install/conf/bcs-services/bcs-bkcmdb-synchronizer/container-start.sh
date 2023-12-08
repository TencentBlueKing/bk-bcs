#!/bin/bash

module="bcs-bkcmdb-synchronizer"
bin="bcs-bkcmdb-synchronizer"

cd /data/bcs/${module}
chmod +x ${bin}

#check configuration render
if [ "x$BCS_CONFIG_TYPE" == "xrender" ]; then
  cat ${module}.json.template | envsubst | tee ${module}.json
fi

#ready to start
exec /data/bcs/${module}/${bin} $@