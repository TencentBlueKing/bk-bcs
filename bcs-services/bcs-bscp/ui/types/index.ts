// 空间列表单个详情
export interface ISpaceItem {
  permission?: boolean;
  space_id: string;
  space_name: string;
  space_type_id: number;
  space_type_name: string;
  space_uid: number;
}

export interface IPermissionQuery {
  biz_id: number|string;
  basic: {
    type: string;
    action: string;
    resource_id: number|string;
  };
  gen_apply_url: boolean;
}

export interface IPermissionResource {
  action: string;
  action_name: string;
  resource_id: string;
  resource_name: string;
  type: string;
  type_name: string;
}