### 描述
该接口提供版本：v1.0.0+
 

创建分组

### 输入参数
| 参数名称     | 参数类型     | 必选   | 描述             |
| ------------ | ------------ | ------ | ---------------- |
| biz_id       | uint32       | 是     | 业务ID     |
| name         | string       | 是     | 应用名称。最大长度128个字符，仅允许使用中文、英文、数字、下划线、中划线，且必须以中文、英文、数字开头和结尾
| public       | bool         | 是     | 是否为公共分组    |
| bind_apps    | []uint32     | 否     | 绑定的应用ID列表 |
| mode         | string       | 是     | 分组枚举类型：custom,debug     |
| selector     | selector     | 否     | 分组选择器   |
| uid          | string       | 否     | debug UID  |

#### selector:
| 参数名称     | 参数类型     | 必选   | 描述             |
| ------------ | ------------ | ------ | ---------------- |
| labels_or    | []label       | 否     | 实例label的匹配规则为or，且label最多设置5个     |
| labels_and   | []label       | 否     | 实例label的匹配规则为and，且label最多设置5个 |
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
	"labels_or": [{
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


### 调用示例
```json
{
    "name": "广东",
    "mode": "custom",
    "public": false,
    "bind_apps": [
        1,
        2
    ],
    "selector": {
        "labels_or": [
            {
                "key": "name",
                "op": "eq",
                "value": "guangzhou"
            },
            {
                "key": "set",
                "op": "in",
                "value": [
                    "guangzhou-1",
                    "guangzhou-2",
                    "guangzhou-3"
                ]
            }
        ]
    },
    "uid": ""
}
```

### 响应示例
```json
{
    "code": 0,
    "message": "ok",
    "data": {
        "id": 1
    }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述     |
| -------- | -------- | -------- |
| data     | object   | 响应数据 |

#### data
| 参数名称     | 参数类型   | 描述                           |
| ------------ | ---------- | ------------------------------ |
|      id        |      uint32      |            分组ID                    |
