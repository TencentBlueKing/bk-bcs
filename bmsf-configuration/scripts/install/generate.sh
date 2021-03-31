#!/usr/bin/env bash

set -eo pipefail
PWD=$(dirname "$(readlink -f "$0")")

source ./bscp.env

# render server config templates.
eval "cat <<EOF
$(<../bk-bscp-apiserver/etc/server.yaml.template)
EOF
" > ../bk-bscp-apiserver/etc/server.yaml

eval "cat <<EOF
$(<../bk-bscp-authserver/etc/server.yaml.template)
EOF
" > ../bk-bscp-authserver/etc/server.yaml

eval "cat <<EOF
$(<../bk-bscp-patcher/etc/server.yaml.template)
EOF
" > ../bk-bscp-patcher/etc/server.yaml

eval "cat <<EOF
$(<../bk-bscp-patcher/etc/cron.yaml.template)
EOF
" > ../bk-bscp-patcher/etc/cron.yaml

eval "cat <<EOF
$(<../bk-bscp-configserver/etc/server.yaml.template)
EOF
" > ../bk-bscp-configserver/etc/server.yaml

eval "cat <<EOF
$(<../bk-bscp-templateserver/etc/server.yaml.template)
EOF
" > ../bk-bscp-templateserver/etc/server.yaml

eval "cat <<EOF
$(<../bk-bscp-datamanager/etc/server.yaml.template)
EOF
" > ../bk-bscp-datamanager/etc/server.yaml

eval "cat <<EOF
$(<../bk-bscp-gse-controller/etc/server.yaml.template)
EOF
" > ../bk-bscp-gse-controller/etc/server.yaml

eval "cat <<EOF
$(<../bk-bscp-tunnelserver/etc/server.yaml.template)
EOF
" > ../bk-bscp-tunnelserver/etc/server.yaml
