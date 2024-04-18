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

export const CLIENT_STATISTICS_SEARCH_DATA = [
  {
    name: localT('标签'),
    value: 'label',
  },
  {
    name: localT('当前配置版本'),
    value: 'current_release_name',
  },
  // {
  //   name: '附加信息',
  //   value: 'annotations',
  // },
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
  {
    name: localT('客户端组件版本'),
    value: 'client_version',
  },
  {
    name: localT('客户端组件类型'),
    value: 'client_type',
  },
];

export const CLIENT_STATUS_MAP = {
  Success: localT('成功'),
  Failed: localT('失败'),
  Processing: localT('处理中'),
  Skip: localT('跳过'),
  Online: localT('在线'),
  Offline: localT('离线'),
  failed: localT('失败'),
  offline: localT('离线'),
};

export const CLIENT_HEARTBEAT_LIST = [
  {
    value: 1,
    label: localT('近 {n} 分钟', { n: 1 }),
  },
  {
    value: 5,
    label: localT('近 {n} 分钟', { n: 5 }),
  },
  {
    value: 60,
    label: localT('近 {n} 小时', { n: 1 }),
  },
  {
    value: 360,
    label: localT('近 {n} 小时', { n: 6 }),
  },
  {
    value: 720,
    label: localT('近 {n} 小时', { n: 12 }),
  },
  {
    value: 1440,
    label: localT('近 {n} 天', { n: 1 }),
  },
  {
    value: 4320,
    label: localT('近 {n} 天', { n: 3 }),
  },
  {
    value: 10080,
    label: localT('近 {n} 天', { n: 7 }),
  },
  {
    value: 20160,
    label: localT('近 {n} 天', { n: 15 }),
  },
  {
    value: 43200,
    label: localT('近 {n} 天', { n: 30}),
  },
];
