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
  };
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
  };
}

// 模板套餐左侧菜单单条配置文件
export interface IPackageMenuItem {
  id: number | string;
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

// 模板套餐下配置文件列表单条数据
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
  };
}
// 模板被套餐引用详情
export interface ITemplateCitedByPkgs {
  template_id: number;
  template_name: string;
  template_set_id: number;
  template_set_name: string;
}

// 单个模板套餐被多个服务引用数据

export interface IPackageCitedByApps {
  template_revision_id: number;
  template_revision_name: string;
  app_id: number;
  app_name: string;
}

// 多个模板套餐被多个服务引用数据
export interface IPackagesCitedByApps {
  template_set_id: number;
  template_set_name: string;
  app_id: number;
  app_name: string;
}

// 模板被服务绑定或套餐引用计数详情
export interface ITemplateCitedCountDetailItem {
  bound_named_app_count: number;
  bound_template_set_count: number;
  bound_unnamed_app_count: number;
  template_id: number;
}

// 模板被未命名版本服务绑定详情
export interface IAppBoundByTemplateDetailItem {
  template_revision_id: number;
  template_revision_name: string;
  app_id: number;
  app_name: string;
}

// 模板版本单条数据
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
    };
  };
  attachment: {
    biz_id: number;
    template_space_id: number;
    template_id: number;
  };
  revision: {
    creator: string;
    create_at: string;
  };
}

// 模板版本编辑数据
export interface ITemplateVersionEditingData {
  revision_name: string;
  revision_memo: string;
  file_type: string;
  file_mode: string;
  user: string;
  user_group: string;
  privilege: string;
  sign: string;
  byte_size: number;
}

// 业务下所有模板套餐列表（按模板空间分组）
export interface IAllPkgsGroupBySpaceInBiz {
  template_space_id: number;
  template_space_name: string;
  template_sets: {
    template_ids: number[];
    template_set_id: number;
    template_set_name: string;
    is_latest: boolean;
  }[];
}

// 业务下所有模板套餐列表树
export interface IPkgTreeItem {
  id: number;
  nodeId: string;
  name: string;
  checked?: boolean;
  disabled?: boolean;
  indeterminate?: boolean;
  parentName?: string;
}

// 模板下多个版本名称
export interface ITemplateVersionsName {
  template_id: number;
  template_name: string;
  latest_template_revision_id: number;
  template_revisions: {
    template_revision_id: number;
    template_revision_name: string;
    template_revision_memo: string;
  }[];
}

// 版本对比
export interface DiffSliderDataType {
  id: number;
  versionId: number;
  name: string;
  permission?: {
    privilege: string;
    user: string;
    user_group: string;
  };
}

// 从历史版本导入配置模板
export interface ImportTemplateConfigItem {
  template_space_name: string;
  template_space_id: number;
  template_set_name: string;
  template_set_id: number;
  template_space_exist: boolean;
  template_set_exist: boolean;
  is_exist: boolean;
  template_set_is_empty: boolean;
  template_revisions: {
    template_id: number;
    template_revision_id: number;
    is_latest: boolean;
    template_space_id: number;
    variables: {
      default_val: string;
      memo: string;
      name: string;
      type: string;
    }[];
  }[];
}
