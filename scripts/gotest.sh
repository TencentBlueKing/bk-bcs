#!/bin/sh
set -e

SKIP_DIR="\
bk-bcs/bcs-common/pkg/scheduler/mesosproto \
bk-bcs/bcs-mesos/bcs-container-executor/mesos \
bk-bcs/bcs-mesos/bcs-process-executor/process-executor/protobuf \
bk-bcs/bcs-mesos/pkg \
bk-bcs/bmsf-mesh/pkg \
bk-bcs/bcs-mesos/bcs-process-daemon"

PACKAGES=$(go list ../...)
for dir in $SKIP_DIR;do
	PACKAGES=`echo "$PACKAGES" | grep -v "$dir"`
done

#test:
	echo "go test"
	go test -cover=true $PACKAGES

#collect-cover-data:
	echo "collect-cover-data"
	echo "mode: count" > coverage-all.out
	for pkg in $PACKAGES;do
		echo "collect package:"${pkg}
		go test -v -coverprofile=coverage.out -covermode=count ${pkg};\
		if [ -f coverage.out ]; then\
			tail -n +2 coverage.out >> coverage-all.out;\
		fi\
	done

#test-cover-html:
	echo "test-cover-html"
	go tool cover -html=coverage-all.out -o coverage.html

#test-cover-func:
	echo "test-cover-func"
	go tool cover -func=coverage-all.out

#get total result
	total=`go tool cover -func=coverage-all.out | tail -n 1 | awk '{print $3}'`
	echo "total coverage: "${total}