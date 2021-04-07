#!/usr/bin/env bash

set -eo pipefail

source ./bscp.env
source ./generate.sh
source ./init_db.sh

mkdir -p ${HOME_DIR}
mkdir -p ${LOG_DIR}

# install list
cp -rf ../bk-bscp-apiserver ${HOME_DIR}
cp -rf ../bk-bscp-authserver ${HOME_DIR}
cp -rf ../bk-bscp-patcher ${HOME_DIR}
cp -rf ../bk-bscp-configserver ${HOME_DIR}
cp -rf ../bk-bscp-templateserver ${HOME_DIR}
cp -rf ../bk-bscp-datamanager ${HOME_DIR}
cp -rf ../bk-bscp-gse-controller ${HOME_DIR}
cp -rf ../bk-bscp-tunnelserver ${HOME_DIR}

# rm temp files
#find ${HOME_DIR} -name *.template | xargs rm
