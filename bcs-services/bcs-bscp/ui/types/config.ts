export interface IConfigVersion {
  id: number;
  attachment: {
    app_id: number;
    biz_id: number;
  };
  revision: {
    create_at: string;
    creator: string;
  };
  spec: {
    name: string;
    memo: string;
  };
}

export interface IConfigItem {
  id: number;
  spec: {
    file_mode: string;
    file_type: string;
    memo: string;
    name: string;
    path: string;
    permission: {
      privilege: string;
      user: string;
      user_group: string;
    }
  };
  attachment: {
    biz_id: number;
    app_id: number;
  };
  revision: {
    creator: string;
    create_at: string;
    reviser: string;
    update_at: string;
  }
}

// 配置项列表查询接口请求参数
export interface IConfigListQueryParams {
  release_id?: number;
  start?: number;
  limit?: number
}