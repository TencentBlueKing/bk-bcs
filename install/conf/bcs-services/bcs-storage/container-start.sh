#!/bin/bash

module="bcs-storage"

cd /data/bcs/${module}
chmod +x ${module}

#check configuration render
if [ "x$BCS_CONFIG_TYPE" == "xrender" ]; then
  cat ${module}.json.template | envsubst | tee ${module}.json
  if [ $mongodbEventHost ] && [ $mongodbEventUsername ] && [ $mongodbEventPassword ] && [ $mongodbEventMaxPoolSize ]
  then
    cat storage-database.conf.template | envsubst | tee storage-database.conf
  else
    mongodbEventHost="$mongodbHost";export mongodbEventHost
    mongodbEventUsername="$mongodbUsername";export mongodbEventUsername
    mongodbEventPassword="$mongodbPassword"; export mongodbEventPassword
    mongodbEventMaxPoolSize="$mongodbMaxPoolSize"; export mongodbEventMaxPoolSize
    cat storage-database.conf.template | envsubst | tee storage-database.conf
  fi
  cat queue.conf.template | envsubst | tee queue.conf
fi

#ready to start
exec /data/bcs/${module}/${module} $@
