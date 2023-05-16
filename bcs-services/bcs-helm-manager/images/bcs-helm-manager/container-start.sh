#!/bin/bash

module="bcs-helm-manager"

cd /data/bcs/${module}
chmod +x ${module}

#ready to start
exec /data/bcs/${module}/${module} $@
