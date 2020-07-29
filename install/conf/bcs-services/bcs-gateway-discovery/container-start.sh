#!/bin/bash

module="bcs-gateway-discovery"

cd /data/bcs/${module}
chmod +x ${module}

#check configuration render
if [ $BCS_CONFIG_TYPE == "render" ]; then
  cd /data/bcs/${module}
  cat ${module}.json.template | envsubst | tee ${module}.json

  #kong configuration
  cd /etc/kong
  cat kong.conf.template | envsubst | tee kong.conf
fi

#starting kong
echo "waiting 5 seconds for postgresql..."
sleep 5
kong migrations bootstrap -c /etc/kong/kong.conf
kong start -c /etc/kong/kong.conf

#starting module
cd /data/bcs/${module}
/data/bcs/${module}/${module} $@
