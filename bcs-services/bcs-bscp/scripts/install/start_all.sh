#!/usr/bin/env bash

cd ../bk-bscp-apiserver && sh bk-bscp-apiserver.sh start
cd ../bk-bscp-configserver && sh bk-bscp-configserver.sh start
cd ../bk-bscp-dataservice && sh bk-bscp-dataservice.sh start
cd ../bk-bscp-feedserver && sh bk-bscp-feedserver.sh start
cd ../bk-bscp-authserver && sh bk-bscp-authserver.sh start
cd ../bk-bscp-cacheservice && sh bk-bscp-cacheservice.sh start