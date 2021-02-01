BK-BSCP 集成文档
================

[TOC]

# 0.概述

本文意在说明如何快速理解BSCP系统概念，进程平台集成或开发自己的配置中心SaaS。

# 1.资源说明

## 1.1 配置资源

配置场景的资源含义说明,

- `App（应用模块）`: BSCP内最小管理单元，可以对应为一个模块如LoginServer、GameServer等, 也可以对应不同集群环境的相同模块如Test-GameServer、Online-GameServer表示GameServer测试环境和线上环境;
- `AppInstance（节点实例）`: 目标配置生效节点，与App（应用模块）对应，可以是进程环境中的一个进程实例，也可以是容器环境中的一个POD;
- `Config（配置）`: 配置，配置文件，由名称(Key)、内容(Value)、路径(path)和响应权限描述构成, 是BSCP系统内配置的最小单元，在传统配置场景下可以理解为一个配置文件;
- `Commit（提交）`: 配置改动提交，即对某个Config（配置）进行修改并提交到系统;
- `MultiCommit （混合提交）`: 相同App（应用模块）下的N个Config（配置）的改动提交，即同时修改并提交N个配置;
- `Content（内容）`: 某个Config（配置）的单次Commit（提交）后的内容，一个Config（配置）可以经模板渲染为不同AppInstance（节点实例）产生不同内容。故此一次Commit（提交）可以关联N个不同的Content（内容），每个Content（内容）中可设置索引（Index）匹配不同AppInstance（节点实例）去获取配置;
- `Release（版本）`: 某个Config（配置）的单次Commit（提交）后，可以创建出Release（版本）并进行发布操作，发布后AppInstance（节点实例）会实时进行Config（配置）的热更新;
- `MultiRelease（混合版本）`: 相同App（应用模块）下的N个Config（配置）的改动提交后，基于MultiCommit （混合提交）创建出MultiRelease（混合版本）并进行发布， 发布后AppInstance（节点实例）会实时进行N个Config（配置）的热更新;
- `Strategy（发布策略）`: 可为某个MultiRelease（混合版本）创建发布策略，支持逻辑与、逻辑或标签集合，支持类Bash的eq、ge、ne、lt、gt、in、not in语义;

## 1.2 模板资源

模板相关的资源说明(若不使用BSCP的模板引擎则不需要关注),

- `Template（配置模板）`: 通用配置模板，其中描述了配置的名称(Key)、内容(Value)、路径(path)和响应权限描述，可以关联多个App（应用模块），关联之后在App（应用模块）下会创建出实例化的Config（配置）。其可以为支持模板渲染的配置或纯粹的文本文件、二进制文件;
- `TemplateVersion（配置模板版本）`: Template（配置模板）的版本定义，相同的Template（配置模板）可以创建N个版本进行管理, 发布时可由某个TemplateVersion（配置模板版本）进行发布;
- `TemplateBind（模板绑定）`: 某个Template（配置模板）与某个App（应用模块）之间的绑定关系， 一个Template（配置模板）与多个App（应用模块）产生绑定后则存在多个绑定关系, 绑定之后模板的信息修改将会同步所有其绑定的App（应用模块）;

模板引擎：

- `Golang Template引擎`: Golang内置模板渲染引擎;
- `Python Mako引擎`: Python Make模板渲染引擎;

# 2.集成说明

## 2.1 拓扑关联

### 2.1.1 模块拓扑关联

App（应用模块）是系统内最小管理单元，可将你的SaaS产品中的模块与此进行关联,

- `服务关联`: 从服务维度进行关联，如LoginServer、GameServer等，每种Server对应一个App（应用模块）;
- `环境关联`: 从环境维度进行关联，如测试集群、线上集群、深圳集群、上海集群、IOS用户集群、PC用户集群等, 相同的服务模块在不同的环境或集群下对应不同的App（应用模块）;

图示,

```shell
LoginServer      ----> app_id_1(xxxxxx)
GameServer       ----> app_id_2(xxxxxx)
Test-LoginServer ----> app_id_3(xxxxxx)
Test-GameServer  ----> app_id_4(xxxxxx)

或

Cluster-ShenZhen-LoginServer      ----> app_id_1(xxxxxx)
Cluster-ShangHai-GameServer       ----> app_id_2(xxxxxx)
Test-Cluster-ShenZhen-LoginServer ----> app_id_3(xxxxxx)
Test-Cluster-ShangHai-GameServer  ----> app_id_4(xxxxxx)
```

综上，App（应用模块）为最小操作单元，可根据上层SaaS产品中的场景按需关联。

### 2.1.2 项目关联

不同的App（应用模块）如LoginServer、GameServer归属于某个业务下，如英雄联盟、王者荣耀等，
故此需要将App（应用模块）按照业务的维度进行收拢管理。

BSCP创建App（应用模块）时，需传入所属的业务ID(biz_id)，即为某个业务创建一个新的App（应用模块）。
业务ID(biz_id)来自集成用户, 一般来自蓝鲸内部CMDB系统中，即BSCP内的App（应用模块）归属于CMDB中的某个业务。

```shell
biz_id_1(xxxxxx) ----> app_id_1(xxxxxx)
                |----> app_id_2(xxxxxx)
                |----> app_id_3(xxxxxx)
                ......

biz_id_2(xxxxxx) ----> app_id_4(xxxxxx)
                |----> app_id_5(xxxxxx)
                |----> app_id_6(xxxxxx)
                ......

biz_id_3(xxxxxx) ----> app_id_7(xxxxxx)
                |----> app_id_8(xxxxxx)
                |----> app_id_9(xxxxxx)
                ......
```

### 2.1.3 模板关联

Template（配置模板）为通用Config（配置）模板，归属于某个业务ID下，可以关联到该业务ID下的所有App（应用模块）上，
配置模板中描述了配置的名称(Key)、内容(Value)、路径(path)和响应权限描述, 关联指定App（应用模块）后会在该App（应用模块）中创建
相同名称(Key)、内容(Value)、路径(path)和响应权限的Config（配置）。

```shell
biz_id_1(xxxxxx) ----> template_id_1(xxxxxx) ----> app_id_1(xxxxxx)
                |                           |----> app_id_2(xxxxxx)
                |
                |----> template_id_2(xxxxxx) ----> app_id_1(xxxxxx)
                |                           |----> app_id_2(xxxxxx)
                |                           |----> app_id_3(xxxxxx)
                |
                |----> template_id_3(xxxxxx) ----> app_id_1(xxxxxx)
                |
                ......

biz_id_2(xxxxxx) ----> template_id_4(xxxxxx) ----> app_id_4(xxxxxx)
                |                           |----> app_id_5(xxxxxx)
                |
                |----> template_id_5(xxxxxx) ----> app_id_4(xxxxxx)
                |                           |----> app_id_5(xxxxxx)
                |                           |----> app_id_6(xxxxxx)
                |
                |----> template_id_6(xxxxxx) ----> app_id_4(xxxxxx)
                |
                ......
```

如上所示，biz_id_1下有3个Template（配置模板）template_id_1、template_id_2、template_id_3，

template_id_1关联了biz_id_1下的app_id_1、app_id_2；
template_id_2关联了biz_id_1下的app_id_1、app_id_2、app_id_3；
template_id_3关联了biz_id_1下的app_id_1；

例如template_id_1为"common.yaml"、template_id_2为"server.yaml"、template_id_3为"allow_list.yaml",
则关联后在各个App（应用模块）下的配置呈现为,

```shell
app_id_1(xxxxxx) ----> common.yaml
                |----> server.yaml
                |----> allow_list.yaml

app_id_2(xxxxxx) ----> common.yaml
                |----> server.yaml

app_id_3(xxxxxx) ----> server.yaml
```

### 2.1.4 关联总结

经`2.1模块拓扑关联` `2.2项目关联` `2.3模板关联`3个章节后，已介绍了如何进行业务下的App（应用模块）、Template（配置模板）关联，以呈现出
服务模块下的配置结构。

整体的模型如下,
```shell
业务下App（应用模块）关系:

biz_id_1(xxxxxx) ----> app_id_1(xxxxxx)
                |----> app_id_2(xxxxxx)
                |----> app_id_3(xxxxxx)
                ......

biz_id_2(xxxxxx) ----> app_id_4(xxxxxx)
                |----> app_id_5(xxxxxx)
                |----> app_id_6(xxxxxx)
                ......


业务下Template（配置模板）以及关联App（应用模块）关系:

biz_id_1(xxxxxx) ----> template_id_1(xxxxxx) ----> app_id_1(xxxxxx)
                |                           |----> app_id_2(xxxxxx)
                |
                |----> template_id_2(xxxxxx) ----> app_id_1(xxxxxx)
                |                           |----> app_id_2(xxxxxx)
                |                           |----> app_id_3(xxxxxx)
                |
                |----> template_id_3(xxxxxx) ----> app_id_1(xxxxxx)
                |
                ......

biz_id_2(xxxxxx) ----> template_id_4(xxxxxx) ----> app_id_4(xxxxxx)
                |                           |----> app_id_5(xxxxxx)
                |
                |----> template_id_5(xxxxxx) ----> app_id_4(xxxxxx)
                |                           |----> app_id_5(xxxxxx)
                |                           |----> app_id_6(xxxxxx)
                |
                |----> template_id_6(xxxxxx) ----> app_id_4(xxxxxx)
                |
                ......

App（应用模块）下配置文件呈现:

biz_id_1(xxxxxx) ----> app_id_1(xxxxxx) ----> common.yaml
                                       |----> server.yaml
                                       |----> allow_list.yaml

biz_id_1(xxxxxx) ----> app_id_2(xxxxxx) ----> common.yaml
                                       |----> server.yaml

biz_id_1(xxxxxx) ----> app_id_3(xxxxxx) ----> server.yaml
```

相关接口说明,

`1.App相关接口`:
    - [创建 - create_app](../api/esb/apidocs/zh_hans/create_app.md)
    - [删除 - delete_app](../api/esb/apidocs/zh_hans/delete_app.md)

`2.Template相关接口`:
    - [创建 - create_template](../api/esb/apidocs/zh_hans/create_template.md)
    - [查询 - get_template](../api/esb/apidocs/zh_hans/get_template.md)
    - [列表 - get_template_list](../api/esb/apidocs/zh_hans/get_template_list.md)
    - [更新 - update_template](../api/esb/apidocs/zh_hans/update_template.md)
    - [渲染 - render_template](../api/esb/apidocs/zh_hans/render_template.md)
    - [删除 - delete_template](../api/esb/apidocs/zh_hans/delete_template.md)

`3.TemplateVersion相关接口`:
    - [创建 - create_template_version](../api/esb/apidocs/zh_hans/create_template_version.md)
    - [查询 - get_template_version](../api/esb/apidocs/zh_hans/get_template_version.md)
    - [列表 - get_template_version_list](../api/esb/apidocs/zh_hans/get_template_version_list.md)
    - [更新 - update_template_version](../api/esb/apidocs/zh_hans/update_template_version.md)
    - [删除 - delete_template_version](../api/esb/apidocs/zh_hans/delete_template_version.md)

`4.TemplateBind相关接口`:
    - [创建 - create_template_bind](../api/esb/apidocs/zh_hans/create_template_bind.md)
    - [查询 - get_template_bind](../api/esb/apidocs/zh_hans/get_template_bind.md)
    - [列表 - get_template_bind_list](../api/esb/apidocs/zh_hans/get_template_bind_list.md)
    - [删除 - delete_template_bind](../api/esb/apidocs/zh_hans/delete_template_bind.md)

`5.Config相关接口（不启用BSCP配置模板相关功能时需使用Config相关接口管理App下的配置）`:
    - [创建 - create_config](../api/esb/apidocs/zh_hans/create_config.md)
    - [查询 - get_config](../api/esb/apidocs/zh_hans/get_config.md)
    - [列表 - get_config_list](../api/esb/apidocs/zh_hans/get_config_list.md)
    - [更新 - update_config](../api/esb/apidocs/zh_hans/update_config.md)
    - [删除 - delete_config](../api/esb/apidocs/zh_hans/delete_config.md)



## 2.2 配置发布

本章节意在介绍配置发布流程，如何基于原子接口实现配置下发。

```shell
拓扑示例:
biz_id_1(xxxxxx) ----> app_id_1(xxxxxx) ----> common.yaml
                                       |----> server.yaml
                                       |----> allow_list.yaml

操作示例:
`提交` ----> `确认` ----> `创建版本` ----> `发布版本` ----> `reload版本` ----> `回滚版本，若需要` ----> `查询结果`


接口调用示例:
   `1.create_multi_commit_with_content（创建混合提交内容）` ----> `2.confirm_multi_commit（确认提交）`
                                                            |
                                                            |
                    `3.create_strategy（创建策略，若需要）` ----> `4.create_multi_release（创建混合版本）`
                                                            |
                                                            |
`5.publish_multi_release（发布混合版本）` ----> `6.reload_multi_release（reload混合版本）` ----> `7.rollback_multi_release（回滚混合版本, 若需要）`
                                                            |
                                                            |
                                    `8.get_effected_app_instance_list（查询生效节点列表）`
```

### 2.2.1 改动提交

```shell
`1.create_multi_commit_with_content（创建混合提交内容）` ----> 2.confirm_multi_commit（确认提交）` ...
```

`相关接口`:
    - [创建混合提交 - create_multi_commit_with_content](../api/esb/apidocs/zh_hans/create_multi_commit_with_content.md)
    - [确认混合提交 - confirm_multi_commit](../api/esb/apidocs/zh_hans/confirm_multi_commit.md)

### 2.2.2 创建策略

```shell
... `3.create_strategy（创建策略，若需要）` ...
```

`相关接口`:
    - [创建策略 - create_strategy](../api/esb/apidocs/zh_hans/create_strategy.md)

### 2.2.3 发布版本

```shell
... `4.create_multi_release（创建混合版本）` ----> `5.publish_multi_release（发布混合版本）` ----> `6.reload_multi_release（reload混合版本）` ...
```

`相关接口`:
    - [创建混合版本 - create_multi_release](../api/esb/apidocs/zh_hans/create_multi_release.md)
    - [发布混合版本 - publish_multi_release](../api/esb/apidocs/zh_hans/publish_multi_release.md)
    - [reload混合版本 - reload_multi_release](../api/esb/apidocs/zh_hans/reload_multi_release.md)

### 2.2.4 回滚版本

```shell
... `7.rollback_multi_release（回滚混合版本, 若需要）` ...
```

`相关接口`:
    - [回滚混合版本 - rollback_multi_release](../api/esb/apidocs/zh_hans/rollback_multi_release.md)

### 2.2.5 查询结果

```shell
... `8.get_effected_app_instance_list（查询生效节点列表）`
```

`相关接口`:
    - [查询生效信息 - get_effected_app_instance_list](../api/esb/apidocs/zh_hans/get_effected_app_instance_list.md)

# 3.其他说明

# 4.Q&A
