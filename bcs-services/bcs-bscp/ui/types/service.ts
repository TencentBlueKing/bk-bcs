import { IFileConfigContentSummary } from './config';
import { IVariableEditParams } from './variable';

export interface IServiceEditForm {
  name: string;
  config_type: string;
  kv_type: string;
  reload_type: string,
  reload_file_path: string;
  mode: string;
  deploy_type: string;
  memo: string;
}

// 版本下的脚本配置
export interface IDiffDetail {
  contentType: 'file'|'text';
  base: {
    content: string|IFileConfigContentSummary;
    language?: string;
    variables?: IVariableEditParams[];
    permission?: {
      privilege: string;
      user: string;
      user_group: string;
    };
  },
  current: {
    content: string|IFileConfigContentSummary;
    language?: string;
    variables?: IVariableEditParams[];
    permission?: {
      privilege: string;
      user: string;
      user_group: string;
    };
  }
}
