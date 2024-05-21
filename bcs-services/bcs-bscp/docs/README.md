# swagger文档生成

## 生成文档命令
```
make sg
```

## api添加扩展字段
在指定的方法中添加扩展字段，扩展字段的key必须是`x-`开头，值可以是字符串、数字、布尔等类型:
```
option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
    extensions: {
        key: "x-"
        value:{}
    }
}
```
示例：
```
  rpc Extensions(ExtensionsReq) returns (ExtensionsResp) {
    option (google.api.http) = {
      delete: "/v1/example/extensions"
    };
    option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
      extensions: {
        key: "x-grpc-gateway-baz-list";
        value {
          list_value: {
            values: {
              string_value: "one";
            }
            values: {
              bool_value: true;
            }
          }
        }
      }
    };
  }
```

## 指定api对外暴露

找到需要生成的proto文件.
在服务中添加标签只能单个：
```
option (google.api.api_visibility).restriction = "BKAPIGW";
```

在指定api中加入以下选项后就可以不对外暴露，可以添加多个标签:
```
option (google.api.method_visibility).restriction = "BKAPIGW";
```
示例：
```
service VisibilityRuleEchoService {
  option (google.api.api_visibility).restriction = "BKAPIGW";
  // Echo method receives a simple message and returns it.
  // It should always be visible in the open API output.
  rpc Echo(VisibilityRuleSimpleMessage) returns (VisibilityRuleSimpleMessage) {
    option (google.api.http) = {post: "/v1/example/echo/{id}"};
  }
  // EchoInternal is an internal API that should only be visible in the OpenAPI spec
  // if `visibility_restriction_selectors` includes "BKAPIGW".
  rpc EchoInternal(VisibilityRuleSimpleMessage) returns (VisibilityRuleSimpleMessage) {
    option (google.api.method_visibility).restriction = "BKAPIGW";
    option (google.api.http) = {get: "/v1/example/echo_internal"};
  }
  // EchoPreview is a preview API that should only be visible in the OpenAPI spec
  // if `visibility_restriction_selectors` includes "PREVIEW".
  rpc EchoPreview(VisibilityRuleSimpleMessage) returns (VisibilityRuleMessageInPreviewMethod) {
    option (google.api.method_visibility).restriction = "PREVIEW";
    option (google.api.http) = {get: "/v1/example/echo_preview"};
  }
  // EchoInternalAndPreview is a internal and preview API that should only be visible in the OpenAPI spec
  // if `visibility_restriction_selectors` includes "PREVIEW" or "BKAPIGW".
  rpc EchoInternalAndPreview(VisibilityRuleSimpleMessage) returns (VisibilityRuleSimpleMessage) {
    option (google.api.method_visibility).restriction = "BKAPIGW,PREVIEW";
    option (google.api.http) = {get: "/v1/example/echo_internal_and_preview"};
  }
}
```
也可以对参数做隐藏：
```
message VisibilityRuleEmbedded {
    int64 progress = 1;
    string note = 2;
    string internal_field = 3 [(google.api.field_visibility).restriction = "BKAPIGW"];
    string preview_field = 4 [(google.api.field_visibility).restriction = "BKAPIGW,PREVIEW"];
}
```

## 生成文档指令参数介绍
生成swagger文档时有指定一些参数:
```
--openapiv2_out                                             // 指定生成的目录 例如：openapiv2_out docs， 生成到docs目录
--openapiv2_opt preserve_rpc_order=true                     // 是否按照proto文件顺序生成
--openapiv2_opt output_format=json                          // 生成文件类型目前就 yaml和json
--openapiv2_opt include_without_visibility=true             // 是否包含不可见标签，如果设置ture那么不添加visibility_restriction_selectors的api也会输出
--openapiv2_opt visibility_restriction_selectors=BKAPIGW    // 如果设置为INTERNAL时，api以及参数是`BKAPIGW`会被显示出来
```
更多参数介绍：https://github.com/grpc-ecosystem/grpc-gateway/blob/main/protoc-gen-openapiv2/main.go

更多示例文件：https://github.com/grpc-ecosystem/grpc-gateway/blob/main/examples/internal/proto/examplepb/visibility_rule_echo_service.proto