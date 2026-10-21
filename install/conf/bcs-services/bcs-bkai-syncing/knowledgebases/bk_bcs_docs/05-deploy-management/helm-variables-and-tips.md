# Helm 内置变量与 Values 使用技巧

在蓝鲸容器管理平台（BCS）中使用 Helm 部署应用时，平台会自动向 Chart 注入一组系统内置变量（`__BCS__`），并提供表单与 `values.yaml` 双向同步机制，帮助用户编写跨环境通用的 Helm Chart。

---

## 一、在 Chart 中使用 BCS 系统内置变量

### 1. BCS 提供的内置 Helm 变量列表

通过 BCS 控制台或 API 部署 Helm Release 时，系统会自动向 Values 中注入以下内置变量：

| 变量名（Key） | 值示例 | 含义说明 |
| :--- | :--- | :--- |
| `SYS_JFROG_DOMAIN` | `hub.bktencent.com` | 平台配置的默认容器镜像仓库域名 |
| `SYS_CLUSTER_ID` | `BCS-K8S-40001` | 当前 Release 部署的目标集群 ID |
| `SYS_PROJECT_ID` | `7538606f025efa007f3e750477982c23` | 当前所属的 BCS 项目 ID |
| `SYS_PROJECT_CODE` | `blueking` | 当前所属的 BCS 项目英文代号 |
| `SYS_CC_APP_ID` | `2` | 当前项目绑定的蓝鲸 CMDB 业务 ID |
| `SYS_NAMESPACE` | `default` | 当前 Release 部署的目标命名空间名称 |

### 2. 在 Chart 模板中引用内置变量的推荐方式

在单 Chart 中可直接通过 `{{ .Values.__BCS__.Key }}` 或 `{{ $.Values.global.__BCS__.Key }}` 引用。为了同时兼容**包含子 Chart（Subcharts）**以及**脱离 BCS 使用原生 `helm` 命令直接安装（此时 `__BCS__` 可能为空）**的场景，推荐使用以下两种防御性写法：

#### 方式 1：使用 `default` 函数直接在模板中引用（简洁推荐）

```gotemplate
image: "{{ default "hub.bktencent.com" $.Values.global.__BCS__.SYS_JFROG_DOMAIN }}/my-org/my-app:{{ .Values.image.tag }}"
```

#### 方式 2：在 `_helpers.tpl` 中定义命名模板（适合多处复用）

在 `templates/_helpers.tpl` 中编写安全逐层判空的辅助模板：

```gotemplate
{{/*
获取 BCS 镜像仓库域名，若未注入则回退到默认值
*/}}
{{- define "bcs.imageRegistry" -}}
  {{- $default := "127.0.0.1" -}}
  {{- $values := .Values -}}
  {{- if $values -}}
    {{- $global := $values.global -}}
    {{- if $global -}}
      {{- $bcs := $global.__BCS__ -}}
      {{- if and $bcs $bcs.SYS_JFROG_DOMAIN -}}
        {{- $bcs.SYS_JFROG_DOMAIN -}}
      {{- else -}}
        {{- $default -}}
      {{- end -}}
    {{- else -}}
      {{- $default -}}
    {{- end -}}
  {{- else -}}
    {{- $default -}}
  {{- end -}}
{{- end -}}
```

在 Deployment 等资源模板中调用：

```gotemplate
image: "{{ template "bcs.imageRegistry" $ }}/my-app:{{ .Values.image.tag }}"
```

---

## 二、Helm Release 表单与 `values.yaml` 参数优先级说明

在 BCS 控制台创建或更新 Helm Release 时，页面同时支持**可视化表单（Form）**和**`values.yaml` 编辑器**两种参数配置方式：

1. **底层传递机制**：
   - 页面上选择的 `values.yaml` 内容会通过 `--values`（`-f`）参数传递给 Helm 渲染引擎（不会篡改 Chart 原包中的文件）。
   - 表单项旨在提升输入体验、规避缩进语法错误，其修改值最终与 `values.yaml` 合并渲染（字符串类型对应 `--set-string`）。
2. **双向实时同步（消除优先级困惑）**：
   - BCS 控制台实现了表单与 `values.yaml` 的**双向自动同步**。
   - 例如：当您在表单中修改了镜像 Tag 并失去焦点时，该新值会立即同步更新到下方的 `values.yaml` 编辑器中；反之，当您直接在 `values.yaml` 编辑器中修改了对应字段，表单中的输入框也会自动同步更新，确保最终生效值所见即所得。
