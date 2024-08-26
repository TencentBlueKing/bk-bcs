interface INodePool {
  enableAutoscale: boolean
  nodeGroupID: string
  name: string
  autoScaling: {
    maxSize: number
    desiredSize: number
  }
}

interface ICloudTemplateDetail {
  cloudID: string
  clusterManagement: {
    availableVersion: string[]
  }
  osManagement: {
    regions: Record<string, string>
    availableVersion: string[]
  }
}

interface IRuntimeModuleParams {
  enable: boolean
  flagName: string
  defaultValue: string
  flagValueList: string[]
  networkType: string
}

type CloudID = 'tencentCloud'|'gcpCloud'|'tencentPublicCloud'|'bluekingCloud'|'azureCloud'|'huaweiCloud'| 'awsCloud';

interface IViewFilter {
  name?: string
  creator?: string[]
  labelSelector?: Array<{
    key: string
    op: '='|'In'|'NotIn'|'Exists'|'DoesNotExist'
    values: string[]
  }>
}

interface IClusterNamespace {
  clusterID: string
  namespaces: string[]
}

interface IViewData {
  id?: string
  name?: string
  projectCode?: string
  scope?: number
  filter: IViewFilter
  clusterNamespaces: IClusterNamespace[]
  createBy?: string
  createAt?: string
  updateAt?: string
}

interface IFieldItem {
  title: string
  id: string
  status: 'added' | ''// added: 已经添加的条件, 空: 为添加的条件
  placeholder?: string
}

// 获得函数参数类型的辅助类型
type ParamTypes<T> = T extends (...args: infer P) => any ? P : never;

type MakeOptional<T, K extends keyof T> = Omit<T, K> & Partial<Pick<T, K>>;

interface IFetchConfig {
  needRes?: boolean// 是否需要返回原始资源
  cancelPrevious?: boolean// 是否需要上一次重复的请求
  cancelWhenRouteChange?: boolean// 当前路由切换时当前请求可以被取消
  globalError?: boolean // 捕获全局异常
}
