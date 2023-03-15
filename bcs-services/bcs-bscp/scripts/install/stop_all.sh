#!/usr/bin/env bash

cd ../bk-bscp-apiserver && sh bk-bscp-apiserver.sh stop
cd ../bk-bscp-configserver && sh bk-bscp-configserver.sh stop
cd ../bk-bscp-dataservice && sh bk-bscp-dataservice.sh stop
cd ../bk-bscp-feedserver && sh bk-bscp-feedserver.sh stop
cd ../bk-bscp-authserver && sh bk-bscp-authserver.sh stop
cd ../bk-bscp-cacheservice && sh bk-bscp-cacheservice.sh stop