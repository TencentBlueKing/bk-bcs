// 平台管理相关类型定义

// 项目管理相关类型定义
interface IPlatformProject {
  projectID: string                    // 项目ID
  name: string                        // 项目名称
  projectCode: string                 // 项目编码
  description: string                 // 项目描述
  creator: string                     // 创建者
  isOffline: boolean                  // 项目状态 true:已启用 false:已停用
  businessID: string                  // 所属业务ID
  businessName: string                // 所属业务名称
  managers: string                    // 管理者(多个用逗号分隔)
  createTime: string                  // 创建时间
  link?: string                       // 项目链接(前往项目)
  useBKRes?: boolean                  // 是否使用蓝鲸提供的资源池，主要用于资源计费，默认false
  kind?: string                       // 项目中集群类型，可选k8s/mesos
  labels?: Record<string, string>     // 项目标签
  annotations?: Record<string, string> // 项目注解
  updateTime?: string                 // 更新时间
  updater?: string                    // 更新者
}

interface IProjectListParams {
  limit?: number                      // 分页大小
  offset?: number                     // 分页偏移
  names?: string                      // 项目名称过滤
  searchName?: string                 // 搜索名称
  all?: boolean                       // 是否获取所有项目
  kind?: string                       // 集群类型过滤
  businessID?: string                 // 业务ID过滤
  projectCode?: string                // 项目编码过滤
}

interface IProjectListResponse {
  total: number                       // 总数
  results: IPlatformProject[]         // 项目列表
}

interface IProjectUpdateParams {
  managers?: string                   // 管理者
  businessID?: string                 // 业务ID
  name?: string                       // 项目名称
  useBKRes?: boolean                  // 是否使用蓝鲸资源
  description?: string                // 项目描述
  kind?: string                       // 集群类型
  labels?: Record<string, string>     // 项目标签
  annotations?: Record<string, string> // 项目注解
}

interface IProjectManagerUpdateParams {
  managers: string                    // 管理者
}

interface IProjectBusinessUpdateParams {
  businessID: string                  // 业务ID
}

interface IProjectStatusUpdateParams {
  isOffline: boolean                  // 项目状态
}

// 业务管理相关类型定义
interface IPlatformBusiness {
  bk_biz_id: number                   // 业务ID
  bk_biz_name: string                 // 业务名称
}

interface IBusinessListResponse {
  data: IPlatformBusiness[]           // 业务列表
}

// 通用 API 响应类型
interface IApiResponse<T = any> {
  code: number                        // 状态码
  message: string                     // 响应消息
  data: T                            // 响应数据
  request_id?: string                // 请求ID
}

// 分页响应类型
interface IPaginatedResponse<T> {
  total: number                       // 总数
  results: T[]                        // 数据列表
}

// 错误响应类型
interface IApiError {
  code: number                        // 错误码
  message: string                     // 错误消息
  details?: string                    // 错误详情
  request_id?: string                // 请求ID
}

// 请求配置类型
interface IRequestConfig {
  domain?: string                     // 请求域名
  prefix?: string                     // 请求前缀
}

// 请求函数类型
type RequestFunction = (params?: any, config?: any) => Promise<any>;

// createRequest 函数类型
declare function createRequest(config?: IRequestConfig): (method: string, url: string) => RequestFunction;
