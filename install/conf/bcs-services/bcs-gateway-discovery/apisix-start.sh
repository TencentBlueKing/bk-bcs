#!/bin/bash

#check configuration render
if [ $BCS_CONFIG_TYPE == "render" ]; then
  #apisix configuration
  cd /usr/local/apisix/conf
  cat config.yaml.template | envsubst | tee config.yaml
fi

apisix start

ps -efww | grep nginx

#setting tls certification by json
if [ "x${bcsSSLJSON}" == "x" ]; then
  echo "lost apisix bkbcs SSL json"
  exit 1
fi

echo "waiting for apisix initialization....(3s)"

sleep 3

echo "ready to registe api-gateway tls certification..."

curl http://127.0.0.1:8000/apisix/admin/ssl/bkbcs \
  -H"X-API-KEY: ${adminToken}" -X PUT -d@${bcsSSLJSON}

#signal trap
echo "waiting for container exit signal~"
trap "apisix stop" INT QUIT TERM
