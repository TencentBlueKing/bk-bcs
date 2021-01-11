#!/usr/bin/env bash

set -eo pipefail

PRO_DIR="../../"

# rm target pkg/internal comms.
# rm -rf ${PRO_DIR}/pkg/xxxx

# rm sensitive codes.
# tunnelserver.
rm -rf ${PRO_DIR}/cmd/atomic-services/bscp-tunnelserver/actions
rm -rf ${PRO_DIR}/cmd/atomic-services/bscp-tunnelserver/cmd
rm -rf ${PRO_DIR}/cmd/atomic-services/bscp-tunnelserver/modules
rm -rf ${PRO_DIR}/cmd/atomic-services/bscp-tunnelserver/service

true > ${PRO_DIR}/cmd/atomic-services/bscp-tunnelserver/etc/server.yaml.template
cp -rf empty_main.go ${PRO_DIR}/cmd/atomic-services/bscp-tunnelserver/main.go

# gse plugin.
rm -rf ${PRO_DIR}/cmd/atomic-services/bscp-gse-plugin/actions
rm -rf ${PRO_DIR}/cmd/atomic-services/bscp-gse-plugin/modules
rm -rf ${PRO_DIR}/cmd/atomic-services/bscp-gse-plugin/service

true > ${PRO_DIR}/cmd/atomic-services/bscp-gse-plugin/etc/bscp.yaml.template
cp -rf empty_main.go ${PRO_DIR}/cmd/atomic-services/bscp-gse-plugin/main.go
