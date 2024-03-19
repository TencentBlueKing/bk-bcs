### 描述
该接口提供版本：v1.0.0+
 

通过当前未命名版本生成一个新的版本并且发布。

### 输入参数
| 参数名称            | 参数类型     | 必选   | 描述             |
| ----------------- | ------------ | ------ | ---------------- |
| biz_id            | uint32       | 是     | 业务ID          |
| app_id            | uint32       | 是     | 应用ID          |
| release_name      | string       | 是     | 生成的版本名称    |
| release_memo      | string       | 否     | 生成的版本说明    |
| all               | bool         | 是     | 是否全量发布      |
| gray_publish_mode | string       | 否     | 灰度发布模式，仅在 all 为 false 时有效，枚举值：publish_by_labels,publish_by_groups      |
| groups            | array        | 否     | 要发布的分组名称列表，仅在 gray_publish_mode 为 publish_by_groups 时生效      |
| labels            | []label     | 否     | 要发布的标签列表，仅在 gray_publish_mode 为 publish_by_labels 时生效      |
| group_name        | string       | 否     | 在 gray_publish_mode 为 publish_by_labels 时生效，用于根据 labels 生成一个分组时对其命名，如果有服务有可用的（绑定了服务）同 labels 的分组存在，则复用旧的分组，不会新创建分组       |
| variables         | []variable   | 否     | 渲染模版时用的模版变量，服务没有引用配置模版时可不设置；如果引用了配置模版，模版中有使用到变量，但没有设置该参数，则使用业务下变量管理中的，如果也没有，则报错 |

#### label
| 参数名称    | 参数类型 | 必选 | 描述                                                         |
| ----------- | ----------- | ---- | ------------------------------------------------------------ |
| key         | string      | 是   | 标签的key。最大长度128个字符，仅允许使用中文、英文、数字、下划线、中划线，且必须以中文、英文、数字开头和结尾 |
| op          | string      | 是   | 标签的运算符（枚举值：eq,ne,gt,ge,lt,le,in,nin）            |
| value       | interface   | 是   | 根据 op 的类型，对应不同的类型，详见说明|

##### label 说明：
```json
label包含了3个元素key,op,value。其中key,value分别为一个label的key与value的值；op为该label的key与value的运算方式，目前
支持的运算符(op)为: eq(等于),ne(不等于),gt(大于),ge(大于等于),lt(小于),le(小于等于),in(包含),nin(不包含）。其中lable的value的
值的类型与运算符(op)有关系，不同的op对应不同的value的类型。具体如下：
  2.1. op为eq,ne时，value的值为string;
  2.2. op为gt,ge,lt,le时，value的值为数值类型;
  2.3. op为in,nin时，value的值为字符串数组类型;
  2.4 value为字符串类型时，最大长度为128;
{
	"labels": [{
			"key": "name",
			"op": "eq",
			"value": "lol"
		},
		{
			"key": "set",
			"op": "in",
			"value": ["1", "2", "3"]
		}
	]
}
```

#### variable
| 参数名称    | 参数类型 | 必选 | 描述                                                         |
| ----------- | -------- | ---- | ------------------------------------------------------------ |
| name        | string   | 是   | 模版变量名称。最大长度128个字符，仅允许使用中文、英文、数字、下划线、中划线，且必须以中文、英文、数字开头和结尾 |
| type        | string   | 是   | 模版变量类型（枚举值：string、number）                       |
| default_val | string   | 是   | 模版变量默认值                                               |
| memo        | string   | 否   | 模版变量描述。最大长度256个字符，仅允许使用中文、英文、数字、下划线、中划线、空格，且必须以中文、英文、数字开头和结尾 |

### 调用示例
```json
```

### 响应示例
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1,
    "have_credentials": false
  }
}
```

### 响应参数说明
| 参数名称     | 参数类型   | 描述                           |
| ------------ | ---------- | ------------------------------ |
|      code        |      int32      |            错误码                   |
|      message        |      string      |         请求信息                  |
|       data       |      object      |            响应数据                  |

#### data
| 参数名称     | 参数类型   | 描述                           |
| ------------ | ---------- | ------------------------------ |
|      id        |      uint32      |            发布策略历史ID                    |
| have_credentials        |      bool      |      服务是否有绑定可用的密钥  |
