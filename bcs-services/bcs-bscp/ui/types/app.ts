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
    deploy_type: string;
    config_type: string;
    mode: string;
    memo: string;
    reload: {
      file_reload_spec: {
        reload_file_path: string;
      };
      reload_type: string;
    };
    alias: string;
    data_type: string;
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
