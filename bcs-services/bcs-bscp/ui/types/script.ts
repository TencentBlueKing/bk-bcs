export const enum EScriptType {
  Shell = 'shell',
  Python = 'python',
}

// 脚本列表查询请求参数
export interface IScriptListQuery {
  start: number;
  limit?: number;
  all?: boolean;
  tag?: string;
  not_tag?: boolean;
  name?: string;
}

// 新建脚本编辑参数
export interface IScriptEditingForm {
  name: string;
  tag: string;
  memo: string;
  type: EScriptType;
  content: string;
  version: string;
}

// 脚本
export interface IScriptItem {
  bound_num: number;
  confirm_delete: boolean;
  hook: {
    id: number;
    spec: {
      name: string;
      type: string;
      tag: string;
      publish_num: number;
      momo: string;
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
  };
  published_revision_id: number;
}

// 脚本标签
export interface IScriptTagItem {
  tag: string;
  counts: number;
}

// 脚本版本
export interface IScriptVersion {
  id: number;
  spec: {
    name: string;
    content: string;
    publish_num: number;
    state: string;
    memo: string;
  };
  attachment: {
    biz_id: number;
    hook_id: number;
  };
  revision: {
    creator: string;
    reviser: string;
    create_at: string;
    update_at: string;
  };
}

// 脚本版本列表项
export interface IScriptVersionListItem {
  bound_num: number;
  confirm_delete: boolean;
  hook_revision: IScriptVersion;
}

// 脚本版本新建、编辑表单
export interface IScriptVersionForm {
  id: number;
  name: string;
  memo: string;
  content: string;
}

// 脚本被引用列表查询参数
export interface IScriptCiteQuery {
  start: number;
  limit: number;
  release_id?: number;
  searchKey?: string;
}

// 脚本被引用数据
export interface IScriptCitedItem {
  hook_release_name: string;
  app_id: number;
  app_name: string;
  config_release_name: string;
  config_release_id: number;
  pub_state: string;
}

// 脚本加工后的列表项
export interface IScriptMapItem {
  id: number;
  name: string;
  content: string;
  isOnline: string;
}
