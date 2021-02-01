BK-BSCP 编译说明
===================

# 编译平台

- `Linux`: CentOS/ubuntu/suse等发行版, 支持编译和运行;
- `Darwin`: macOS, 支持编译和简单调试;

# 依赖命令

## golint代码检查命令

安装:

```shell
go get -u golang.org/x/lint/golint
REAL_GOLINT=$(go list -f {{.Target}} golang.org/x/lint/golint)
cp -rf ${REAL_GOLINT} /usr/bin/
```

## sed(gsed)命令

Linux平台默认附带了GNU的sed命令，Darwin(macOS)默认不带，需要自行安装gsed:

```shell
brew install gnu-sed
```

# 编译

- `make`: 编译全部模块;
- `make server`: 编译服务端模块;
- `make plugin`: 编译插件;
- `make lint`: 代码质量检查;
