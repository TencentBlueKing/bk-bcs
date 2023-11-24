
// 空间详情
export interface ISpaceDetail {
  permission?: boolean;
  space_id: string;
  space_name: string;
  space_type_id: number;
  space_type_name: string;
  space_uid: number;
}

// 分页参数
export interface IPagination {
  current: number;
  limit: number;
  count: number;
}

export interface ICommonQuery {
  start: number;
  limit?: number;
  all?: boolean;
  search_key?: string;
  search_fields?: string;
  search_value?: string;
  ids?: string;
}

// 权限查询参数单个资源条目
export interface IPermissionQueryResourceItem {
  biz_id: number|string;
  basic: {
    type: string;
    action: string;
    resource_id?: number|string;
  };
}

// 权限申请资源信息
export interface IPermissionResource {
  action: string;
  action_name: string;
  resource_id: string;
  resource_name: string;
  type: string;
  type_name: string;
}
