## BCS前端配置说明

BCS前端配置主要用于特性开关（eg: 版本差异，功能差异等），菜单显隐，文档链接和功能特性说明，可通过`*values*`文件配置

### 菜单显示和隐藏配置

可通过`*values*`下`feature_flags`配置菜单显示和隐藏, 如未开启但通过链接进入会出现`404`，配置如下：

```yaml
feature_flags:
  # 当 enabled 为 true 时，默认开启该特性，list 作用为黑名单，list 中的资源不开启该特性
  # 当 enabled 为 false 时，默认关闭该特性，list 作用为白名单，仅 list 中的资源开启该特性
  # 节点模板
  NODETEMPLATE:
    enabled: true # 开启节点模板菜单
  # 指标管理
  METRICS:
    enabled: false # 关闭指标菜单
  # 模板文件
  TEMPLATE_FILE:
    enabled: true
    list:
      - "test" # 只开启 test 项目的模板文件功能
```

菜单ID如下：

| 类型                         |
|------------------------------|
| CLUSTERRESOURCE              |
| WORKLOAD                     |
| DEPLOYMENT                   |
| STATEFULSET                  |
| DAEMONSET                    |
| JOB                          |
| CRONJOB                      |
| POD                          |
| NETWORK                      |
| INGRESS                      |
| SERVICE                      |
| ENDPOINTS                    |
| CONFIGURATION                |
| CONFIGMAP                    |
| SECRET                       |
| STORAGE                      |
| PERSISTENTVOLUME             |
| PERSISTENTVOLUMECLAIM        |
| STORAGECLASS                 |
| RBAC                         |
| SERVICEACCOUNT               |
| HORIZONTALPODAUTOSCALER      |
| CRD                          |
| CUSTOM_GAME_RESOURCE              |
| GAMEDEPLOYMENT               |
| GAMESTATEFULSET              |
| HOOKTEMPLATE                 |
| CLUSTERMANAGE                |
| CLUSTER                      |
| NODETEMPLATE                 |
| DEPLOYMENTMANAGE             |
| HELM                         |
| RELEASELIST                  |
| CHARTLIST                    |
| TEMPLATESET_v1               |
| TEMPLATESET                  |
| TEMPLATESET_DEPLOYMENT       |
| TEMPLATESET_STATEFULSET      |
| TEMPLATESET_DAEMONSET        |
| TEMPLATESET_JOB              |
| TEMPLATESET_INGRESSE         |
| TEMPLATESET_SERVICE          |
| TEMPLATESET_CONFIGMAP        |
| TEMPLATESET_SECRET           |
| TEMPLATESET_HPA              |
| TEMPLATESET_GAMEDEPLOYMENT   |
| TEMPLATESET_GAMESTATEFULSET  |
| TEMPLATESET_CUSTOMOBJECT     |
| TEMPLATE_FILE                |
| VARIABLE                     |
| PROJECTMANAGE                |
| EVENT                        |
| AUDIT                        |
| CLOUDTOKEN                   |
| TENCENTCLOUD                 |
| TENCENTPUBLICCLOUD           |
| GOOGLECLOUD                  |
| AZURECLOUD                   |
| HUAWEICLOUD                  |
| AMAZONCLOUD                  |
| PROJECT                      |
| PLUGINMANAGE                 |
| TOOLS                        |
| METRICS                      |
| LOG                          |
| MONITOR                      |

### 特性开关

特性开关用于特定场景，如某些版本不支持的功能或暂时不对外开放的特性需要通过特性开关处理，配置同`feature_flags`一样。注意：特性开关命名尽量简洁，保持风格统一

```yaml
feature_flags:
  VCLUSTER:
    enabled: false # 不开启VCLUSTER集群创建
```

目前支持的特性开关如下：

| 类型    | 描述 |
|--------| ---- |
| k8s    | 是否开启原生k8s集群创建 |
| VCLUSTER | 是否开始VCLUSTER集群创建 |
| BKAI | 是否开启AI |

备注：`feature_flags`下的数据是通过`接口`读取的

### 文档链接

前端所有文档相关地址通过可通过`*values*`下`frontend_conf`配置，全量配置如下：

```yaml
frontend_conf:
  docs:
    applyrecords: ''  # 主机申请记录
    cl5: '' # CL5路由使用方式
    cmdbhost: '' # CMDB地址
    contact: '' # 联系我们
    helm: '' # 推送helm chart到项目仓库
    help: '' # 帮助文档
    k8sConfigmap: '' # configmap文档
    k8sDaemonset: '' # Daemonset文档
    k8sDeployment: '' # Deployment文档
    k8sHpa: '' # HPA文档
    k8sIngress: '' # Ingress文档
    k8sJob: '' # job文档
    k8sSecret: '' # secret 文档
    k8sService: '' # service 文档
    k8sStatefulset: '' # Statefulset 文档
    nodeTemplate: '' # 节点模板文档
    nodemanHost: '' # 节点管理文档
    quickStart: '' # 入门文档
    rule: '' # 容器日志采集文档
    serviceAccess: ''
    teaApply: ''
    token: '' # token配置文档
    uiPrefix: '' # UI前缀，用于特性环境
    webConsole: '' # web console文档
```

配置好后，会渲染到前端的`BCS_CONFIG`全局变量上(在`index.html`上有定义)，如下：

```js
// 加载配置文件
window.BCS_CONFIG = JSON.parse('<%= process.env.BK_BCS_CONFIG %>' || '{}')
```

### HOST配置

HOST配置是必须的，主要用于接口地址，第三方系统依赖等，可通过`*values*`下`frontend_conf`配置, 最终会渲染到前端`index.html`上的全局变量上, 配置如下：

```yaml
frontend_conf:
  hosts:
    bk_paas_host: https://sg.crosgame.com
    bk_iam_host: https://bkiam.sg.crosgame.com
    bk_cc_host: https://cmdb.sg.crosgame.com
    bk_monitor_host: https://bkmonitor.sg.crosgame.com
    devops_host: ""
    devops_bcs_api_url: https://bcs.sg.crosgame.com
    devops_artifactory_host: ""
    login_full_url: "https://o.sg.crosgame.com/login/"
    bk_user_host: "https://bkapi.sg.crosgame.com"
    site_url: /bcs
    bk_log_host: https://bklog.sg.crosgame.com
    bk_shared_res_url: "https://bkrepo.sg.crosgame.com/generic/blueking/bk-config"
```

对应前端变量如下(.bk.development.env会把下面这些 `process.env` 的变量渲染成上`values`的key, 最终后端模板引擎再渲染成真实变量)：

```js
var LOGIN_FULL = '<%= process.env.BK_LOGIN_FULL %>'
var DEVOPS_HOST = '<%= process.env.BK_DEVOPS_HOST %>'
var DEVOPS_BCS_API_URL = '<%= process.env.BK_DEVOPS_BCS_API_URL %>'
var BK_STATIC_URL = '<%= process.env.BK_STATIC_URL %>'
var RUN_ENV = '<%= process.env.BK_RUN_ENV %>'
var DEVOPS_ARTIFACTORY_HOST = '<%= process.env.BK_DEVOPS_ARTIFACTORY_HOST %>'
var REGION = '<%= process.env.BK_REGION %>'
var SITE_URL = '<%= process.env.BK_SITE_URL %>'
var BK_IAM_HOST = '<%= process.env.BK_IAM_HOST %>'
var PAAS_HOST = '<%= process.env.BK_PAAS_HOST %>'
var BKMONITOR_HOST = '<%= process.env.BK_BKMONITOR_HOST %>'
var BCS_API_HOST = '<%= process.env.BK_BCS_API_HOST %>'
var PREFERRED_DOMAINS = '<%= process.env.BK_PREFERRED_DOMAINS %>'
var BK_CC_HOST = '<%= process.env.BK_CC_HOST %>'
var BK_SRE_HOST = '<%= process.env.BK_SRE_HOST %>'
var BK_USER_HOST = '<%= process.env.BK_USER_HOST %>'
var BCS_NAMESPACE_PREFIX = '<%= process.env.BK_BCS_NAMESPACE_PREFIX %>'
var BK_LOG_HOST = '<%= process.env.BK_LOG_HOST %>'
var BK_DOMAIN = '<%= process.env.BK_DOMAIN %>'
var BK_SHARED_RES_BASE_JS_URL = '<%= process.env.BK_SHARED_RES_BASE_JS_URL %>'
```