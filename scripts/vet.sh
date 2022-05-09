#!/bin/sh
set -e

SKIP_DIR="\
bk-bcs/build \
bk-bcs/vendor \
bk-bcs/bcs-services/bcs-storage/storage/drivers/mongodb \
bk-bcs/bcs-services/bcs-metricservice/pkg/zk \
bk-bcs/bcs-common/pkg/scheduler/mesosproto \
bk-bcs/bcs-mesos/bcs-process-executor/process-executor/protobuf \
bk-bcs/bcs-services/bcs-webhook-server/pkg/client \
bk-bcs/bcs-k8s/bcs-k8s-watch/pkg/kubefed/client \
bk-bcs/bcs-k8s/bcs-egress \
bk-bcs/bcs-mesos/bcs-container-executor/mesos \
bk-bcs/bmsf-configuration \
bk-bcs/bcs-services/bcs-project"

PACKAGES=$(go list ../...)
for dir in $SKIP_DIR;do
	PACKAGES=`echo "$PACKAGES" | grep -v "$dir"`
done

# vet:
echo "go vet"
go vet -all $PACKAGES
