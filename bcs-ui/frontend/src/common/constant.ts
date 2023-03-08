export const CUR_SELECT_CRD = '__CUR_SELECT_CRD__';
// 节点管理表格列展示配置
export const CLUSTER_NODE_TABLE_COL = '_CLUSTER_NODE_TABLE_COL_';
export const nodeStatusColorMap = {
  initialization: 'blue',
  running: 'green',
  deleting: 'blue',
  'add-failure': 'red',
  'remove-failure': 'red',
  'REMOVE-CA-FAILUR': '',
  removable: '',
  notready: 'red',
  unknown: '',
};
export const nodeStatusMap = {
  initialization: window.i18n.t('初始化中'),
  running: window.i18n.t('正常'),
  deleting: window.i18n.t('删除中'),
  'add-failure': window.i18n.t('上架失败'),
  'remove-failure': window.i18n.t('下架失败'),
  'REMOVE-CA-FAILUR': window.i18n.t('缩容成功,下架失败'),
  removable: window.i18n.t('不可调度'),
  notready: window.i18n.t('不正常'),
  unknown: window.i18n.t('未知状态'),
};
export const taskStatusTextMap = {
  initialzing: window.i18n.t('初始化中'),
  running: window.i18n.t('运行中'),
  success: window.i18n.t('成功'),
  failure: window.i18n.t('失败'),
  timeout: window.i18n.t('超时'),
  notstarted: window.i18n.t('未执行'),
};
export const taskStatusColorMap = {
  initialzing: 'blue',
  running: 'blue',
  success: 'green',
  failure: 'red',
  timeout: 'red',
  notstarted: 'blue',
};
export const NODE_TEMPLATE_ID = '_node-template-id_';
export const SPECIAL_REGEXP = /[`\s~!@#$%^&*()_+<>?:"{},./;'[\]]/;
export const LABEL_KEY_REGEXP = '^(?=.{1,253}$)([a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*\\/)?([A-Za-z0-9][-A-Za-z0-9_.]{0,61})?[A-Za-z0-9]$';
export const KEY_REGEXP = '^(([A-Za-z0-9][-A-Za-z0-9_./]*)?[A-Za-z0-9])?$';
export const VALUE_REGEXP = '^(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?$';

