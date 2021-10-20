# BCS版本升级模块方案设计


### 背景

bcs的版本升级，有时会涉及到数据的变更，比如：

- db数据结构的调整，如表、字段、索引的增删改
- 已有数据的迁移改造等

通过开发该版本升级模块，能通过一键化运行升级程序，将bcs后台从代码变更前的任意版本升级到当前代码版本。

比如能从v1.21.2升级到v1.21.7，也能从v1.21.5升级到v1.21.7，中间能跨多个版本，完成多次数据迁移。



### 升级方式

1. 运行升级脚本upgrade.sh
2. 调用升级API



### 升级程序实现原理

#### 原理概述

bcs数据版本升级主要涉及到当前版本号和标记了版本号的升级程序。

当前版本可在mongodb中执行 `db.bcs_upgrader.find({"type": "version"}) `输出当前版本信息，升级程序的具体实现逻辑放在 bcs-services/bcs-upgrader/upgrades 的多个目录下，每个目录都代表一个升级程序（一次数据迁移）。

比如 u1.21.202109151209 目录为一个升级程序，则该升级程序的版本号即为目录名本身u1.21.202109151209，执行完该目录下的升级程序后，bcs数据版本将会变为 u1.21.202109151209。

假如upgrades目录下有如下多个升级程序目录，则升级完成后，将升级为版本号最高的那个版本，即u1.21.202109171502。

```bash
bcs_upgrader: tester$ tree .
.
└── upgrades
    ├── u1.21.202109151209
    ├── u1.21.202109161305
    └── u1.21.202109171502
```



#### 升级程序执行逻辑

获取版本号从小到大排序的所有升级程序--> 获取当前版本 --> 遍历所有升级程序版本，和当前版本比较，只执行比当前版本大的升级程序-->更新db里的当前版本信息



#### 升级程序版本号说明

以u1.21.202107261209为例，由四部分组成：

版本号前缀：u（u代表upgrade）

主版本号（Major）：1

次版本号（Minor）：21

修订版本号（Patch）：202107261209（为编写升级程序的时间：年月日时分秒，用日期而不用bcs自身的发行修订版本号，是因为发行版本号可能会调整，不一定由开发者管理）



#### 版本比较说明

依次比较主版本，次版本和修订版本号

eg：u1.21.202109121940 < u1.22.202109121940< u1.22.202109151940



### mongodb里bcs_upgrader表设计

```go
type VersionInfo struct {
Type           string    `bson:"type"`
PreVersion     string    `bson:"pre_version"`
CurrentVersion string    `bson:"current_version"`
Edition        string    `bson:"edition"`
LastTime       time.Time `bson:"last_time"`
}
```



##### bcs_upgrader表中的数据示例

```json
{
  "_id" : ObjectId("610292814d9dfcabbda79f51"),
  "type": "version",
  "pre_version" : "u1.21.202109121940",
  "current_version" : "u1.21.202109151520",
  "edition" : "community",
  "last_time" : ISODate("2021-09-15T11:35:45.359Z")
}
```



### BCS版本升级操作示例

从 v1.21.2升级到 v1.21.x的输出：

```bash
bk-bcs: tester$ ./upgrade.sh
{
    "result": true,
    "code": 0,
    "message": "success",
    "data": {
        "msg": "upgrade success",
        "pre_version": "u1.21.202109121940",
        "current_version": "u1.21.202109151520",
        "finished_upgrades": [
            "u1.21.202109121940",
            "u1.21.202109131230",
            "u1.21.202109151520"
        ]
    }
}
bk-bcs: tester$ 
```

#### 输出字段说明

通用字段

- result ：请求成功与否。true:请求成功；false请求失败
- code：错误码。 0表示success，>0表示失败错误
- message ：请求失败返回的错误信息
- data ：升级结果详情



db 升级相关字段

- msg为升级执行结果描述
- pre_version 为db升级前数据库中bcs数据处于的版本
- current_version 为执行完升级脚本后bcs数据处于的版本
- finished_upgrades 为本次升级脚本执行时完成的升级版本

