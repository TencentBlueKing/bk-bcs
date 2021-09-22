#!/bin/bash

module="bcs-public-cluster-webhook"

cd /data/bcs/${module}
chmod +x ${module}


#ready to start
/data/bcs/${module}/${module} $@
