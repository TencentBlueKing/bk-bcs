# 逻辑集成功能
> 只需一个yaml文件，即可完成复杂的逻辑操作。

* 逻辑集成, 可帮助做快速接入，只需单一接口即可完成配置发布操作。
* 以yaml形式对所需要操作逻辑进行描述，BSCP逻辑集成模块将解析并按照描述完成全部逻辑操作。
* 逻辑集成接口只对复杂场景做支持，常规查询等场景请使用非yaml形式的RESTful API接口。

# 逻辑集成yaml规则说明

## Example

*创建business操作:*

```yaml
kind: business
version: v1
op: create
spec:
    businessName: business-itg
    depid: Blueking
    dbid: db01
    dbname: testdb
    memo: created by itg
```

* .kind必填字段标识操作资源business
* .version必填字段标识接口版本
* .op必填字段标识操作类型create为创建操作
* .spec必填字段为基础描述block
* .spec.businessName必填字段标识business名称
* .spec.depid必填字段标识business归属部门信息
* .spec.dbid必填字段标识business存储分片(一般由管理员分配后提供)
* .spec.dbname必填字段标识business存储所用mysql database名称(一般由管理员分配后提供)
* .spec.memo选填字段标识备注信息

*构建app及集群、大区结构操作*

```yaml
kind: construction
version: v1
op: create
spec:
    businessName: business-itg
    appName: app-itg
    deployType: 0
    memo: created by itg

construction:
    clusters:
    - name: cluster-itg-1
      rclusterid: rclusterid
      memo: created by itg
      zones:
      - name: zone-itg-1
        memo: created by itg
      - name: zone-itg-2
        memo: created by itg

    - name: cluster-itg-2
      rclusterid: rclusterid
      memo: created by itg
      zones:
      - name: zone-itg-3
        memo: created by itg
      - name: zone-itg-4
        memo: created by itg
```

* .kind必填字段标识操作资源app
* .version必填字段标识接口版本
* .op必填字段标识操作类型create为创建操作
* .spec必填字段为基础描述block
* .spec.businessName必填字段标识app所属business名称
* .spec.appName必填字段标识app名称
* .spec.deployType字段标识app部署类型，0：BCS容器部署， 1：GSE非容器部署，默认0
* .spec.memo选填字段标识备注信息
* .construction字段用于描述app的集群大区结构，若无该字段则只创建app, app已存在时不会重复创建
* .construction.clusters字段用于描述app的集群结构，若集群不存在则新建，若存在则不会重复创建
* .construction.clusters.name字段标识集群名称
* .construction.clusters.rclusterid字段标识集群物理位置信息
* .construction.clusters.memo字段为集群备注信息
* .construction.clusters.zones字段用于描述指定集群下的大区结构, 若大区不存在则新建，若已存在则不会重复创建
* .construction.clusters.zones.name字段描述大区的名称
* .construction.clusters.zones.memo字段为大区的备注信息

*修改配置操作(不使用模板渲染):*

```yaml
kind: commit
version: v1
op: commit
spec:
    businessName: business-itg
    appName: app-itg
    configSetName: server.yaml
    memo: created by itg

configs: 'dGhpcyBpcyBhIGV4YW1wbGU='

changes: changes
```

* .kind必填字段标识修改配置操作commit
* .version必填字段标识接口版本
* .op必填字段标识操作类型commit为配置变更操作
* .spec必填字段为基础描述block
* .spec.businessName必填字段标识目标配置集合所属business名称
* .spec.appName必填字段标识目标配置集合所属app名称
* .spec.configSetName必填字段标识目标配置集合名称, 若配置集合不存在则自动创建
* .spec.memo字段标识备注信息
* .configs字段在不使用模板渲染时必填，用于标识配置集合实体内容(二进制需以base64形式表示)
* .changes字段用于描述版本变化差异

*修改配置操作(使用模板渲染):*

```yaml
kind: commit
version: v1
op: commit
spec:
    businessName: business-itg
    appName: app-itg
    configSetName: server.yaml
    memo: created by itg

template:
    templateid: templateid
    template: |
        # single values
        k1: {{ .k1 }}
        k2: {{ .k2 }}

        # array values {{ range .k3 }}
        k3:
            - {{ . }}  {{end}}

    templateRule: |
        [
            {"type": 0, "name": "cluster-itg-1", "vars": { "k1": "v1a", "k2": 0, "k3": ["v3a", "v3b"]}},
            {"type": 1, "name": "zone-itg-1", "vars": {"k1": "v1b", "k2": 1, "k3": ["v3c", "v3d"]}}
        ]

changes: changes
```

* .kind必填字段标识修改配置操作commit
* .version必填字段标识接口版本
* .op必填字段标识操作类型commit为配置变更操作
* .spec必填字段为基础描述block
* .spec.businessName必填字段标识目标配置集合所属business名称
* .spec.appName必填字段标识目标配置集合所属app名称
* .spec.configSetName必填字段标识目标配置集合名称, 若配置集合不存在则自动创建
* .spec.memo字段标识备注信息
* .template字段用于描述模板相关信息
* .template.templateid可选字段用于描述使用的目标标识, 一般为其他系统对模板的管理标识
* .template.template必填字段用于标识配置模板内容(使用|进行长字符串换行), 规则参见模板渲染介绍文档
* .template.templateRule必填字段用于标识模板渲染规则(使用|进行长字符串换行), 规则参见模板渲染介绍文档
* .changes字段用于描述版本变化差异

*发布配置操作(不使用发布策略):*

```yaml
kind: publish
version: v1
op: publish
spec:
    businessName: business-itg
    appName: app-itg
    memo: created by itg

release:
    name: release-itg
    commitid: bc293eec-f558-11e9-a541-525400f99278
```

* .kind必填字段标识配置发布操作publish
* .version必填字段标识接口版本
* .op必填字段标识操作类型publish为配置发布操作
* .spec必填字段为基础描述block
* .spec.businessName必填字段标识目标配置集合所属business名称
* .spec.appName必填字段标识目标配置集合所属app名称
* .spec.memo字段标识备注信息
* .release必填字段为release相关信息描述block
* .release.name必填字段标识release名称
* .release.commitid字段标识目标提交ID, 系统将会根据该提交创建release进行发布

*发布配置操作(使用发布策略):*

```yaml
kind: publish
version: v1
op: publish
spec:
    businessName: business-itg
    appName: app-itg
    memo: created by itg

release:
    name: release-itg
    commitid: bc293eec-f558-11e9-a541-525400f99278
    strategyName: mystrategy
    strategy:
        clusterNames:
            - cluster-itg-1
            - cluster-itg-2
        zoneNames:
            - zone-itg-1
            - zone-itg-2
        dcs:
            - dc01
        ips:
            - 127.0.0.1
        labels:
            k: v
```

* .kind必填字段标识配置发布操作publish
* .version必填字段标识接口版本
* .op必填字段标识操作类型publish为配置发布操作
* .spec必填字段为基础描述block
* .spec.businessName必填字段标识目标配置集合所属business名称
* .spec.appName必填字段标识目标配置集合所属app名称
* .spec.memo字段标识备注信息
* .release必填字段为release相关信息描述block
* .release.name必填字段标识release名称
* .release.commitid字段标识目标提交ID, 系统将会根据该提交创建release进行发布
* .release.strategyName字段标识指定发布策略名称，若存在则复用，不存在则根据下方描述新建策略
* .release.strategy字段为发布策略strategy相关描述block
* .release.strategy.clusterNames字段描述策略规则涉及的全部cluster名称
* .release.strategy.zoneNames字段描述策略规则涉及的全部zone名称
* .release.strategy.dcs字段描述策略规则涉及的全部机房信息
* .release.strategy.ips字段描述策略规则涉及的全部实例ip信息
* .release.strategy.labels字段描述策略规则涉及的全部实例标签信息
