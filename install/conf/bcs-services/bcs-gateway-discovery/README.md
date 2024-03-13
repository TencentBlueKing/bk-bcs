### 镜像构建

1. 目前BCS的API网关最新版本已升级到蓝鲸微网关方案，Dockerfile使用Dockerfile.micro-gateway-apisix。
2. BCS的API网关基于蓝鲸API网关微网关方案，其镜像也基于蓝鲸微网关镜像。蓝鲸微网关镜像构建详见：[蓝鲸 API 网关](https://github.com/TencentBlueKing/blueking-apigateway-apisix/tree/master)。
3. BCS网关依赖组件gomicro-discover-operator的代码以及构建方法详见：bcs-services/bcs-gateway-discovery/gomicro-discovery-operator。
