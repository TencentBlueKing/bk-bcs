# Metric API Doc



## 1. Set Metric

`POST` : `/metric/clustertype/{clusterType}/metrics`



| 参数               | 类型                   | 说明                                 | 必须     |
| :--------------- | -------------------- | ---------------------------------- | ------ |
| version          | string               | metric版本，同一个metric更新需要升版本号         | 是      |
| name             | string               | metric名字                           | 是      |
| namespace        | string               | metric所属的namespace                 | 是      |
| networkType      | string               | mesos networkType                  | mesos是 |
| networkMode      | string               | mesos networkMode                  | mesos是 |
| hostNetwork      | bool                 | k8s hostNetwork                    | k8s是   |
| dnsPolicy        | string               | k8s dnsPolicy                      | k8s是   |
| clusterID        | string               | 集群ID                               | 是      |
| clusterType      | string               | 集群类型  mesos/k8s                    | 是      |
| dataID           | int                  | 数据平台申请的dataid                      | 是      |
| port             | uint                 | 采集端口                               | 是      |
| uri              | string               | 采集uri                              | 是      |
| method           | string               | 采集http method   GET/POST           | 是      |
| head             | dict(string: string) | http head key-value pair           | 否      |
| parameters       | dict(string: string) | http parameters key-value pair     | 否      |
| selector         | dict(string: string) | metric选择器，匹配容器label                | 是      |
| frequency        | int                  | metric采集频率，秒/次                     | 是      |
| timeout          | int                  | 单次metric采集超时，秒                     | 是      |
| metricType       | string               | 一半格式为空，特殊格式如prometheus             | 否      |
| constLabels      | dict(string: string) | prometheus格式下，允许附加额外key-value pair | 否      |
| imageBase        | string               | 采集容器的镜像仓库domain                    | 是      |
| imagePullSecrets | list                 | k8s采集器拉取镜像时的用户权限secret             | 是      |
| tlsConfig        | tlsConfig            |                                    | 否      |



tlsConfig

| 参数           | 类型     | 说明                            | 必须   |
| ------------ | ------ | ----------------------------- | ---- |
| isTLS        | bool   | 是否https                       | 是    |
| ca           | string | 若为自签名证书，ca证书的base64           | 否    |
| clientCert   | string | 若有client端校验，clientCert的base64 | 否    |
| clientKey    | string | 若有client端校验，clientKey的base64  | 否    |
| clientKeyPwd | string | 若有clientKey有密码，clientKey的密码   | 否    |



## 2. Get Metric

`POST`:  `/metric/metrics`



| 参数        | 类型     | 说明                                    | 必须   |
| --------- | ------ | ------------------------------------- | ---- |
| clusterID | list   | clusterId的范围， ["clusterA", "cluster"] | 是    |
| name      | string | metric名字                              | 是    |



返回值：同`setmetric`的参数



## 3. Delete Metric

`DELETE`: `metric/clustertype/{clusterType}/clusters/{clusterID}/namespaces/{namespace}/metrics?name=val,val2,val3`

