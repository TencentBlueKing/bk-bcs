#!/bin/sh
set -e

SKIP_DIR="\
bk-bcs/build \
bk-bcs/bcs-services/bcs-storage/storage/drivers/mongodb \
bk-bcs/bcs-services/bcs-metricservice/pkg/zk \
bk-bcs/bcs-mesos/bcs-scheduler/src/mesosproto \
bk-bcs/bcs-mesos/bcs-process-executor/process-executor/protobuf \
bk-bcs/bcs-services/bcs-log-webhook-server/pkg/client \
bk-bcs/bcs-mesos/bcs-container-executor/mesos"

PACKAGES=$(go list ../...)
for dir in $SKIP_DIR;do
	PACKAGES=`echo "$PACKAGES" | grep -v "$dir"`
done

# vet:
    echo "go vet"
    go vet -all $PACKAGES