import { localT } from '../i18n';
export const CLIENT_SEARCH_DATA = [
  {
    name: 'UID',
    value: 'uid',
  },
  {
    name: 'IP',
    value: 'ip',
  },
  {
    name: localT('标签'),
    value: 'label',
  },
  {
    name: localT('当前配置版本'),
    value: 'current_release_name',
  },
  {
    name: localT('目标配置版本'),
    value: 'target_release_name',
  },
  {
    name: localT('最近一次拉取配置状态'),
    value: 'release_change_status',
    children: [
      {
        name: localT('成功'),
        value: 'Success',
      },
      {
        name: localT('失败'),
        value: 'Failed',
      },
      {
        name: localT('处理中'),
        value: 'Processing',
      },
      {
        name: localT('跳过'),
        value: 'Skip',
      },
    ],
  },
  // {
  //   name: '附加信息',
  //   value: 'annotations',
  // },
  {
    name: localT('在线状态'),
    value: 'online_status',
    children: [
      {
        name: localT('在线'),
        value: 'online',
      },
      {
        name: localT('未在线'),
        value: 'offline',
      },
    ],
  },
  {
    name: localT('客户端组件版本'),
    value: 'client_version',
  },
];

export const CLIENT_STATUS_MAP = {
  Success: localT('成功'),
  Failed: localT('失败'),
  Processing: localT('处理中'),
  Skip: localT('跳过'),
  online: localT('在线'),
  offline: localT('未在线'),
  success: localT('成功'),
  failed: localT('失败'),
  processing: localT('处理中'),
  skip: localT('跳过'),
};

export const CLIENT_HEARTBEAT_LIST = [
  {
    value: 1,
    label: `${localT('近')}1${localT('分钟')}`,
  },
  {
    value: 5,
    label: `${localT('近')}5${localT('分钟')}`,
  },
  {
    value: 60,
    label: `${localT('近')}1${localT('小时')}`,
  },
  {
    value: 360,
    label: `${localT('近')}6${localT('小时')}`,
  },
  {
    value: 720,
    label: `${localT('近')}12${localT('小时')}`,
  },
  {
    value: 1440,
    label: `${localT('近')}1${localT('天')}`,
  },
  {
    value: 4320,
    label: `${localT('近')}3${localT('天')}`,
  },
  {
    value: 10080,
    label: `${localT('近')}7${localT('天')}`,
  },
  {
    value: 20160,
    label: `${localT('近')}15${localT('天')}`,
  },
  {
    value: 43200,
    label: `${localT('近')}30${localT('天')}`,
  },
];
