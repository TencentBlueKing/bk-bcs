# i18N国际化-使用介绍

### 目录结构
```
.
├── cmd
|   ├── export_i18n_key.go
├── ctx.go
├── go.mod
├── i18n
│   ├── common
│   │   ├── assets.go
│   │   └── lang
│   │       ├── de.yaml
│   │       ├── en.yaml
│   │       ├── ja.yaml
│   │       ├── ru.yaml
│   │       └── zh.yaml
│   └── testdata
│       ├── assets.go
│       └── lang
│           ├── de.yaml
│           ├── en.yaml
│           ├── ja.yaml
│           ├── ru.yaml
│           └── zh.yaml
├── i18n.go
├── i18n_test.go
├── instance.go
├── manager.go
├── README.MD
└── utils
    ├── iconv
    │   └── iconv.go
    ├── ifile
    │   └── ifile.go
    ├── imap
    │   └── str_any_map.go
    ├── iregex
    │   ├── iregex_cache.go
    │   └── iregex.go
    ├── iyaml
    │   └── iyaml.go
    └── rwmutex
        └── rwmutex.go
```
### 转译文件
i18n 目录下存放对应模块的转译目录, 例如： testdata。

每个目录下需要有一个 `assets.go` 文件，只需把该文件复制到对应模块的转译目录下。

lang 目录下是具体的转译文件内容，要求是 `yaml` 格式。

建议按照标准化的国际化地区语言码进行命名，具体请参考WIKI：https://zh.wikipedia.org/wiki/ISO_639-1

例如：zh.yaml
```yaml
hello: "你好"
world: "世界"
OrderPay: "您已成功完成订单号 %d 支付，支付金额￥%.2f。"
```
可支持 `.` 连接 例如：order.pay

### 使用示例

##### 单例对象
荐使用i18n单例对象，并可自定义配置不同的单例对象，但是需要注意的是，单例对象的配置修改是全局有效的。
```go
 i18n.Instance(name...).T(context.TODO(), "{#hello} {#world}")
```
所有的转译方法中，第一个参数要求输入 `Context` 上下文变量参数，用于上下文变量的传递、转译语言的指定。 也可以传 `nil`，但是为保证程序的严谨性，我们建议您当不知道传递什么或者没有特殊要求的时候传递`context.TODO()`或者`context.Background()`来替代。

##### 独立对象
也可以模块化独立使用gi18n模块，通过i18n.New()方法创建独立的i18n对象，然后开发者自行进行管理。
```go
i18n := i18n.New()
i18n.T(context.Background(),"{#hello} {#world}")

```
##### 语言设置
设置转译语言有两种方式，一种是通过`SetLanguage`方法设置当前I18N对象统一的转译语言，另一种是通过上下文设置当前执行转译的语言。

##### SetLanguage
通过 `i18n.SetLanguage("en")`，即可设置当前转译对象的转译语言，随后使用该对象都将使用en进行转译，需要注意的是，组件的配置方法往往都不是并发安全的，该方法也同样如此，需要在程序初始化时进行设置，随后不能在运行时进行更改。

##### WithLanguage
`WithLanguage` 方法可以创建一个新的上下文变量，并临时设置您当前转译的语言，由于该方法作用于`Context`上下文，因此是并发安全的，常用于运行时转译语言设置。我们来看一个例子：

```go
ctx := i18n.WithLanguage(context.Background(), "zh")
i18n.Translate(ctx, `hello`)
```

### 常用的方法
`T` 方法

T方法为Translate方法的别名，也是大多数时候我们推荐使用的方法名称。T方法可以给定关键字名称，也可以直接给定模板内容，将会被自动转译并返回转译后的字符串内容。

**关键字转译**

关键字的转译直接给T方法传递关键字即可，例如：T(context.TODO(), "hello")、T(context.TODO(), "world")。I18N组件将会优先将给定的关键字进行转译，转译成后返回转译后的内容，否则直接展示原内容。

**关键字转译**

T方法支持模板内容转换，模板中的关键字默认使用`{{.`和`}}`标签进行包含，模板解析时将会自动替换该标签中的关键字内容，也可以使用 `i18n.SetDelimiters("{#", "}")` 方法替换规则，建议不更改。

`Tf` 方法

格式模板内容中的变量，`Tf`为`TranslateFormat`的别名，该方法支持格式化转译内容，字符串格式化语法参考标准库`fmt`包的`Sprintf`方法。

### 额外脚本

cmd目录有个遍历项目目录导出模板格式的`{{.变量}}`变量脚本，可以 `go build` 构建对应环境的执行文件。

该脚本有几个参数：

```go
  -l string
        指定导出语种, 多个英文下逗号拼接 示例：-l en,zh, 语种文件存在追加写不存在创建, 默认为空
  -o string
        输出yaml文件目录
  -p string
        项目根目录, 默认当前根目录
```



