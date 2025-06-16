// 前端文档配置
interface ILink {
  applyrecords: string // 主机申请记录
  cl5: string // CL5路由使用方式
  cmdbhost: string // CMDB地址
  contact: string // 联系我们
  helm: string // 推送helm chart到项目仓库
  help: string // 帮助文档
  k8sConfigmap: string // configmap文档
  k8sDaemonset: string // Daemonset文档
  k8sDeployment: string // Deployment文档
  k8sHpa: string // HPA文档
  k8sIngress: string// Ingress文档
  k8sJob: string // job文档
  k8sSecret: string // secret 文档
  k8sService: string // service 文档
  k8sStatefulset: string // Statefulset 文档
  nodeTemplate: string // 节点模板文档
  nodemanHost: string // 节点管理文档
  quickStart: string // 入门文档
  rule: string // 容器日志采集文档
  serviceAccess: string
  teaApply: string
  token: string // token配置文档
  uiPrefix: string // UI前缀，用于特性环境
  webConsole: string // web console文档
  backToLegacyButtonUrl: string // 回退到旧版link
  bkBcsEnvID: string // 客户端环境ID
}

interface Window {
  _project_code_: string
  _project_id_: string
  bus: any
  mainComponent: any
  readonly BCS_API_HOST: string
  readonly DEVOPS_BCS_API_URL: string
  readonly i18n: {
    t: (word: string) => string
  }
  readonly BCS_CONFIG: ILink // 文档链接配置
  readonly REGION: string
  readonly PAAS_HOST: string
  readonly BK_IAM_HOST: string
  readonly DEVOPS_HOST: string
  readonly LOGIN_FULL: string
  readonly BKMONITOR_HOST: string
  readonly RUN_ENV: string
  readonly BK_USER_HOST: string
  readonly PREFERRED_DOMAINS: string
  $loginModal: any
  BkTrace: any
  readonly BK_STATIC_URL: string
  readonly BCS_NAMESPACE_PREFIX: string
  readonly BK_LOG_HOST: string
  readonly BK_DOMAIN: string
  readonly BK_SHARED_RES_BASE_JS_URL: string
  readonly BK_CC_HOST: string
}

declare const BK_BCS_WELCOME: string;

declare const BK_BCS_VERSION: string;

declare const SITE_URL: string;

declare const DEVOPS_BCS_API_URL: string;
