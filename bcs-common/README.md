# bcs-common

bcs 公共模块库，提供日志，metrics, i18n, 双栈监听, 服务注册, ESB接口封装等公共依赖函数

## common 目录
- blog 日志相关
- encrypt 加密相关
- version 编译版本号

## pkg 目录
- otel metrics相关
- i18n 国际化相关

## 使用方式
根据依赖的库, 使用go get进行依赖
```
go get github.com/Tencent/bk-bcs/bcs-common
```

## 独立 go.mod 模块
- [common/encryptv2](./common/encryptv2/) 国密，需要开启cgo
- [pkg/bcsapiv4](./pkg/bcsapiv4/) 依赖运行时
