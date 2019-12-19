#!/usr/bin/env bash

while getopts ":h:k:p:z:i:r:c:" opt; do
  case $opt in
    k) KUBE_MASTER_URL="$OPTARG"
    ;;
    h) HOST_IP="$OPTARG"
    ;;
    p) HOST_PORT="$OPTARG"
    ;;
    z) ZK_URLS="$OPTARG"
    ;;
    i) REPORT_IP="$OPTARG"
    ;;
    r) REPORT_PORT="$OPTARG"
    ;;
    c) CLUSTER_ID="$OPTARG"
    ;;
    \?) echo "Invalid option -$OPTARG" >&2
    ;;
  esac
done

echo ${REPORT_IP}

chmod +x bcs-k8s-driver

./bcs-k8s-driver \
 --address=${HOST_IP} \
 --secure-port ${HOST_PORT} \
 --kube-master-url=${KUBE_MASTER_URL} \
 --host-ip=${HOST_IP} \
 --server-ca-file=bcs-ca.crt \
 --server-cert-file=bcs-server.crt \
 --server-key-file=bcs-server.key \
 --zk-url=${ZK_URLS} \
 --custom-report-address=${REPORT_IP} \
 --custom-report-port=${REPORT_PORT} \
 --custom-cluster-id=${CLUSTER_ID}
 > ./k8s-driver.log 2>&1 &
 