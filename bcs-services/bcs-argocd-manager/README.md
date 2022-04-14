# 编译
> *要求：*
> - golang >= 1.16
> - GOPATH，GOBIN 配置正确

1. `make tools`
   - 安装一系列代码生成工具链
2. `make proto`
   - 根据 pkg/sdk/xxx 下的 proto 文件生成对应的 xxx.pb.go，xxx.pb.micro.go, xxx.pb.gw.go
   - 根据 pkg/apis 下的资源 type 生成对应的 proto message，生成的文件为 pkg/apis/tkex/v1alpha1 目录下的 generated.proto,generated.pb.go
3. `make client`
   - 根据 pkg/apis 下的资源 type 生成对应 k8s clientset、informer、lister，生成的文件位于 pkg/client 目录下
4. `make crds`
   - 根据 pkg/apis 下的资源 type 生成对应的 CRD，生成的文件位于 crds 目录下
5. `make build-all`
   - 编译所有组件
> *注意：*若生成失败或生成过程被终端，generated 文件夹为生成器中间产物，可以手动删除
