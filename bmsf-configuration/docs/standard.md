BK-BSCP 项目规范
==========================

[TOC]

# 项目规范

## 模块实现

### 基于cobra

`注意:` 严禁随意创建修改目录结构，项目内所有服务模块需按照规范保持一致，不可私自无意义的创建子目录,
        需要单独抽象为一个组件的在modules目录内进行扩展，保持项目整体可读性!

[github spf13/cobra](https://github.com/spf13/cobra)

```
  ▾ xxxxserver/
    ▾ etc/
      - server.yaml

    ▾ cmd/
      - root.go
      - run.go
      - version.go

    ▾ service/
      - config.go
      - service.go
      - rpcs.go

    ▾ modules/
      - xxxx.go
      - xxxx_test.go

    main.go
    Makefile
    README.md
```

* etc: 配置存放目录
* cmd: 基础命令实现，root为服务根命令实现, run为服务的启动命令，完成系统前置处理并启动真正的逻辑模块, version为服务的版本信息管理
* service: 服务主体实现, 基于modules内的各种实现提供服务
* modules: 服务模块，实现各类服务所需的逻辑模块
* main.go: 调用根命令启动服务
* Makefile: 编译
* README.md: 服务说明

## 服务配置模块

### 基于viper

[github spf13/viper](https://github.com/spf13/viper)


``` go
// viper demo
viper := viper.New()
or
viper := viper.GetViper()

viper.SetConfigFile("server.yaml")
viper.ReadInConfig()

viper.WatchConfig()

viper.Get("name")
```

## BIN文件命名约定

* bin文件名称约定, projectname-modulename-servername

> eg: bk-bscp-apiserver, 即蓝鲸服务配置平台API网关服务, bk为蓝鲸、bscp为模块系统、apiserver为模块系统内服务名

## 包名引用

```
import (
    "system-package"

    "3rd-package"

    "inner-package"
)
```

> 包名引用顺序: 系统包名 -->> 第三方包名(github.com) -->> 内部自建包名

## 测试用例

### 模块测试用例

```
  ▾ xxxxserver/
    ▾ modules/
      xxxx_test.go
      xxxx_test.go
      xxxx_test.go
      ......
```

* 在每个服务下的modules目录中，实现xxxx_test.go进行对应模块的测试

### 系统测试用例

```
  ▾ test/
      xxxx_test.go
      xxxx_test.go
      xxxx_test.go
      ......
```

* 项目根目录下的test中实现系统整体的测试用例，针对主要的流程进行验证

## 部分命名约束

### 文件、目录命名

* 目录命名以横线(-)形式，如bk-bscp-xxxx;
* 文件命名以下划线(_)形式，如xxxx_xxxx.go、init_shell.sh;

### 常量命名

* 局部常量定义以驼峰模式，遵循Golang规范，如 defaultTimeout、defaultMaxRetryTimes;
* 全局公共常量定义以BSCP开头且全部大写(由早期代码检查工具规则而来，目前延续该约定), 如BSCPIDLIMITLEN；

### 数据库结构命名

* database命名以bscp开头，便于数据库统一授权管理，其中系统默认管理DB为bscpdb不可修改, 其他业务数据ShardingDB以下划线形式命名，如bscp_default、bscp_agame、bscp_bgame;
* table命名以小写t开头，以下划线形式进行组织，如t_system、t_config、t_release;
* 字段命名以大写F开头，以下划线形式进行组织, 如Fid、Fstate、Fname;
* 索引命名以小写idx或uidx开头，普通索引、联合索引以idx开头, 唯一索引以uidx开头，以下划线形式进行组织，如idx_name、idx_appid_name、uidx_name、uidx_appid_name;

### 日志规范

* 默认日志使用V(2)进行打印，默认部署日志界别为3；
* 日志中需要包含请求的SEQ（RID），便于定位指定请求；
* 日志头部需要包含模块（接口）标识，例如RPC方法名、内部模块名，e.g. 'CreateApp' 、'Filter', 便于排查问题时按照标识同意过滤筛查;

## 代码规范

* [go coding style](https://github.com/golang/go/wiki/CodeReviewComments)

## 项目结构

* [go standard project layout](https://github.com/golang-standards/project-layout)

## Protobuf协议规范

* [grpc-gateway protobuf style guide](https://buf.build/docs/style-guide/#files-and-packages)

# Q&A
