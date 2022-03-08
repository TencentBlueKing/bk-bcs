## 本次升级说明

- 本次升级将bcs-Saas cc项目、集群、集群节点数据，迁移到clusterManager中
- 升级流程，如bcs-SaaS cc有项目【A、B、C】，先迁移A项目的数据、A项目的集群数据、A项目的集群节点数据，再迁移B、C项目
- 数据写入到clusterManager前会与clusterManager原数据进行比较，再执行创建或更新操作，所以重复执行升级工具不导致clusterManager写入重复数据
- 每次运行升级工具都会从ssm获取token
