# BSCP

## 开发指南
要求 1.17 版本的 golang

编译 pb
```bash
# 下载正确的 protoc 二进制版本到 .bin 目录
make init

# 把 .bin/protoc 加到路径中
export PATH=`pwd`/.bin:$PATH

# 创建 bscp.io 软连接, 已经 gitignore 了这个文件，不会提交到 git 库
cd .. && ln -sf bcs-bscp bscp.io && cd bscp.io

# 前面的步骤一次性， OK后编译
make pb
```

编译二进制
```bash
make build_bscp
```

编译前端和UI模块
要求 1.14 版本的 NodeJS

```bash
make build_frontend
make build_ui
```
