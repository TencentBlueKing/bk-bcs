import { IConfigVersion } from '../../types/config';
import { localT } from '../i18n';
// kv类型的配置项包含的子类型
export const CONFIG_KV_TYPE = [
  { id: 'string', name: 'String' },
  { id: 'number', name: 'Number' },
  { id: 'text', name: 'Text' },
  { id: 'json', name: 'JSON' },
  { id: 'xml', name: 'XML' },
  { id: 'yaml', name: 'YAML' },
];

// 文件类型的配置项包含的子类型
export const CONFIG_FILE_TYPE = [
  { id: 'text', name: localT('文本文件') },
  { id: 'binary', name: localT('二进制文件') },
];

export const CONFIG_STATUS_MAP = {
  ADD: {
    text: localT('新增'),
    color: '#3a84ff',
    bgColor: '#edf4ff',
  },
  DELETE: {
    text: localT('删除'),
    color: '#ea3536',
    bgColor: '#feebea',
  },
  REVISE: {
    text: localT('修改'),
    color: '#fe9c00',
    bgColor: '#fff1db',
  },
  UNCHANGE: {
    text: '--',
    color: '',
    bgColor: '',
  },
};

export const VERSION_STATUS_MAP = {
  not_released: localT('未上线'),
  partial_released: localT('灰度中'),
  full_released: localT('已上线'),
};

export const GET_UNNAMED_VERSION_DATA = (): IConfigVersion => ({
  id: 0,
  attachment: {
    app_id: 0,
    biz_id: 0,
  },
  revision: {
    create_at: '',
    creator: '',
  },
  spec: {
    name: localT('未命名版本'),
    memo: '',
    deprecated: false,
    publish_num: 0,
  },
  status: {
    publish_status: 'editing',
    released_groups: [],
    fully_released: false,
  },
});
