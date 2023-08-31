import { IGroupRuleItem } from "./group";

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
  };
  status: {
    publish_status: string;
    released_groups: IReleasedGroup[];
  }
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
  };
  file_state: string;
}

// 配置项详情（包含签名信息）
export interface IConfigDetail {
  config_item: IConfigItem;
  content: {
    signature: string;
    byte_size: string;
  }
}

// 配置项编辑表单参数
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
}

// 文件配置概览内容
export interface IFileConfigContentSummary {
  id?: number;
  name: string;
  signature: string;
  size: string;
  update_at?: string;
}

// 配置对比差异详情
export interface IConfigDiffDetail {
  id: number;
  name: string;
  type: string;
  current: string;
  base: string;
}

// 配置项列表查询接口请求参数
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
    labels_and: IGroupRuleItem[]
  };
  old_selector: {
    labels_and: IGroupRuleItem[]
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
  }[]
}

// 服务绑定下的模板配置详情数据
export interface IBoundTemplateDetail {
  template_space_id: number;
  template_space_name: string;
  template_set_id: number;
  template_set_name: string
  template_id: number;
  name: string;
  path: string;
  template_revision_id: number;
  is_latest: true;
  template_revision_name: string;
  template_revision_memo: string;
  file_type: string;
  file_mode: string;
  user: string;
  user_group: string;
  privilege: string;
  signature: string;
  byte_size: string;
  creator: string;
  create_at: string;
}
