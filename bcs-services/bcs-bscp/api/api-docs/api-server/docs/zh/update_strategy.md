### 描述
该接口提供版本：v1.0.0+


更新策略集。

### 输入参数
| 参数名称     | 参数类型     | 必选   | 描述             |
| ------------ | ------------ | ------ | ---------------- |
| biz_id         | uint32       | 是     | 业务ID     |
| app_id         | uint32       | 是     | 应用ID     |
| id         | uint32       | 是     | 策略ID     |
| release_id         | uint32       | 否     | 版本ID     |
| as_default         | bool       | 否     | 是否作为兜底策略，如果是兜底策略，不能设置 scope 或 namespace，默认不是兜底策略。一个策略集下只能有一个兜底策略，即as_default的策略只能有一个为true    |
| name         | string       | 否     | 策略名称。最大长度128个字符，仅允许使用中文、英文、数字、下划线、中划线，且必须以中文、英文、数字开头和结尾    |
| scope         | object       | 看情况     | 发布范围，该策略所属策略集是 normal 类型，该值必填，namespace模式下不能填    |
| memo         | string       | 否     | 备注。最大长度256个字符，仅允许使用中文、英文、数字、下划线、中划线、空格，且必须以中文、英文、数字开头和结尾    |

#### scope:
| 参数名称     | 参数类型     | 必选   | 描述             |
| ------------ | ------------ | ------ | ---------------- |
| selector         | object       | 否     | 发布范围, 该对象的json序列化后的字符串大小不能超过1KB     |
| sub_strategy         | object       | 否     | 子策略    |

#### sub_strategy:
| 参数名称     | 参数类型     | 必选   | 描述             |
| ------------ | ------------ | ------ | ---------------- |
| name         | string       | 否     | 策略名称。最大长度128个字符，仅允许使用中文、英文、数字、下划线、中划线，且必须以中文、英文、数字开头和结尾    |
| release_id         | uint32       | 否     | 版本ID     |
| selector         | object       | 否     | 发布范围, 该对象的json序列化后的字符串大小不能超过1KB     |
| memo         | string       | 否     | 备注。最大长度256个字符，仅允许使用中文、英文、数字、下划线、中划线、空格，且必须以中文、英文、数字开头和结尾    |

#### selector:
| 参数名称     | 参数类型     | 必选   | 描述             |
| ------------ | ------------ | ------ | ---------------- |
| match_all         | bool       | 否     | 是否是全量发布，如果是全量发布，labels_or 和 labels_and无效且禁止设置    |
| labels_or         | object       | 否     | 实例label的匹配规则为or，且label最多设置5个     |
| labels_and         | object       | 否     | 实例label的匹配规则为and，且label最多设置5个 |
注：labels_or 和 labels_and 同一个策略只能设置一个，不能同时使用labels_or 和 labels_and。

##### labels_or/labels_and说明：
```json
1. labels包含了期望的节点实例标签逻辑或集合, 该维度支持多个标签，每个标签之间为逻辑与的关系, labels_or与labels_and之间为或的关系。
2. 每个label包含了3个元素key,op,value。其中key,value分别为一个label的key与value的值；op为该label的key与value的运算方式，目前
支持的运算符(op)为: eq(等于),ne(不等于),gt(大于),ge(大于等于),lt(小于),le(小于等于),in(包含),nin(不包含）。其中lable的value的
值的类型与运算符(op)有关系，不同的op对应不同的value的类型。具体如下：
  2.1. op为eq,ne时，value的值为string;
  2.2. op为gt,ge,lt,le时，value的值为数值类型;
  2.3. op为in,nin时，value的值为字符串数组类型;
  2.4 value为字符串类型时，最大长度为128;
{
    "labels_or": [
        {
	    "key": "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/ns/biz",
            "op": "eq",
            "value": "lol"
	},
	{
	    "key": "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/ns/group_id",
            "op": "in",
            "value": ["1", "2", "3"]
	}
    ]
}
```

### 调用示例
```json
{
    "name": "strategy",
    "release_id": 1,
    "scope": {
	"selector": {
	    "labels_or": [
	        {
		    "key": "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/ns/biz",
		    "op": "eq",
		    "value": "lol"
	        },
	        {
		    "key": "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/ns/group_id",
		    "op": "in",
		    "value": ["1", "2", "3"]
		}
	    ]
        },
        "sub_strategy": {
            "name": "sub_strategy",
            "release_id": 2,
            "selector": {
	        "labels_or": [
	        {
		    "key": "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/ns/biz",
		    "op": "eq",
		    "value": "lol"
	        },
	        {
		    "key": "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/ns/group_id",
		    "op": "in",
		    "value": ["1", "2", "3"]
		}
	        ]},
            "memo": "my_sub_strategy"
        }
    },
    "memo": "my_first_strategy"
}
```

### 响应示例
```json
{
    "code": 0,
    "message": "ok"
}
```

### 响应参数说明
| 参数名称     | 参数类型   | 描述                           |
| ------------ | ---------- | ------------------------------ |
|      code        |      int32      |            错误码                   |
|      message        |      string      |             请求信息                  |
