import { IFileConfigContentSummary } from './config';
import { IVariableEditParams } from './variable';

export interface IServiceEditForm {
  name: string;
  alias: string;
  config_type: string;
  memo: string;
  data_type?: string;
  is_approve: boolean;
  approver: string;
  approve_type: string;
  // encryptionSwtich: boolean;
  // encryptionKey: string;
}

export interface ISingleLineKVDIffItem {
  id: number;
  name: string;
  diffType: string;
  is_secret: boolean;
  secret_hidden: boolean;
  base: {
    content: string;
  };
  current: {
    content: string;
  };
  isCipherShowValue?: boolean;
}

// 版本下的脚本配置
export interface IDiffDetail {
  contentType: 'file' | 'text' | 'singleLineKV';
  id: number | string;
  is_secret: boolean;
  secret_hidden: boolean;
  base: {
    content: string | IFileConfigContentSummary;
    language?: string;
    variables?: IVariableEditParams[];
    permission?: {
      privilege: string;
      user: string;
      user_group: string;
    };
  };
  current: {
    content: string | IFileConfigContentSummary;
    language?: string;
    variables?: IVariableEditParams[];
    permission?: {
      privilege: string;
      user: string;
      user_group: string;
    };
  };
  singleLineKVDiff?: ISingleLineKVDIffItem[];
}
