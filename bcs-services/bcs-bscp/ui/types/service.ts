import { IFileConfigContentSummary } from './config';

// 版本下的脚本配置
export interface IDiffDetail {
  contentType: 'file'|'text';
  base: {
    content: string|IFileConfigContentSummary;
    language?: string;
  },
  current: {
    content: string|IFileConfigContentSummary;
    language?: string;
  }
}
