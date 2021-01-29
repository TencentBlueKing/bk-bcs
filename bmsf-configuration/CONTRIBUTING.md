# 开发指南

开发BSCP项目需先阅读了解项目规范，明确项目的模块实现方式、配置规范、包引用方式、代码规范、目录结构、PB协议设计规范等,

[项目规范](docs/standard.md)

### 分支与Issues

请确保关键性的问题都关联其对应的Issues，并确保该Issues是否已经存在, 尽量将相同话题在一个Issue处理或进行关联，不要冗余创建Issues话题。

* 关键特性需要关联Issue;
* 代码提交需在单独的特性分支不要在master分支提交;

### Commit信息格式

Each commit message should include a **type**, a **scope** and a **subject**:

Commit信息需包含**type**, **scope**（若需要）, **subject**:

```
 <type>(<scope>): <subject>
```

Commit信息不要超过100个字符，尽量保证内容的可读性, 比较好的提交示例如下,

```
 #8 FEAT(template): add new mongodb template // 该提交说明新增了一个mongodb模板文件支持，type为'FEAT', scope为'template', subject描述了具体改动
 #7 FIX(dockerfile): fix an issue with source image in Dockerfile // 该提交说明修复了一个镜像问题，type为'FIX', scope为'dockerfile', subject描述了具体改动
 #6 DOCS(project): update available templates section // 该提交说明更新了项目文档，type为'DOCS', scope为'project', subject描述了具体改动
```

#### Commit Type类型说明

Commit Type必须是下面类型中的一个:

* **FEAT**: 新特性
* **FIX**: 修复问题
* **OPT**: 优化
* **DOCS**: 仅仅文档更新
* **STYLE**: 修改不影响代码逻辑的格式问题
* **REFACTOR**: 既不是新特性也不是修复问题的重构
* **TEST**: 更新测试相关
* **CHORE**: 修改构建、部署等相关内容

#### Commit Scope说明

Scope为可选，该信息需简短说明改动提交的关联内容， 如`dockerfile`, `project`, `etc`...

#### Commit Subject说明

Subject需简短准确描述改动内容:

* 准确适用时态语义: "change" not "changed" nor "changes"
* 首单次字母不要大写
* 结尾不要使用中英文句号，读个简短描述可以使用;分隔，但不建议单个Commit信息提交过多
