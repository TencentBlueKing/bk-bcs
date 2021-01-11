#!/usr/bin/env bash

set -eo pipefail

# generate templates for support-files.
cp -rf $1/middle-services/bscp-apiserver/etc/server.yaml.template      $2/\#etc\#bscp\#bk-bscp-apiserver.yaml
cp -rf $1/middle-services/bscp-authserver/etc/server.yaml.template     $2/\#etc\#bscp\#bk-bscp-authserver.yaml
cp -rf $1/atomic-services/bscp-configserver/etc/server.yaml.template   $2/\#etc\#bscp\#bk-bscp-configserver.yaml
cp -rf $1/atomic-services/bscp-templateserver/etc/server.yaml.template $2/\#etc\#bscp\#bk-bscp-templateserver.yaml
cp -rf $1/atomic-services/bscp-datamanager/etc/server.yaml.template    $2/\#etc\#bscp\#bk-bscp-datamanager.yaml
cp -rf $1/atomic-services/bscp-bcs-controller/etc/server.yaml.template $2/\#etc\#bscp\#bk-bscp-bcs-controller.yaml
cp -rf $1/atomic-services/bscp-gse-controller/etc/server.yaml.template $2/\#etc\#bscp\#bk-bscp-gse-controller.yaml
cp -rf $1/atomic-services/bscp-connserver/etc/server.yaml.template     $2/\#etc\#bscp\#bk-bscp-connserver.yaml
cp -rf $1/atomic-services/bscp-tunnelserver/etc/server.yaml.template   $2/\#etc\#bscp\#bk-bscp-tunnelserver.yaml
cp -rf $1/middle-services/bscp-patcher/etc/server.yaml.template        $2/\#etc\#bscp\#bk-bscp-patcher.yaml

# generate template vars.
cd $2 && sed -i 's/${/__/g' ./*.yaml && sed -i 's/}/__/g' ./*.yaml
