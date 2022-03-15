# 代码生成工具链
> 依赖：
> 生成器用到了 vendor 中的一些依赖，需要提前 go mod vendor -e 下载依赖到 vendor 目录
> - protoc-gen-go v1.3.2 (不能高于 v1.3.5)
> - protoc-gen-micro v4.6.0
### `make proto` 
1. 根据 `pkg/apis` 下的资源 type 生成对应的 proto message，生成的文件为 `pkg/apis/tkex/v1alpha1` 目录下的 `generated.proto`,`generated.pb.go`
2. 根据 `pkg/sdk/xxx` 下的 proto 文件生成对应的 `xxx.pb.go`，`xxx.pb.micro.go`
### `make client` 
1. 根据 `pkg/apis` 下的资源 type 生成对应 k8s clientset、informer、lister，生成的文件位于 `pkg/client` 目录下