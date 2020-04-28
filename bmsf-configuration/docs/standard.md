BK-BSCP 项目规范
==========================

[TOC]

# 项目规范

## 模块实现

### 基于cobra

[github spf13/cobra](https://github.com/spf13/cobra)

```
  ▾ xxxxserver/
    ▾ bin/
    ▾ etc/
    ▾ cmd/
      root.go
      run.go
      version.go
    ▾ service/
      service.go
    ▾ modules/
      xxxx.go
      xxxx_test.go

    main.go
    Makefile
    README.md
```

* bin: 二进制生成与运行目录
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

viper.SetConfigFile("server.yaml")
viper.ReadInConfig()

viper.WatchConfig()

viper.Get("name")
```

## BIN文件命名约定

* bin文件名称约定, projectname-modulename-servername

> eg: bk-bscp-gateway, 即蓝鲸服务配置平台网关服务

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

## 代码规范

* [go coding style](https://github.com/golang/go/wiki/CodeReviewComments)

## 项目结构

* [go standard project layout](https://github.com/golang-standards/project-layout)

# Q&A
