export interface IAppListQuery {
  operator?: string;
  name?: string;
  start?: number;
  limit?: number;
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
        reload_file_path: string
      };
      reload_type: string
    }
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
  }
}

export interface IAppEditParams {
  id?: number;
  biz_id?: number|string;
  app_id?: number;
  name: string;
  file_type: string;
  path?: string;
  file_mode?: string;
  user?: string;
  user_group?: string;
  privilege?: string;
}
