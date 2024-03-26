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
}

// 客户端查询条件公共接口
export interface IClinetCommonQuery {
  start: number;
  limit: number;
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
}

// 客户端查询列表接口查询条件
export interface IClientSearchParams {
  uid?: string;
  ip?: string;
  label?: { [key: string]: string };
  current_release_name?: string;
  target_release_name?: string;
  release_change_status?: string[];
  annotations?: string;
  online_status?: string[];
  client_version?: string;
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
