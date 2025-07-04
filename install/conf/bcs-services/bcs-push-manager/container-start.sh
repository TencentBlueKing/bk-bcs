#!/bin/bash

module="bcs-push-manager"

cd /data/bcs/${module}
chmod +x ${module}

#ready to start
exec /data/bcs/${module}/${module} $@