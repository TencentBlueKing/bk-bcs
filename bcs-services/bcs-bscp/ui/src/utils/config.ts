import { CONFIG_FILE_TYPE } from '../constants/config';
import dayjs from 'dayjs';

// 查询配置文件类型名称
export const getConfigTypeName = (type: string) => {
  const fileType = CONFIG_FILE_TYPE.find((item) => item.id === type);
  return fileType ? fileType.name : '未知格式';
};

export function getDefaultConfigItem() {
  return {
    id: 0,
    spec: {
      file_mode: '',
      file_type: '',
      memo: '',
      name: '',
      path: '',
      permission: {
        privilege: '644',
        user: 'root',
        user_group: 'root',
      },
    },
    attachment: {
      biz_id: 0,
      app_id: 0,
    },
    revision: {
      creator: '',
      create_at: '',
      reviser: '',
      update_at: '',
    },
  };
}

export function getDefaultKvItem() {
  return {
    id: 0,
    spec: {
      kv_type: '',
      key: '',
      value: '',
    },
    content_spec: {
      byte_size: '',
      signature: '',
    },
    kv_state: '',
    attachment: {
      biz_id: 0,
      app_id: 0,
    },
    revision: {
      creator: '',
      create_at: '',
      reviser: '',
      update_at: '',
    },
  };
}

// 配置文件编辑参数
export function getConfigEditParams() {
  return {
    name: '',
    memo: '',
    path: '',
    file_type: 'text',
    file_mode: 'unix',
    user: 'root',
    user_group: 'root',
    privilege: '644',
    revision_name: `v${dayjs().format('YYYYMMDDHHmmss')}`,
  };
}

// 拼接文件型配置项路径和文件名称
export function joinPathName(path: string, name: string) {
  return path.endsWith('/') ? `${path}${name}` : `${path}/${name}`;
}
