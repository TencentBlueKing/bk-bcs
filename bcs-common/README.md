# bcs-common

bcs 公共模块库，提供日志，metrics, i18n, 双栈监听, 服务注册,

<a href="https://pkg.go.dev/github.com/Tencent/bk-bcs/bcs-common/common><img src="https://pkg.go.dev/badge/github.com/Tencent/bk-bcs/bcs-common" alt="GoDoc"></a>

## common 目录
- [blog](./common/blog/) 日志相关
- [encrypt](./common/encrypt/) 加密相关
- [version](./common/version/) 编译版本号

## pkg 目录
- [otel](./pkg/otel/) metrics相关
- [i18n](./pkg/i18n/) 国际化相关
- [audit](./pkg//audit/) 操作审计相关
- [auth](./pkg/auth/) iam 鉴权相关

## 使用方式
根据依赖的库, 使用go get进行依赖
```
go get github.com/Tencent/bk-bcs/bcs-common
```

## 独立 go.mod 模块
- [common/encryptv2](./common/encryptv2/) 国密，需要开启cgo
- [pkg/bcsapiv4](./pkg/bcsapiv4/) 依赖运行时
