#!/bin/bash

#check configuration render
if [ "x$BCS_CONFIG_TYPE" == "xrender" ]; then
  #apisix configuration
  cd /usr/local/apisix/conf
  cat config.yaml.template | envsubst | tee config.yaml
  echo ""
fi

apisix start

pid=`cat /usr/local/apisix/logs/nginx.pid`
ps -efww | grep nginx

echo "\n waiting for apisix initialization....(3s)"

sleep 3

echo "ready to register api-gateway tls certification..."

certContent=`cat ${apiGatewayCert} | sed ':label;N;s/\n/\\n/g;b label'`
keyContent=`cat ${apiGatewayKey} | sed ':label;N;s/\n/\\n/g;b label'`

curl -vv http://127.0.0.1:8000/apisix/admin/ssl/bkbcs \
  -H"X-API-KEY: ${adminToken}" -X PUT -d "{\"cert\":\"${certContent}\",\"key\":\"${keyContent}\",\"sni\":\"${ingressHostPattern}\"}"

#signal trap
echo "waiting for container exit signal~"
trap "apisix stop; echo api-gateway exit; exit" INT QUIT TERM

tail --pid $pid -f /dev/null
