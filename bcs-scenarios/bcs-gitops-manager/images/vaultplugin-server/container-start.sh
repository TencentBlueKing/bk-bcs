#!/bin/bash

module="bcs-gitops-vaultplugin-server"

cd /data/bcs/${module}
chmod +x ${module}

#ready to start
exec /data/bcs/${module}/${module} $@