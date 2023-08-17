// 模板空间列表单条数据
export interface ITemplateSpaceItem {
  id: number;
  spec: {
    name: string;
    memo: string;
  };
  attachment: {
    biz_id: number;
  };
  revision: {
    creator: string;
    reviser: string;
    create_at: string;
    update_at: string;
  }
}

// 模板套餐列表单条数据
export interface ITemplatePackageItem {
  id: number;
  spec: {
    name: string;
    memo: string;
    template_ids: number[];
    public: boolean;
    bound_apps: number[];
  };
  attachment: {
    biz_id: number;
    template_space_id: number;
  };
  revision: {
    creator: string;
    reviser: string;
    create_at: string;
    update_at: string;
  }
}

// 模板套餐左侧菜单单条配置项
export interface IPackageMenuItem {
  id: number|string;
  name: string;
  count: number;
}

// 模板套餐编辑参数
export interface ITemplatePackageEditParams {
  template_set_id?: number;
  name: string;
  memo: string;
  template_ids?: number[];
  public: boolean;
  bound_apps: number[];
  force?: boolean;
}

// 模板套餐下配置项列表单条数据
export interface ITemplateConfigItem {
  id: number;
  spec: {
    name: string;
    path: string;
    memo: string;
  };
  attachment: {
    biz_id: number;
    template_space_id: number;
  };
  revision: {
    creator: string;
    reviser: string;
    create_at: string;
    update_at: string;
  }
}

export interface ITemplateVersionItem {
  id: number;
  spec: {
    revision_name: string;
    revision_memo: string;
    name: string;
    path: string;
    file_type: string;
    file_mode: string;
    permission: {
      user: string;
      user_group: string;
      privilege: string;
    };
    content_spec: {
      signature: string;
      byte_size: number;
    }
  };
  attachment: {
    biz_id: number;
    template_space_id: number;
    template_id: number;
  };
  revision: {
    creator: string;
    create_at: string;
  }
}

