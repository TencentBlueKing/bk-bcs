# 国际化

## 国际化使用方式

说明：这里主要是对错误message的国际化，其他message的国际化同理类似

1. 在源码中有调用golang.org/x/text/message包中的`func (p *Printer) Sprintf/Fprintf/Printf`三个方法中的任意一个
2. 首次部署项目或没有gotext命令行工具时执行`make init`，该操作会安装gotext命令行工具<br>
   （该工具能自动从源码中提取要翻译的message以及合并message，并生成对应的目录文件）
3. 执行`make i18n`，从源码中提取要翻译的message以及合并message并更新到pkg/i18n目录下的对应文件中
4. 将pkg/i18n/translations/locales/zh/messages.gotext.json文件中未翻译的内容进行补充，补充完后再次执行`make i18n`，<br>
   这次不会有翻译缺失提示，所有message均有了对应翻译，out.gotext.json和catalog.go文件也都被更新，重新编译部署项目后生效
5. http请求的header中，通过X-Bkapi-Language指定对应的语言，当前可以是zh（中文）或en（英文），不指定默认使用zh，从而获取对应语言的message

## 使用方式详解

说明：这里主要是对错误message的国际化，其他message的国际化同理类似

### 1.在源码中有调用golang.org/x/text/message包中的`func (p *Printer) Sprintf/Fprintf/Printf`三个方法中的任意一个

- 项目中的调用路径为：`errf包Errorf方法 -> localizer包Translate方法 -> message包Sprintf方法`
- 调用示例

```go
// CreateTemplateVariable create a template variable
func (s *Service) CreateTemplateVariable(ctx context.Context, req *pbcs.CreateTemplateVariableReq) (
*pbcs.CreateTemplateVariableResp, error) {
...

if !strings.HasPrefix(strings.ToLower(req.Name), constant.TemplateVariablePrefix) {
return nil, errf.Errorf(grpcKit, errf.InvalidArgument, "template variable name must start with %s",
constant.TemplateVariablePrefix)
}

...
}
```

### 2. 首次部署项目或没有gotext命令行工具时执行`make init`，该操作会安装gotext命令行工具<br>

### （该工具能自动从源码中提取要翻译的message以及合并message，并生成对应的目录文件）

具体操作可见Makefile文件：

```makefile
.PHONY: init
init:
	...
	@echo Download gotext
	go install golang.org/x/text/cmd/gotext@v0.14.0
```

### 3. 执行`make i18n`，从源码中提取要翻译的message以及合并message并更新到pkg/i18n目录下的对应文件中

主要会执行如下操作：

- 从源码中自动提取到要翻译的message到pkg/i18n/translations/locales/{language}/out.gotext.json文件，<br>
  源码中任何调用了golang.org/x/text/message包中的`func (p *Printer) Sprintf/Fprintf/Printf`三个方法中任意一个的message均会被提取
- 合并pkg/i18n/translations/locales/{language}下的out.gotext.json和messages.gotext.json文件，并更新到out.gotext.json<br>
- 自动更新所有语言的翻译内容映射关系到pkg/i18n/translations/locales/catalog.go文件
- 复制pkg/i18n/translations/locales/{language}下的out.gotext.json到messages.gotext.json文件，便于下次修改补充翻译内容
- **注意：如果没有messages.gotext.json文件，out.gotext.json将只是提取后的内容，之前补充的翻译后内容会丢失，<br>
  因为messages.gotext.json文件的存在，重复执行`make i18n`，已翻译过的内容会依然存在**

具体操作可见Makefile文件：

```makefile
.PHONY: i18n
i18n:
	@go generate ./pkg/i18n/translations/translations.go
	@cp ./pkg/i18n/translations/locales/zh/out.gotext.json ./pkg/i18n/translations/locales/zh/messages.gotext.json
```

关于gotext命令行工具的使用见pkg/i18n/translations/translations.go文件中的注释说明：

```go
/*
* 通过gotext命令行工具，能够自动从源码中提取要翻译的message以及合并message，并生成对应的目录文件
* -srclang 指定在源码代中使用的语言，这里是英语en, 语言名称需符合BCP 47规范https://en.wikipedia.org/wiki/IETF_language_tag
* update 该子命令用来从源码中提取要翻译的message以及进行合并操作，并生成对应的目录文件
* -out 指定要生成的message catalog文件，主要是存放不同语言翻译前后的映射关系
* -lang 指定要翻译的目标语言，这里是英语en和中文zh，多个语言之间以逗号分隔
* 最后的参数表示要提取翻译message的包路径，多个包路径以空格分隔
*
* gotext命令行工具更多使用，见gotext help输出
* gotext源码使用example：https://cs.opensource.google/go/x/text/+/refs/tags/v0.14.0:cmd/gotext/examples/
 */
//go:generate gotext -srclang=en update -out=catalog.go -lang=en,zh bscp.io/cmd/... bscp.io/pkg/... bscp.io/test/...
```

执行操作后的示例, 如果存在没有翻译的message，输出会提示哪些message缺少对应的翻译：

```shell
$ make i18n
zh: Missing entry for "template variable name must start with {TemplateVariablePrefix}".
zh: Missing entry for "db operation failed".
zh: Missing entry for "invalid argument".
zh: Missing entry for "id should not be set".
zh: Missing entry for "spec not set".
zh: Missing entry for "attachment not set".
zh: Missing entry for "revision not set".
zh: Missing entry for "invalid name, length should >= 9 and must start with prefix bk_bscp_ (ignore case)".
zh: Missing entry for "invalid name, length should <= 128".
zh: Missing entry for "invalid name: {Name}, only allows to include english、numbers、underscore (_), and must start with prefix bk_bscp_ (ignore case)".
zh: Missing entry for "default_val {DefaultVal} is not a number type".
zh: Missing entry for "unsupported variable type: {VariableType}".
```

查看./pkg/i18n/translations/locales/zh/messages.gotext.json文件，能看到translation字段为空，等待去填充翻译：

```json
{
  "language": "zh",
  "messages": [
    {
      "id": "template variable name must start with {TemplateVariablePrefix}",
      "message": "template variable name must start with {TemplateVariablePrefix}",
      "translation": "",
      "placeholders": [
        {
          "id": "TemplateVariablePrefix",
          "string": "%[1]s",
          "type": "string",
          "underlyingType": "string",
          "argNum": 1,
          "expr": "constant.TemplateVariablePrefix"
        }
      ]
    },
    {
      "id": "db operation failed",
      "message": "db operation failed",
      "translation": ""
    },
    {
      "id": "invalid name: {Name}, only allows to include english、numbers、underscore (_), and must start with prefix bk_bscp_ (ignore case)",
      "message": "invalid name: {Name}, only allows to include english、numbers、underscore (_), and must start with prefix bk_bscp_ (ignore case)",
      "translation": "",
      "placeholders": [
        {
          "id": "Name",
          "string": "%[1]s",
          "type": "string",
          "underlyingType": "string",
          "argNum": 1,
          "expr": "name"
        }
      ]
    }
  ]
}
```

### 4. 将pkg/i18n/translations/locales/zh/messages.gotext.json文件中未翻译的内容进行补充，补充完后再次执行`make i18n`，<br>

### 这次不会有翻译缺失提示，所有message均有了对应翻译，out.gotext.json和catalog.go文件也都被更新，重新编译部署项目后生效

**
说明：因为源码语言用的en，翻译后的message和源码相同，会自动更新生成out.gotext.json和messages.gotext.json文件，<br>
所以pkg/i18n/translations/locales/en目录不需要做任何调整**

为translation字段增加翻译内容后的示例：

```json
{
  "language": "zh",
  "messages": [
    {
      "id": "template variable name must start with {TemplateVariablePrefix}",
      "message": "template variable name must start with {TemplateVariablePrefix}",
      "translation": "模版变量名必须以{TemplateVariablePrefix}前缀开头",
      "placeholders": [
        {
          "id": "TemplateVariablePrefix",
          "string": "%[1]s",
          "type": "string",
          "underlyingType": "string",
          "argNum": 1,
          "expr": "constant.TemplateVariablePrefix"
        }
      ]
    },
    {
      "id": "db operation failed",
      "message": "db operation failed",
      "translation": "db操作失败"
    },
    {
      "id": "invalid name: {Name}, only allows to include english、numbers、underscore (_), and must start with prefix bk_bscp_ (ignore case)",
      "message": "invalid name: {Name}, only allows to include english、numbers、underscore (_), and must start with prefix bk_bscp_ (ignore case)",
      "translation": "无效名称：{Name}，只允许英文、数字、下划线（_），且必须以bk_bscp_前缀开头（忽略大小写）",
      "placeholders": [
        {
          "id": "Name",
          "string": "%[1]s",
          "type": "string",
          "underlyingType": "string",
          "argNum": 1,
          "expr": "name"
        }
      ]
    }
  ]
}
```

再次执行`make i18n` ，这时所有message均有了对应翻译，out.gotext.json和catalog.go文件也都被更新 <br>
重新编译部署项目后生效

### 5. http请求的header中，通过X-Bkapi-Language指定对应的语言，当前可以是zh（中文）或en（英文），不指定默认使用zh，从而获取对应语言的message

翻译后的请求结果错误示例

```json
{
  "error": {
    "code": "INVALID_ARGUMENT",
    "message": "模版变量名必须以bk_bscp_前缀开头",
    "data": null,
    "details": []
  }
}
```

```json
{
  "error": {
    "code": "INVALID_ARGUMENT",
    "message": "无效名称：BK_BSCP_AGE{}，只允许英文、数字、下划线（_），且必须以bk_bscp_前缀开头（忽略大小写）",
    "data": null,
    "details": []
  }
}
```

```json
{
  "error": {
    "code": "INTERNAL",
    "message": "db操作失败",
    "data": null,
    "details": []
  }
}
```