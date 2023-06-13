import $i18n from '@/i18n/i18n-setup';
export const CUR_SELECT_CRD = '__CUR_SELECT_CRD__';
// 节点管理表格列展示配置
export const CLUSTER_NODE_TABLE_COL = '_CLUSTER_NODE_TABLE_COL_';
export const NODE_TEMPLATE_ID = '_node-template-id_';
export const SPECIAL_REGEXP = /[`\s~!@#$%^&*()_+<>?:"{},./;'[\]]/;
export const LABEL_KEY_REGEXP = '^(?=.{1,253}$)([a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*\\/)?([A-Za-z0-9][-A-Za-z0-9_.]{0,61})?[A-Za-z0-9]$';
export const KEY_REGEXP = '^(([A-Za-z0-9][-A-Za-z0-9_./]*)?[A-Za-z0-9])?$';
export const VALUE_REGEXP = '^(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?$';

// 集群环境
export const CLUSTER_ENV = {
  stag: 'UAT',
  debug: $i18n.t('测试'),
  prod: $i18n.t('正式'),
};
