export const enum EScriptType {
  Shell = 'shell',
  Python = 'python'
}

export interface IScriptEditingForm {
  name: string;
  tag: string;
  release_name: string;
  memo: string;
  type: EScriptType;
  content: string;
}

export interface IScriptItem {
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
  },
  revision: {
      creator: string;
      reviser: string;
      create_at: string;
      update_at: string;
  }
}

export interface IScriptVersion {
  id: number;
  spec: {
      name: string;
      content: string;
      publish_num: number;
      pub_state : string;
      momo: string;
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
  }
}

// 脚本被引用列表查询参数
export interface IScriptCiteQuery {
  start: number;
  limit: number;
  searchKey?: string;
}