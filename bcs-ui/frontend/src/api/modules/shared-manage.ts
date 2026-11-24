import { createRequest } from '../request';

// request.js 为 JS 模块，TS 推断 createRequest 的参数类型时仅保留了有默认值的 prefix，
// 丢失 domain 字段，这里断言补全（与 project.ts 等既有模块的传参方式保持一致）
const createRequestWithDomain = createRequest as (
  options: { domain?: string, prefix?: string }
) => ReturnType<typeof createRequest>;

// 共享集群管理（bcs-project-manager）
// 接口路径参考：/bcsproject/v1/projects/{projectCode}/clusters/{clusterID}/...
const request = createRequestWithDomain({
  domain: window.BCS_API_HOST,
  prefix: '/bcsapi/v4/bcsproject/v1/projects/$projectCode',
});

/**
 * 资源配额
 * cpuRequests / cpuLimits 如 "8" 表示 8 核；memoryRequests / memoryLimits 如 "16Gi"
 */
export interface IResourceQuota {
  cpuRequests: string
  cpuLimits: string
  memoryRequests: string
  memoryLimits: string
}

/**
 * 命名空间数据（ListNamespaces 返回，NamespaceData）
 * 注：quota / used / source / updateTime / allocatable / disableReason 等为后端扩展字段
 */
export interface INativeNamespace {
  uid: string
  name: string
  clusterID: string
  // 命名空间状态（k8s Phase，如 Active）
  status: string
  createTime: string
  updateTime?: string
  // 归属项目（为空表示未分配）
  projectID: string
  projectCode: string
  // 配额 / 已用量（后端扩展字段）
  quota?: IResourceQuota
  used?: IResourceQuota
  // 使用率（小数，如 0.00625 表示 0.625%）
  cpuUseRate?: number
  memoryUseRate?: number
  // 来源：使用方申请 / 所属方分配
  source?: string
  // 是否可操作（vcluster 命名空间、审批中返回 false）
  allocatable?: boolean
  // 不可操作原因（后端下发，优先展示）
  disableReason?: string
  // 系统命名空间标识
  isSystem?: boolean
  // ITSM 审批单状态（非空表示创建审批中）
  itsmTicketStatus?: string
  // ITSM 审批单链接（单据跳转）
  itsmTicketURL?: string
}

/**
 * 配额数值解析：CPU 支持 "8"（核）/ "100m"（毫核）
 */
export const parseCpu = (val?: string) => {
  const str = String(val ?? '').trim();
  if (!str) return 0;
  if (str.endsWith('m')) return (Number.parseFloat(str) || 0) / 1000;
  return Number.parseFloat(str) || 0;
};

/**
 * 配额数值解析：内存统一转 GiB（支持 Ki / Mi / Gi / Ti，无单位按 GiB）
 */
export const parseMemoryGiB = (val?: string) => {
  const str = String(val ?? '').trim();
  if (!str) return 0;
  const num = Number.parseFloat(str) || 0;
  if (str.endsWith('Ki')) return num / 1024 ** 2;
  if (str.endsWith('Mi')) return num / 1024;
  if (str.endsWith('Ti')) return num * 1024;
  return num;
};

/**
 * 命名空间分配 / 变更归属入参（AllocateNamespace）
 */
export interface IAllocateNamespaceParams {
  // 目标归属项目编码，必须在集群可见范围内
  targetProjectCode: string
  // 命名空间配额（无配额的命名空间必填）
  quota?: IResourceQuota
  // 是否覆盖命名空间已有配额，默认 false：quota 仅对无配额的命名空间生效
  overwrite?: boolean
}

/**
 * 命名空间回收入参（RecycleNamespace，回收 = 解除归属）
 */
export interface IRecycleNamespaceParams {
  // 是否同时删除命名空间中的 ResourceQuota，默认 true
  deleteQuota?: boolean
}

// 列出集群全量命名空间（ListNamespaces，含已分配 / 未分配，返回配额、归属项目等字段）
export const listNamespaces = request('get', '/clusters/$clusterId/namespaces');

// 分配已有命名空间 / 变更归属
export const allocateNamespace = request('post', '/clusters/$clusterId/namespaces/$namespace/allocate');

// 回收命名空间（解除归属，ns 与工作负载保留）
export const recycleNamespace = request('post', '/clusters/$clusterId/namespaces/$namespace/recycle');

// 命名空间配额更新（AllocateNamespace overwrite = true 的单点调整场景备用）
export const updateNamespaceQuota = request('put', '/clusters/$clusterId/namespaces/$namespace');

export default {
  listNamespaces,
  allocateNamespace,
  recycleNamespace,
  updateNamespaceQuota,
};
