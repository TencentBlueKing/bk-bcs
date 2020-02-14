#!/bin/bash
objects=("versions" "admissionwebhookconfigurations" "agents" "agentschedinfos" "applications" "bcsclusteragentsettings" "bcscommandinfos" "bcsconfigmaps" "bcsendpoints" "bcssecrets" "bcsservices" "deployments" "frameworks" "taskgroups" "tasks")

for o in ${objects[@]};
do
  ns=`kubectl get $o.v2.bkbcs.tencent.com --all-namespaces |grep -v NAMESPACE |awk '{print $1}'`
  for i in $ns; do kubectl delete $o.v2.bkbcs.tencent.com --all -n $i; done
done