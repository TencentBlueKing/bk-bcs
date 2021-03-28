#!/usr/bin/env bash

source ./bscp.env

cd ${HOME_DIR}/bk-bscp-apiserver && sh bk-bscp-apiserver.sh restart
cd ${HOME_DIR}/bk-bscp-authserver && sh bk-bscp-authserver.sh restart
cd ${HOME_DIR}/bk-bscp-configserver && sh bk-bscp-configserver.sh restart
cd ${HOME_DIR}/bk-bscp-templateserver && sh bk-bscp-templateserver.sh restart
cd ${HOME_DIR}/bk-bscp-datamanager && sh bk-bscp-datamanager.sh restart
cd ${HOME_DIR}/bk-bscp-gse-controller && sh bk-bscp-gse-controller.sh restart
cd ${HOME_DIR}/bk-bscp-patcher && sh bk-bscp-patcher.sh restart
cd ${HOME_DIR}/bk-bscp-tunnelserver && sh bk-bscp-tunnelserver.sh restart
