### 功能描述

获取内容摘要

#### 接口方法

- Path: /api/v2/file/content/biz/{biz_id}
- Method: HEAD

#### 接口参数

| 字段        |  类型     | 必选   |  描述      |
|-------------|-----------|--------|------------|
| biz_id      |  string   | 是     | 业务ID     |

##### HEADER设置

- `X-Bkapi-Request-Id`: 蓝鲸内部请求ID
- `X-Bkapi-App-Code`: 蓝鲸内部调用方AppCode
- `X-Bkapi-User-Name`: 蓝鲸内部用户名
- `X-Bkapi-File-Content-Id`: 上传内容的SHA256值

```json
curl -X HEAD -I http://localhost:8080/api/v2/file/content/biz/biz01 \
     -H "X-Bkapi-File-Content-Id:4c2d4c4231d1ff93975879226fe92250616082cbaed6a4a888d2adc490ba9b44" \
     -H "X-Bkapi-User-Name: melo" \
     -H "X-Bkapi-Request-Id: abc" \
     -H "X-Bkapi-App-Code: 1"
```

### 返回结果示例

```json
200 OK // 存在该内容资源，且返回详细资源信息（遵循HTTP语义）
```

或

```json
404 NOT FOUND // 不存在该内容资源（遵循HTTP语义）
```
