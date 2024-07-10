import { localT } from '../i18n';

// 错误大类映射
export const CLIENT_ERROR_CATEGORY_MAP = [
  {
    name: localT('前置脚本失败'),
    value: 'PreHookFailed',
  },
  {
    name: localT('后置脚本失败'),
    value: 'PostHookFailed',
  },
  {
    name: localT('下载配置文件失败'),
    value: 'DownloadFailed',
  },
  {
    name: localT('跳过失败'),
    value: 'SkipFailed',
  },
  {
    name: localT('Token错误'),
    value: 'TokenFailed',
  },
  {
    name: localT('版本太低'),
    value: 'VersionIsTooLowFailed',
  },
  {
    name: localT('获取AppMeta失败'),
    value: 'AppMetaFailed',
  },
  {
    name: localT('删除旧文件失败'),
    value: 'DeleteOldFilesFailed',
  },
  {
    name: localT('更新meatdata数据失败'),
    value: 'UpdateMetadataFailed',
  },
  {
    name: localT('未知错误'),
    value: 'UnknownFailed',
  },
];

// 错误子类映射
export const CLIENT_ERROR_SUBCLASSES_MAP = [
  {
    name: localT('新建文件夹失败'),
    value: 'NewFolderFailed',
  },
  {
    name: localT('遍历文件夹失败'),
    value: 'TraverseFolderFailed',
  },
  {
    name: localT('删除文件夹失败'),
    value: 'DeleteFolderFailed',
  },
  {
    name: localT('写入文件夹失败'),
    value: 'WriteFileFailed',
  },
  {
    name: localT('打开文件失败'),
    value: 'OpenFileFailed',
  },
  {
    name: localT('文件不存在'),
    value: 'StatFileFailed',
  },
  {
    name: localT('写入环境变量失败'),
    value: 'WriteEnvFileFailed',
  },
  {
    name: localT('检测文件是否存在失败'),
    value: 'CheckFileExistsFailed',
  },
  {
    name: localT('脚本类型不支持'),
    value: 'ScriptTypeNotSupported',
  },
  {
    name: localT('执行脚本失败'),
    value: 'ScriptExecutionFailed',
  },
  {
    name: localT('没有下载文件的权限'),
    value: 'NoDownloadPermission',
  },
  {
    name: localT('生成下载连接失败'),
    value: 'GenerateDownloadLinkFailed',
  },
  {
    name: localT('校验下载文件失败'),
    value: 'ValidateDownloadFailed',
  },
  {
    name: localT('重试下载文件失败'),
    value: 'RetryDownloadFailed',
  },
  {
    name: localT('数据为空'),
    value: 'DataEmpty',
  },
  {
    name: localT('序列化失败'),
    value: 'SerializationFailed',
  },
  {
    name: localT('格式化失败'),
    value: 'FormattingFailed',
  },
  {
    name: localT('token无权限'),
    value: 'TokenPermissionFailed',
  },
  {
    name: localT('sdk版本太低'),
    value: 'SDKVersionIsTooLowFailed',
  },
  {
    name: localT('未知特殊错误'),
    value: 'UnknownSpecificFailed',
  },
];

// 客户端组件类型映射
export const CLIENT_COMPONENT_TYPES_MAP = [
  {
    name: `SideCar ${localT('客户端')}`,
    value: 'sidecar',
  },
  {
    name: `SDK ${localT('客户端')}`,
    value: 'sdk',
  },
  {
    name: localT('主机插件客户端'),
    value: 'agent',
  },
  {
    name: `CLI ${localT('客户端')}`,
    value: 'command',
  },
];

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
    name: localT('目标配置版本'),
    value: 'current_release_name',
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
        name: localT('离线'),
        value: 'offline',
      },
    ],
  },
  {
    name: localT('客户端组件类型'),
    value: 'client_type',
    children: CLIENT_COMPONENT_TYPES_MAP,
  },
  {
    name: localT('客户端组件版本'),
    value: 'client_version',
  },
  {
    name: localT('配置拉取时间范围'),
    value: 'pull_time',
  },
  {
    name: localT('错误类别'),
    value: 'failed_reason',
    children: CLIENT_ERROR_CATEGORY_MAP,
  },
];

export const CLIENT_STATISTICS_SEARCH_DATA = [
  {
    name: localT('标签'),
    value: 'label',
  },
  {
    name: localT('目标配置版本'),
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
        name: localT('离线'),
        value: 'offline',
      },
    ],
  },
  {
    name: localT('客户端组件类型'),
    value: 'client_type',
    children: [
      {
        name: `SideCar ${localT('客户端')}`,
        value: 'sidecar',
      },
      {
        name: `SDK ${localT('客户端')}`,
        value: 'sdk',
      },
      {
        name: localT('主机插件客户端'),
        value: 'agent',
      },
      {
        name: `CLI ${localT('客户端')}`,
        value: 'command',
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
  Online: localT('在线'),
  Offline: localT('离线'),
  failed: localT('失败'),
  online: localT('在线'),
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
    label: localT('近 {n} 天', { n: 30 }),
  },
];
