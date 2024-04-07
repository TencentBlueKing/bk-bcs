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
    name: '标签',
    value: 'label',
  },
  {
    name: '当前配置版本',
    value: 'current_release_name',
  },
  {
    name: '目标配置版本',
    value: 'target_release_name',
  },
  {
    name: '最近一次拉取配置状态',
    value: 'release_change_status',
    children: [
      {
        name: '成功',
        value: 'Success',
      },
      {
        name: '失败',
        value: 'Failed',
      },
      {
        name: '处理中',
        value: 'Processing',
      },
      {
        name: '跳过',
        value: 'Skip',
      },
    ],
  },
  {
    name: '在线状态',
    value: 'online_status',
    children: [
      {
        name: '在线',
        value: 'online',
      },
      {
        name: '未在线',
        value: 'offline',
      },
    ],
  },
  {
    name: '客户端组件版本',
    value: 'client_version',
  },
];

export const CLIENT_STATISTICS_SEARCH_DATA = [
  {
    name: '标签',
    value: 'label',
  },
  {
    name: '当前配置版本',
    value: 'current_release_name',
  },
  // {
  //   name: '附加信息',
  //   value: 'annotations',
  // },
  {
    name: '最近一次拉取配置状态',
    value: 'release_change_status',
    children: [
      {
        name: '成功',
        value: 'Success',
      },
      {
        name: '失败',
        value: 'Failed',
      },
      {
        name: '处理中',
        value: 'Processing',
      },
      {
        name: '跳过',
        value: 'Skip',
      },
    ],
  },
  {
    name: '客户端组件版本',
    value: 'client_version',
  },
  {
    name: '客户端组件类型',
    value: 'client_type',
  },
];

export const CLIENT_STATUS_MAP = {
  Success: '成功',
  Failed: '失败',
  Processing: '处理中',
  Skip: '跳过',
  online: '在线',
  offline: '离线',
  success: '成功',
  failed: '失败',
  processing: '处理中',
  skip: '跳过',
};

export const CLIENT_HEARTBEAT_LIST = [
  {
    value: 1,
    label: '近1分钟',
  },
  {
    value: 5,
    label: '近5分钟',
  },
  {
    value: 60,
    label: '近1小时',
  },
  {
    value: 360,
    label: '近6小时',
  },
  {
    value: 720,
    label: '近12小时',
  },
  {
    value: 1440,
    label: '近1天',
  },
  {
    value: 4320,
    label: '近3天',
  },
  {
    value: 10080,
    label: '近7天',
  },
  {
    value: 20160,
    label: '近15天',
  },
  {
    value: 43200,
    label: '近30天',
  },
];
