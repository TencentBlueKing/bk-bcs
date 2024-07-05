import { IGroupRuleItem } from './group';

// 单个版本详情
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
    deprecated: boolean;
    publish_num: number;
  };
  status: {
    publish_status: string;
    released_groups: IReleasedGroup[];
    fully_released: boolean;
    fully_release?: boolean;
  };
}

// 单个配置详情
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
    };
  };
  commit_spec: {
    content: {
      byte_size: string;
      origin_byte_size: string;
      origin_signature: string;
      signature: string;
    };
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
  };
  file_state: string;
  is_conflict: boolean;
}

// 配置文件详情（包含签名信息）
export interface IConfigDetail {
  config_item: IConfigItem;
  content: {
    signature: string;
    byte_size: string;
  };
}

// 配置文件编辑表单参数
export interface IConfigEditParams {
  id?: number;
  name: string;
  memo: string;
  file_type: string;
  path?: string;
  file_mode?: string;
  user?: string;
  user_group?: string;
  privilege?: string;
  fileAP?: string;
  revision_name?: string;
}

// kv配置文件编辑表单参数
export interface IConfigKvEditParams {
  key: string;
  kv_type: string;
  value: string;
  memo: string;
}

// 文件配置概览内容
export interface IFileConfigContentSummary {
  id?: number;
  name: string;
  signature: string;
  size: string;
  update_at?: string;
}

// 配置文件列表查询接口请求参数
export interface IConfigListQueryParams {
  searchKey?: string;
  release_id?: number;
  start?: number;
  limit?: number;
  all?: boolean;
}

// 版本列表查询接口请求参数
export interface IConfigVersionQueryParams {
  searchKey?: string;
  start?: number;
  limit?: number;
  all?: boolean;
  deprecated?: boolean;
}

// 分组发布到某个版本数据
export interface IReleasedGroup {
  edited: boolean;
  id: number;
  mode: string;
  name: string;
  new_selector: {
    labels_and?: IGroupRuleItem[];
    labels_or?: IGroupRuleItem[];
  };
  old_selector: {
    labels_and?: IGroupRuleItem[];
    labels_or?: IGroupRuleItem[];
  };
  uid: string;
}

// 模板套餐与服务的绑定数据
export interface ITemplateBoundByAppData {
  template_set_id: number;
  template_revisions: {
    template_id: number;
    template_revision_id: number;
    is_latest: boolean;
  }[];
  template_set_name?: string;
}

// 服务绑定下的模板配置文件按照套餐分组数据
export interface IBoundTemplateGroup {
  template_space_id: number;
  template_space_name: string;
  template_set_id: number;
  template_set_name: string;
  template_revisions: IBoundTemplateDetail[];
}

// 服务绑定下的模板配置详情数据
export interface IBoundTemplateDetail {
  template_id: number;
  name: string;
  path: string;
  template_revision_id: number;
  is_latest: true;
  template_revision_name: string;
  template_revision_memo: string;
  file_type: string;
  file_mode: string;
  file_state: string;
  user: string;
  user_group: string;
  privilege: string;
  origin_signature: string;
  signature: string;
  origin_byte_size: string;
  byte_size: string;
  creator: string;
  create_at: string;
  update_at: string;
  is_conflict: boolean;
}

// 配置文件对比选中项
export interface IConfigDiffSelected {
  pkgId: number; // 套餐ID
  id: number; // 非模板或模板配置文件 ID
  version: number; // 版本ID
  permission?: {
    privilege: string;
    user: string;
    user_group: string;
  };
}

// 导入配置项
export interface IConfigImportItem {
  byte_size: number;
  file_mode: string;
  name: string;
  path: string;
  file_type: string;
  memo: string;
  privilege: string;
  user: string;
  user_group: string;
  signature: string;
  id: number;
  variables: {
    default_val: string;
    memo: string;
    name: string;
    type: string;
  }[];
  file_name?: string;
  is_exist: boolean;
}

// kv类型
export interface IConfigKvItem {
  key: string;
  kv_type: string;
  value: string;
  memo: string;
  is_exist?: boolean;
}

// 单个kv配置详情
export interface IConfigKvType {
  id: number;
  spec: IConfigKvItem;
  content_spec: {
    signature: string;
    byte_size: string;
  };
  kv_state: string;
  attachment: {
    biz_id: number;
    app_id: number;
  };
  revision: {
    creator: string;
    create_at: string;
    reviser: string;
    update_at: string;
  };
}
