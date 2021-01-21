#!/usr/bin/env bash

set -eo pipefail
PWD=$(dirname "$(readlink -f "$0")")

source ./bscp.env

# op args
operator=$1
limit_version=$2

if  [ ! -n "$operator" ] ;then
    echo   "Usage:
     sh $0 {operator}"
else
 if [ ! -n "$limit_version" ] ;then
    # update all patchs
    curl -vv -X POST http://localhost:${PATCHER_PORT}/api/v2/patch/$operator
 else
    # update target limit version
    curl -vv -X POST http://localhost:${PATCHER_PORT}/api/v2/patch/$limit_version/$operator
 fi
fi
