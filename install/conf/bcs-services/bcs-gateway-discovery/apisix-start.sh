#!/bin/bash

sed -i '/^plugins:/a\  - bkbcs-auth' /usr/local/apisix/conf/config-default.yaml
sed -i '/^plugins:/a\  - bcs-auth' /usr/local/apisix/conf/config-default.yaml
sed -i 's/#- log-rotate/- log-rotate/ ' /usr/local/apisix/conf/config-default.yaml

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

echo "ready to registe api-gateway tls certification..."

certContent=`cat ${apiGatewayCert} | sed ':label;N;s/\n/\\n/g;b label'`
keyContent=`cat ${apiGatewayKey} | sed ':label;N;s/\n/\\n/g;b label'`

curl -vv http://127.0.0.1:8000/apisix/admin/ssl/bkbcs \
  -H"X-API-KEY: ${adminToken}" -X PUT -d "{\"cert\":\"${certContent}\",\"key\":\"${keyContent}\",\"snis\":[\"${ingressHostPattern}\", \"bcs-api-gateway\", \"bcs-api-gateway.${namespace}\", \"bcs-api-gateway.${namespace}.svc\", \"bcs-api-gateway.${namespace}.svc.cluster.local\"]}"

curl -vv -X PUT -H "X-API-KEY: ${adminToken}" 127.0.0.1:8000/apisix/admin/routes/kube-agent-tunnel -d "{\"name\":\"kube-agent-tunnel\",\"uri\":\"/clusters/*\",\"service_id\":\"clustermanager-http\",\"service_protocol\":\"http\",\"enable_websocket\":true,\"plugins\":{\"bkbcs-auth\":{\"token\":\"${gatewayToken}\",\"keepalive\":60,\"timeout\":1,\"module\":\"kubeagent\",\"bkbcs_auth_endpoints\":\"https:\\/\\/usermanager.bkbcs.tencent.com\"},\"request-id\":{\"include_in_response\":true,\"header_name\":\"X-Request-Id\"},\"proxy-rewrite\":{\"regex_uri\":[\"/clusters/(.*)\",\"/clustermanager/clusters/\$1\"],\"scheme\":\"https\"}}}"

#signal trap
echo "waiting for container exit signal~"
trap "apisix stop; echo api-gateway exit; exit" INT QUIT TERM

tail --pid $pid -f /dev/null
