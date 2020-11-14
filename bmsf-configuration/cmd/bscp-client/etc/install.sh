#!/bin/bash

BSCP_CONFIG_PATH=/etc/bscp
BSCP_CONFIG_FILE=client.yaml
BSCP_CLIENT_NAME=bk-bscp-client
BSCP_CLIENT_VERSION=1.0.0

# isValidAddr check host and port valid
function isValidAddr() {
  local hostname=$1
  local hostname=(${hostname//:/ })
  local host=${hostname[0]}
  local port=${hostname[1]}
  local ret=1

  if [ ! -z "$host" ]; then
    ret=0
  fi

  if [ $ret -ne 0 ]; then
    return $ret
  fi

  if [[ $port -gt 65534 || $port -lt 1025 ]]; then
        ret=1
  fi
  return $ret
}

# add client to env
if [ -f "./$BSCP_CLIENT_NAME" ]
then
    cp -rf ./$BSCP_CLIENT_NAME /usr/bin/
else
    echo "The current file does not exist $BSCP_CLIENT_NAME"
    exit 1
fi

# create /etc/bscp/client.yaml
if [ ! -d "$BSCP_CONFIG_PATH" ]; then
    mkdir -p $BSCP_CONFIG_PATH
fi

if [ ! -f "$BSCP_CONFIG_FILE" ]; then
    touch "$BSCP_CONFIG_PATH/$BSCP_CONFIG_FILE"
fi

cd $BSCP_CONFIG_PATH
if [ $# -ne 1 ];then
    echo "Error: Initialization command format: ./init.sh bscp.bk.com:9510"
    exit 1
fi

address=$1
# judge address valid
isValidAddr $address
if [ $? -ne 0 ]; then
    echo "$address not a standard address format (bscp.bk.com:9510)"
    exit 1
fi

# write config content
echo -e "kind: bscp-client \nversion: $BSCP_CLIENT_VERSION \nhost: $address" > $BSCP_CONFIG_FILE