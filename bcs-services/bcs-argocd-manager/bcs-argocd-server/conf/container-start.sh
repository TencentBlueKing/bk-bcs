#!/bin/bash

module="bcs-argocd-manager"
submodule="bcs-argocd-server"

cd /data/bcs/${module}/${submodule}
chmod +x ${submodule}

#check configuration render
if [ "x$BCS_CONFIG_TYPE" == "xrender" ]; then
  cat ${submodule}.json.template | envsubst | tee ${submodule}.json
fi

#ready to start
/data/bcs/${module}/${submodule}/${submodule} $@
