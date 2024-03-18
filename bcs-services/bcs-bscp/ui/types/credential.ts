import { IAppItem } from './app';
export interface ICredentialItem {
  id: number;
  attachment: {
    biz_id: number;
  };
  revision: {
    create_at: string;
    creator: string;
    expired_at: string;
    reviser: string;
    update_at: string;
  };
  spec: {
    credential_type: string;
    enable: boolean;
    enc_algorithm: string;
    enc_credential: string;
    memo: string;
    name: string;
  };
  visible?: boolean; // 另外增加，用于在标识是否在界面上展示明文
}

// 密钥关联规则
export interface ICredentialRule {
  id: number;
  spec: {
    scope: string;
    app: string;
  };
  attachment: {
    biz_id: number;
    credential_id: number;
  };
  revision: {
    creator: string;
    reviser: string;
    create_at: string;
    update_at: string;
    expired_at: string;
  };
}

// 关联规则编辑数据
export interface IRuleEditing {
  id: number;
  type: string;
  content: string;
  original: string;
  app: IAppItem | null;
  originalApp: string;
  isRight: boolean;
  isSelectService: boolean;
  needPreview: boolean;
}

// 调用关联规则更新接口参数
interface IRuleUpdateItem {
  app: string;
  scope: string;
  id?: number;
}
export interface IRuleUpdateParams {
  add_scope: IRuleUpdateItem[];
  del_id: number[];
  alter_scope: IRuleUpdateItem[];
}

// 关联规则预览项参数
export interface IPreviewRule {
  id: number; // 规则id
  scopeContent: string; // 规则内容
  appName: string; // 服务名称
}

export interface IPreviewRuleParams {
  start: number;
  limit: number;
  app_name: string;
  scope: string;
  search_value: string;
}
