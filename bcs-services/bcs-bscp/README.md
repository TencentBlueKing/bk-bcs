# BSCP

## 开发指南
要求 1.17 版本的 golang

编译 pb
```bash
# 下载正确的 protoc 二进制版本到 bin/proto 目录
make init

# 把 bin/protoc 加到路径中
export PATH=`pwd`/bin/proto:$PATH

# clang-format 请按系统自行安装 https://github.com/llvm/llvm-project/releases/
# protobuf 格式会自动使用 clang-format 格式化, 使用 Google 风格, 最大宽度是 120 个字符, make 会自动执行格式, 也可以使用下面命令手动执行
find pkg/protocol -type f -name "*.proto"|xargs clang-format -i --style="{BasedOnStyle: Google, ColumnLimit: 120}"

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

## vscode 插件
- [vscode-proto3](https://marketplace.visualstudio.com/items?itemName=zxh404.vscode-proto3) 支持 proto3 高亮提示，定义跳转，自动格式化

配置
```json
{
    "protoc": {
        "path": "${workspaceRoot}/bin/proto/protoc",
        "options": [
            "--proto_path=pkg/thirdparty/protobuf",
        ]
    },
    "clang-format.style": "{BasedOnStyle: Google, ColumnLimit: 120}",
}
```