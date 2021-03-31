#!/usr/bin/env bash

source ./bscp.env

cd ${HOME_DIR}/bk-bscp-apiserver && sh bk-bscp-apiserver.sh start
cd ${HOME_DIR}/bk-bscp-authserver && sh bk-bscp-authserver.sh start
cd ${HOME_DIR}/bk-bscp-configserver && sh bk-bscp-configserver.sh start
cd ${HOME_DIR}/bk-bscp-templateserver && sh bk-bscp-templateserver.sh start
cd ${HOME_DIR}/bk-bscp-datamanager && sh bk-bscp-datamanager.sh start
cd ${HOME_DIR}/bk-bscp-gse-controller && sh bk-bscp-gse-controller.sh start
cd ${HOME_DIR}/bk-bscp-patcher && sh bk-bscp-patcher.sh start
cd ${HOME_DIR}/bk-bscp-tunnelserver && sh bk-bscp-tunnelserver.sh start
