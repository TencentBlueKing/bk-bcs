# 配置模板语法

> 基于GO语言内建模板引擎规则(<https://golang.org/pkg/text/template/>)。

*单一变量替换*

```shell
Hello {{ .Var }} !   -> rule.vars(World)   =   Hello World!
```

*多变量替换*

```shell
Here is what they said, {{ range .Vars }} {{ . }} {{end}}!   -> rule.vars(["unity", "practical", "highly effective", "enterprising"])   =  Here is what they said, unity practical highly effective enterprising!
```

# 配置模板规则定义

```json
    [
        {
            "type": 0,
            "name": "cluster1",
            "vars": {
                "k1": "v1a",
                "k2": 0,
                "k3": ["v3a", "v3b"]
            }
        },
        {
            "type": 1,
            "name": "zone1",
            "vars": {
                "k1": "v1b",
                "k2": 1,
                "k3": ["v3c", "v3d"]
            }
        }
    ]
```

# Example

*模板文件 template.tpl*

```yaml
# single values
k1: {{ .k1 }}
k2: {{ .k2 }}

# array values
{{ range .k3 }}
k3:
    - {{ . }}
{{end}}
```

*模板规则 rules*

```json
    [
        {
            "type": 0,
            "name": "cluster1",
            "vars": {
                "k1": "v1a",
                "k2": 0,
                "k3": ["v3a", "v3b"]
            }
        },
        {
            "type": 1,
            "name": "zone1",
            "vars": {
                "k1": "v1b",
                "k2": 1,
                "k3": ["v3c", "v3d"]
            }
        }
    ]
```

*模板渲染结果*

**cluster1 渲染结果:**

```yaml
# single values
k1: v1a
k2: 0

# array values
k3:
    - v3a
    - v3b
```

**zone1 渲染结果:**

```yaml
# single values
k1: v1b
k2: 1

# array values
k3:
    - v3c
    - v3d
```

# 模板渲染接口

## Render Interface(渲染接口)
> 模板渲染接口，根据rule规则渲染指定应用的某个配置集合。

* 支持灵活规则配置，指定cluster区域或指定zone区域；
* 当指定cluster区域与指定zone区域重叠时将分别进行渲染，既指定zone的实例会生效zone级别的配置，指定cluster的实例则生效cluster级别的配置内容;
* 模板中cluster、zone使用集群大区的名称做索引，既不同应用相同的集群大区索引可共用;
* 进行渲染时若指定cluster或zone不存在，则无法进行渲染(可提前使用预览接口验证后提交)。

## PreviewRendering Interface(渲染预览接口)
> 模板渲染预览接口，提供渲染结果的预览功能和集群大区信息验证能力。

* 将指定模板和规则进行渲染，返回渲染后的配置内容结果；
* 预览会验证模板规则中的集群大区信息，验证是否存在是否可渲染;
* 预览完成后确认无误则可确认进行正式的渲染操作。
