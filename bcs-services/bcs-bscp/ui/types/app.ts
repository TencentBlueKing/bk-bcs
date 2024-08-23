export interface IAppListQuery {
  operator?: string;
  name?: string;
  start?: number;
  limit?: number;
  all?: boolean;
}

export interface IAppItem {
  id?: number;
  biz_id: number;
  space_id: string;
  spec: {
    name: string;
    config_type: string;
    memo: string;
    alias: string;
    data_type: string;
    is_approve: boolean;
    approver: string;
    approve_type: string;
  };
  revision: {
    creator: string;
    reviser: string;
    create_at: string;
    update_at: string;
  };
  config?: {
    count: number;
    update_at: string;
  };
  permissions: {
    [key: string]: string;
  };
}
