import { ICredentialItem } from './credential';
// 搜索项
export interface ISelectorItem {
  name: string;
  value: string;
  children?: ISelectorItem[];
}

// 查询条件
export interface ISearchCondition {
  key: string;
  value: string;
  content: string;
  isEdit: boolean;
}

// 客户端查询条件公共接口
export interface IClinetCommonQuery {
  start?: number;
  limit?: number;
  last_heartbeat_time?: number;
  all?: boolean;
  order?: {
    desc?: string;
    asc?: string;
  };
  search?: IClientSearchParams;
  start_time?: string;
  end_time?: string;
  search_value?: string;
  search_type?: string;
  pull_time?: number;
  is_duplicates?: boolean;
}

// 客户端查询列表接口查询条件
export interface IClientSearchParams {
  uid?: string;
  ip?: string;
  label?: string[];
  current_release_name?: string;
  target_release_name?: string;
  release_change_status?: string[];
  annotations?: string;
  online_status?: string[];
  client_version?: string;
  client_type?: string;
  start_pull_time?: string;
  end_pull_time?: string;
  failed_reason?: string;
  client_ids?: number[];
}

export interface IGetClientSearchListQuery {
  start: number;
  limit: number;
  last_heartbeat_time: number;
  all?: boolean;
  order?: {
    desc?: string;
    asc?: string;
  };
  search?: IClientSearchParams;
}

// 新增常用查询接口参数
export interface ICreateClientSearchRecordQuery {
  search_type: string;
  search_name?: string;
  search_condition: IClientSearchParams;
}

// 常用查询列表项
export interface ICommonlyUsedItem {
  id: number;
  spec: {
    creator: string;
    search_name: string;
    search_condition: IClientSearchParams;
  };
  search_condition: ISearchCondition[];
}

// 客户端配置版本
export interface IClientConfigVersionItem {
  count: number;
  target_release_id: number; // 版本ID
  target_release_name: string; // 版本名称
  percent: number;
}

// 拉取成功率
export interface IPullSuccessRate {
  count: number;
  percent: number;
  release_change_status: string;
}

// 信息展示卡片
export interface IInfoCard {
  value: string | number;
  name: string;
  key: string;
  unit?: string;
}

export interface IPullErrorReason {
  count: number;
  percent: number;
  release_change_failed_reason: string;
}

// 拉取数量趋势
export interface IPullCount {
  time: {
    count: number;
    time: string;
  }[];
  time_and_type: {
    time: string;
    value: number;
    type: string;
  }[];
}

// 客户端标签
export interface IClientLabelItem {
  count: number;
  foreign_key: string;
  foreign_val: string;
  percent: number;
  primary_key: string;
  primary_val: string;
  x_field: string; // 平铺柱状图渲染参数
}

// 组件版本发布(柱状图和表格)
export interface IVersionDistributionItem {
  client_type: string;
  client_version: string;
  percent: number;
  value: number;
  name?: string;
}

// 组件版本发布(旭日图)
export interface IVersionDistributionPie {
  name: string;
  children: IVersionDistributionPieItem[];
}

export interface IVersionDistributionPieItem {
  name: string;
  client_type: string;
  value: number;
  percent: number;
  children?: IVersionDistributionPieItem[];
}

// interface IClusterInfo {
//   name: string;
//   value: string;
// }
export interface IExampleFormData {
  clientKey: string;
  privacyCredential: string;
  tempDir: string;
  configName?: string;
  labelArr: string[];
  clusterSwitch?: boolean;
  clusterInfo?: string;
  rules?: string[];
  systemType?: 'Unix' | 'Windows';
  selectedLineBreak?: 'LF' | 'CRLF';
}

export type newICredentialItem = ICredentialItem & {
  spec: {
    privacyCredential: string;
  };
};
