# 开发原则
插件开发采用 SDK 与 CLI 工具分离的形式
# SDK 开发规范
> BCS service 层各模块间通过集群内服务发现相互调用，为了便于 service 层提供的接口能够供集群外调用，因此，各模块实现 kubectl 插件时需要统一实现外部可调用的 client SDK，SDK 通过 API-Gateway 调用各模块服务

## 资源规范
根据各模块服务抽象出资源
资源有唯一标识，如ID、name

资源命名规则：
- 表意明确
- 资源使用全称，避免简写，例如使用 `Repository` 而不是 `Repo`
- 多单词资源可以拼接，如 NodeGroup
## 接口规范
推荐抽象出服务调用的接口 client，不考虑底层通信逻辑，实现接口可选采用 http 和 grpc 等方式通信
对于每一个资源，需要实现一个 client，通过 SDK 调用不同资源的 client 组成 client 集合
例如：
`poolCli := client.Pool()`
`ngCli := client.NodeGroup()`
**外部调用 SDK 示例**
```go
import rmclient "kubectl-bcs-resource-manager/pkg/client"

func main() {
    cfg := rmclient.Config{
       APIServer: "localhost:31443",
       AuthToken: "abcdefghijklmnopqrstuvwxyz",
    }
    client := rmclient.New(c)
    consumers, err := client.Consumer().List()
    pools, err := client.Pools().List()
}
```
# kubectl 插件开发规范
## 插件命名指南
krew 官方提供了[插件命名指南](https://krew.sigs.k8s.io/docs/developer-guide/develop/naming-guide/)，主要采用有以下内容：
- 使用小写字母和连字符，不要使用驼峰式命名
- 表意明确，独一无二
- 如果是供应商插件，前缀请使用供应商
- 不能包含 `kube` 前缀
  bcs 插件明明为 kubectl-bcs-模块名
  如 `kubectl-bcs-resource-manager`
## 配置管理
【强制】配置文件默认路径： `/etc/bcs/xxx.yaml`
【推荐】配置文件推荐格式： `yaml`
【推荐】配置文件默认名称：`bcs-resource-manager.yaml`
【推荐】插件提供 `--config`参数可以动态获取配置文件路径，优先级：参数 > 默认
## 第三方库：
### 命令行工具：[cobra](https://github.com/spf13/cobra)
go 常用命令行库：
1. `/pkg/flag`
2. `/spf13/pflag`
3. `/urfave/cli`
4. `/spf13/cobra`

`/pkg/flag` 和 `/spf13/pflag` 功能比较单一，仅支持简单的 flag 解析，不适用于复杂 cli 工具开发。
`cobra` 和 `cli` 是比较优秀的命令行库，在开源项目中使用较多，主要优势有：
- 完全兼容 posix 命令行模式
- 嵌套子命令 subcommand
- 支持全局，局部，串联 flags
- 自动生成 commands 和 flags 的帮助信息
- 自动生成详细的 help 信息，如 app help
- 自动识别帮助 flag -h，--help
- 自动生成应用程序在 bash 下命令自动完成功能
- 自动生成应用程序的 man 手册
- 命令行别名
- 自定义 help 和 usage 信息

`cobra` 被 Docker、Kubernetes、etcd 等知名开源项目采用，拥有更成熟，丰富的功能：
- 可以使用 cobra 生成应用程序和命令，使用 cobra create  [appname] 和 cobra add [cmdname]
- 如果命令输入错误，将提供智能建议
- 可选的与 viper apps 的紧密集成

因此，kubectl 插件采用应用广泛，功能健全的 `cobra`作为命令行工具库
### 表格展示工具：[tablewriter](https://github.com/olekukonko/tablewriter)
## 项目结构：
项目结构 layout：
```shell
├── Makefile
├── main.go
├── cmd
│   ├── create.go
│   ├── update.go
│   ├── delete.go
│   ├── get.go
│   ├── root.go
│   └── version.go
├── pkg
│    └── client
│        ├── client.go
│        ├── consumer.go
│        ├── pool.go
│        └── nodegroup.go
└── printer
     ├── consumer
     ├── pool
     └── nodegroup
```
`cmd` 包中每个文件对应一个命令
`pkg/client` 用于提供可供外部调用的 SDK，每个文件对应一个抽象资源的 client，组成一个 client 集合
`print` 提供格式化输出，对每一个资源封装一个 printer，包中每个文件对应一个资源

# kubectl 插件入参格式
## 输入参数采用类 kubectl 的声明式风格：
1. 若指定资源ID，则对特定资源操作，否则对所有资源进行操作
   `command [资源唯一标识ID]`
   举例：
    - `get resourc ID`
    - `get resource`
    - `delete resource ID`
2. 资源无唯一标识时，通过参数筛选确定资源
   举例：
    - `get subnet --region ap-nanjing --region ap-nanjing3 --vpc vpc-a1b2c3`
3. 需要输入结构化参数时，统一采用 `json` 格式，可以从文件中读取，也可以直接从 `flag` 中读取
   举例：
    - `create resource -f data.json`
    - `create resource -d '{"id":"aaa","name":"bbb"}'`
    - `update resource ID -f data.json`
## 参数描述
**`Use`**
1. 描述只占一行
2. `[ ]` 标识可选参数
3. `...` 标识可以为一个参数指定多个值
4.  `|`    标识互斥参数

举例：
- `create resource [-f file | -d data]`
- `get resource [ID] [-o wide|json|yaml]`

**`Aliases`**
对于长度大于4的资源推荐设置别名

**`Short`**
用于 help 命令展示的首行字段，对于命令/资源的简短描述

**`Long`**
用于 help 命令展示的字段，推荐每行不超过 80 字符。

**`Example`**
推荐对每一个参数给出一个示例，一行注释对应一行示例

## flag 规范
- 【强制】字符串或数字类型全写和简写一律使用小写
- 【强制】多个单词之间用中划线 `-` 拼接
- 【强制】布尔类型默认值为 false，全写使用小写，简写需大写，例如 `--all-namespace | -A`
- 【推荐】预留全局flag：**配置文件：**`--config | -c`
- 【推荐】预留：**命名空间：**`--namespace | -n`
- 【推荐】预留：**所有命名空间：**`--all-namespace | -A`

尽量减少全局flag
配置文件推荐默认路径：`/etc/bcs/`
配置文件推荐默认格式：`yaml`

# kubectl 插件输出格式
## 基本格式
输出格式符合 kubectl 风格，默认为按无边框居左表格展示，提供 `-o` 可选输出格式
```shell
ID                         NAME            CLUSTER          POOLS                      
deviceconsumer-A1b2C3d4    qqfc            BCS-K8S-12345    devicepool-AbCdEfGh                       
deviceconsumer-A1b2C3d4    nizhan          BCS-K8S-12345    devicepool-AbCdEfGh                       
deviceconsumer-A1b2C3d4    avatar          BCS-K8S-12345    devicepool-AbCdEfGh                       
deviceconsumer-A1b2C3d4    hhw             BCS-K8S-12345    devicepool-AbCdEfGh                       
deviceconsumer-A1b2C3d4    qqfc            BCS-K8S-12345    devicepool-AbCdEfGh
```
- 默认展示数据每行不要超过 80 个字符，这样可以方便窄屏幕的用户使用。
- `-o wide`: 展示更多参数，
- `-o json`: 以 json 格式展示数据
- `-o yaml`: 以 yaml 格式展示数据

## title 命名规范
表格表头使用以下划线 `_` 拼接的大写字母进行命名，例如 `CREATE_TIME`，`tablewriter` 将自动去除下划线 => `CREATE TIME`

## 版本格式
```json
{
  "Version": "1.2.4",
  "BuildTime": "2021-12-21 16:50:26",
  "GitCommit": "8d26c4d1dbfbed919b6d03fffb449ae60c67c033",
  "GoVersion": "go version go1.15.2 linux/amd64"
}
```
版本信息从 `Makefile` 中获取
```makefile
VERSION = 1.2.4
BUILDTIME = $(shell date '+%Y-%m-%d %T')
GITCOMMIT = $(shell git rev-parse HEAD)
GOVERSION = $(shell go version)

LDFLAG=-ldflags "-X 'github.com/Tencent/bk-bcs/bcs-services/bcs-cactl/cmd.BuildVersion=${VERSION}' \
 -X 'github.com/Tencent/bk-bcs/bcs-services/bcs-cactl/cmd.BuildTime=${BUILDTIME}' \
 -X 'github.com/Tencent/bk-bcs/bcs-services/bcs-cactl/cmd.GitCommit=${GITCOMMIT}' \
 -X 'github.com/Tencent/bk-bcs/bcs-services/bcs-cactl/cmd.GoVersion=${GOVERSION}'"

.PHONY: bin
bin: fmt vet
   go build -o bin/cactl ${LDFLAG} github.com/Tencent/bk-bcs/bcs-services/bcs-cactl
```

## 输出时间格式
`yyyy-MM-dd HH:mm:SS`
在 golang 中：
`2006-01-02 15:04:05`
