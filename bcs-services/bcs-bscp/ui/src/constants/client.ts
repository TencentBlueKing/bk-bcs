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
    name: '附加信息',
    value: 'annotations',
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
