PB协议开发须知
==============

开发PB协议需要阅读并遵守一下事项!


## PB协议格式规范

* [grpc-gateway protobuf style guide](https://buf.build/docs/style-guide/#files-and-packages)

最基本的，`message`需要`大驼峰`, `字段`需要`下划线`, 编译后生成的*pb.go文件中都会转为Golang规范的大驼峰。
message内部嵌套的message定义以及枚举定义会变为大写和下划线的Golang变量，这个也是PB内部的规则不计入代码扫描。

## 字段类型特殊约定

系统外吐接口采用grpc-gateway完成HTTP2和HTTP1.X之间的转换，在社区实现上存在uint64、int64精度转换问题，即对外的协议中若为
`64位整型`则会被转换为`string`, 故此在内部设计时需要注意，暂行规范为:

    database层涉及遵循实际精度：即该用64位就用64位，只在协议层外吐时做转换

举例，total_count字段直接使用uint32，uint32可以达到十亿级别，对于列表返回需求已经足够使用。
content_size字段暂时使用uint32, 系统内不对该字段产生读写文件依赖，只是用户自行生成和消费的数据，按字节为单位uint32可以
表述40GB大小，若超过该大小用户侧可以自行约定单位，例如约定单位为KB，这样足够使用。



* [关于grpc-gateway对64位整型转换的讨论](https://github.com/grpc-ecosystem/grpc-gateway/issues/438)
* [protobuf关于64位整型转换的说明](https://developers.google.com/protocol-buffers/docs/proto3#json)

## 外吐协议

外吐协议需要基于grpc-gateway，杜绝私自使用其他HTTP框架外吐, 且外吐协议需要严格编写结构体以及字段的注释描述信息。

## 协议开发

- 协议位置：系统内部协议均需放到protocol中维护，严格杜绝肆意在其他地方定义协议，不利于项目代码维护;
- 协议复用：协议上尽量复用，不要因为一个两个字段就重新定义，造成整个项目中协议冗余混乱, 差异较大的需单独设计单独开接口;

## PB协议序列化问题

默认PB生成的*pb.go源文件中字段为`omitempty`, 故此在序列化时需要注意空字段问题，约定内部对PB结构序列化外吐时需要使用
`bk-bscp/pkg/json`下的疯转进行，保证序列化外吐后的json不缺失空字段和默认值字段。

* [关于空值字段序列化的讨论](https://github.com/grpc-ecosystem/grpc-gateway/issues/233)
