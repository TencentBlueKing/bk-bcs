#!/bin/bash

# clb-controller的镜像地址和tag
CLB_CTRL_IMAGE=''
# clb-controller的metric端口
CLB_CTRL_METRIC_PORT=''
# clb-controller使用的服务发现的机制可选[kubernetes, custom]
# kubernetes: 从k8s核心数据进行服务发现
# custom: 从k8s的crd进行服务发现
CLB_CTRL_REGISTRY=''
# 后端容器使用的ip地址类型，可选[underlay, overlay]
# underlay表示容器使用与物理机同一层的IP地址，clb会直接转发到容器ip
# overlay表示容器使用overlay网络，clb会转发到相应service节点的nodePort上
CLB_CTRL_BACKEND_IP_TYPE=''
# 绑定的clb的名字，如果clb实例不存在，则clb-controller会使用该名字创建一个新实例
CLB_NAME=''
# clb的网络类型，可选[public, private]
CLB_NET_TYPE=''
# clb-controller的实现方式，可选[api, sdk]
# api表示使用腾讯云2017版负载均衡api接口
# sdk表示使用腾讯云sdk3.0
CLB_IMPLEMENT=''
# 在调用腾讯云api时，可以选择通过CVM实例id来绑定clb，也可以通过弹性网卡ip地址来绑定clb
# 参数可选[cvm, eni]
# 当CLB_IMPLEMENT='api'时，CLB_BACKENDMODE支持eni方式（需要开启测试白名单），以及cvm方式
# 当CLB_IMPLEMENT='sdk'时，CLB_BACKENDMODE只支持cvm方式
CLB_BACKENDMODE=''
# clb所处的region，如[api-shanghai]
CLB_REGION=''
# 腾讯云账号对应的secret id
CLB_SECRETID=''
# 腾讯云账号对应的secret key
CLB_SECRETKEY=''
# 腾讯云账号对应的project id
CLB_PROJECTID=''
# 后端cvm实例所处的vpc id
CLB_VPCID=''
# private型clb使用的ip地址所属的子网id
CLB_SUBNET=''


CLUSTER_ROLE_BINDING_TEMPL='./cluster-role-binding.yaml'
CLUSTER_ROLE_TEMPL='./cluster-role.yaml'
SERVICE_ACCOUNT_TEMPL='./service-account.yaml'
DEPLOYMENT_TEMPL='./deployment.yaml'
SERVICE_TEMPL='./service.yaml'

cat ${CLUSTER_ROLE_BINDING_TEMPL}

echo "---"

cat ${CLUSTER_ROLE_TEMPL}

echo "---"

cat ${SERVICE_ACCOUNT_TEMPL}

echo "---"

sed -e "s#{{CLB_CTRL_IMAGE}}#${CLB_CTRL_IMAGE}#g" \
    -e "s/{{CLB_CTRL_METRIC_PORT}}/${CLB_CTRL_METRIC_PORT}/g" \
    -e "s/{{CLB_CTRL_REGISTRY}}/${CLB_CTRL_REGISTRY}/g" \
    -e "s/{{CLB_CTRL_BACKEND_IP_TYPE}}/${CLB_CTRL_BACKEND_IP_TYPE}/g" \
    -e "s/{{CLB_NAME}}/${CLB_NAME}/g" \
    -e "s/{{CLB_NET_TYPE}}/${CLB_NET_TYPE}/g" \
    -e "s/{{CLB_IMPLEMENT}}/${CLB_IMPLEMENT}/g" \
    -e "s/{{CLB_BACKENDMODE}}/${CLB_BACKENDMODE}/g" \
    -e "s/{{CLB_REGION}}/${CLB_REGION}/g" \
    -e "s/{{CLB_SECRETID}}/${CLB_SECRETID}/g" \
    -e "s/{{CLB_SECRETKEY}}/${CLB_SECRETKEY}/g" \
    -e "s/{{CLB_PROJECTID}}/${CLB_PROJECTID}/g" \
    -e "s/{{CLB_VPCID}}/${CLB_VPCID}/g" \
    -e "s/{{CLB_SUBNET}}/${CLB_SUBNET}/g" \
    "${DEPLOYMENT_TEMPL}"

echo "---"

sed -e "s/{{CLB_CTRL_METRIC_PORT}}/${CLB_CTRL_METRIC_PORT}/g" \
    "${SERVICE_TEMPL}"
