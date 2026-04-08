## ADDED Requirements

### Requirement: drplan-gen CLI tool

提供一个独立 CLI 工具，从 Helm 渲染后的 YAML 自动生成 DRPlan 编排文件。

#### Scenario: 基本用法 - 从文件输入

- **WHEN** 用户执行 `drplan-gen --name my-app --namespace default -f rendered.yaml`
- **THEN** 工具读取 rendered.yaml，解析所有 K8s 资源，在当前目录生成 drplan.yaml、workflow-*.yaml、drplanexecution-*.yaml

#### Scenario: 基本用法 - 从 stdin 管道输入

- **WHEN** 用户执行 `helm template my-app ./chart | drplan-gen --name my-app --namespace default`
- **THEN** 工具从 stdin 读取 YAML，生成同样的输出文件

#### Scenario: 指定输出目录

- **WHEN** 用户执行 `drplan-gen --name my-app -f rendered.yaml -o ./output/`
- **THEN** 所有生成文件写入 ./output/ 目录（目录不存在则自动创建）

#### Scenario: Hook 资源识别 - pre-install

- **WHEN** 输入 YAML 中包含 annotation `helm.sh/hook: pre-install` 的 Job 资源
- **THEN** 生成独立的 pre-install Stage，其 Workflow 包含对应的 Subscription Action（hook Job 作为 feed），`--wait` 时设置 `waitReady: true`

#### Scenario: Hook 资源识别 - post-install

- **WHEN** 输入 YAML 中包含 annotation `helm.sh/hook: post-install` 的 Job 资源
- **THEN** 生成独立的 post-install Stage（dependsOn: [install]），其 Workflow 包含对应的 Subscription Action（hook Job 作为 feed）

#### Scenario: Hook 权重排序

- **WHEN** 同一 hook 类型内有多个资源，annotation `helm.sh/hook-weight` 值分别为 "-5" 和 "0"
- **THEN** weight=-5 的资源排在 Workflow Actions 列表前面，weight=0 的排在后面

#### Scenario: Hook 清理策略

- **WHEN** Hook 资源有 annotation `helm.sh/hook-delete-policy: hook-succeeded`
- **THEN** 分类结果记录 deletePolicy，用户可后续在子集群 Job 上配置 TTL

#### Scenario: 主资源聚合为 Subscription feeds

- **WHEN** 输入 YAML 中包含无 hook annotation 的 ConfigMap、Secret、Deployment、Service
- **THEN** 所有这些资源聚合到 install Stage 的 Subscription Action feeds 列表中

#### Scenario: Stage DAG 生成

- **WHEN** 输入同时包含 pre-install hook、主资源、post-install hook
- **THEN** 生成的 DRPlan stages 为：pre-install（无 dependsOn）→ install（dependsOn: [pre-install]）→ post-install（dependsOn: [install]）

#### Scenario: 无 hook 资源的输入

- **WHEN** 输入 YAML 中没有任何 hook annotation
- **THEN** 只生成 install Stage（无 dependsOn），Subscription feeds 包含全部资源

#### Scenario: --wait 参数

- **WHEN** 用户使用 `--wait` 参数
- **THEN** 所有生成的 Subscription Action（包括 hook 和主资源）都设置 `waitReady: true`

#### Scenario: 缺少必要参数

- **WHEN** 用户未提供 `--name` 参数
- **THEN** 工具输出错误信息并退出，提示 `--name` 是必需参数

#### Scenario: 空输入

- **WHEN** 输入 YAML 为空或不包含任何有效 K8s 资源
- **THEN** 工具输出错误信息并退出，提示未找到可解析的资源
