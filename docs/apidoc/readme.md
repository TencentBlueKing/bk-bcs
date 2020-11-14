# BCS API说明

## API入口

* bcs-api：http接口，计划持续维护至1.22.x，不再集成1.17.x后增加的模块转发
* bcs-api-gateway：兼容bcs-api http接口，增加https、grpc支持

api-gateway模块转发规则说明：

* bcs-storage模块(全局容器数据存储)
  * /bcsapi/v4/storage/              => /bcsstorage/v1/
* bcs-mesos-driver模块(mesos api转发规则)
  * /bcsapi/v4/scheduler/mesos/      => /mesosdriver/v4/
  * /bcsapi/v1/                      => /mesosdriver/v4/
* bcs-kube-agent模块（kube-apiserver透传）
  * /tunnels/clusters/${clusterId}/  => /
* bcs-k8s-driver模块(deprecated)
  * /bcsapi/v4/scheduler/k8s/        => /k8sdriver/v4/
* bcs-network-detection
  * /bcsapi/v4/detection/            => /detection/v4/
* bcs-user-manager
  * /bcsapi/v4/usermanager/          => /usermanager/
* （新增）bcs-mesh-manager
  * http: /bcsapi/v4/meshmanager/    => /meshmanager/
  * grpc: /meshmanager.MeshManager/  => /meshmanager.MeshManager/
* （新增）bcs-log-manager
  * http: /bcsapi/v4/logmanager/     => /logmanager/
  * grpc: /logmanager.LogManager/    => /logmanager.LogManager/
* （新增）bcs-data-manager
  * http: /bcsapi/v4/datamanager/    => /datamanager/
  * grpc: /datamanager.DataManager/  => /datamanager.DataManager/

请求api-gateway依赖：

* http header： Authorization

该token由bkbcs系统运维签发。

## 文档说明

* [bcs-api-gateway使用说明](../features/bcs-api-gateway/api-gateway方案.md)
* [bcs-user-manager API](./bcs-user-manager.md)
* [mesos方案API](./api-scheduler.md)
* [k8s方案API](./k8s.md)
* [bcs-storage全局容器数据存储](./api-storage.md)
* [detection方案API](./detection.md)
