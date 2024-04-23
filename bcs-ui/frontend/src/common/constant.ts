import $i18n from '@/i18n/i18n-setup';
export const CUR_SELECT_CRD = '__CUR_SELECT_CRD__';
// 节点管理表格列展示配置
export const CLUSTER_NODE_TABLE_COL = '_CLUSTER_NODE_TABLE_COL_';
export const NODE_TEMPLATE_ID = '_node-template-id_';
export const SPECIAL_REGEXP = /[`\s~!@#$%^&*()+<>?:"{},./;'[\]]/;
export const LABEL_KEY_REGEXP = '^(?=.{1,253}$)([a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*\\/)?([A-Za-z0-9][-A-Za-z0-9_.]{0,61})?[A-Za-z0-9]$';
export const KEY_REGEXP = '^(([A-Za-z0-9][-A-Za-z0-9_./]*)?[A-Za-z0-9])?$';
export const VALUE_REGEXP = '^(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?$';

// K8S正则
export const K8S_LABEL_KEY = '^((([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9])(.)?)+/)?[a-zA-Z0-9]([-_.a-zA-Z0-9]{0,61}[a-zA-Z0-9])?$';
export const K8S_LABEL_VALUE = '^([a-zA-Z0-9]?([-_.a-zA-Z0-9]{0,61}[a-zA-Z0-9])?)?$';
export const K8S_ANNOTATIONS_KEY = '^(?:(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?.)+[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?/)?[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$';

// 凭证 正则
export const NAME_REGEX = '^[0-9a-zA-Z-]+$';
export const SECRET_REGEX = '^[0-9a-zA-Z-~]+$';

// 集群环境
export const CLUSTER_ENV = {
  stag: 'UAT',
  debug: $i18n.t('cluster.tag.debug'),
  prod: $i18n.t('cluster.tag.prod'),
};

export const LOG_COLLECTOR = 'bk-log-collector';

export const ENCODE_LIST = [
  {
    id: 'UTF-8',
    name: 'UTF-8',
  },
  {
    id: 'GBK',
    name: 'GBK',
  },
  {
    id: 'GB18030',
    name: 'GB18030',
  },
  {
    id: 'BIG5',
    name: 'BIG5',
  },
  {
    id: 'ISO8859-6E',
    name: 'ISO8859-6E',
  },
  {
    id: 'ISO8859-6I',
    name: 'ISO8859-6I',
  },
  {
    id: 'ISO8859-8E',
    name: 'ISO8859-8E',
  },
  {
    id: 'ISO8859-8I',
    name: 'ISO8859-8I',
  },
  {
    id: 'ISO8859-1',
    name: 'ISO8859-1',
  },
  {
    id: 'ISO8859-2',
    name: 'ISO8859-2',
  },
  {
    id: 'ISO8859-3',
    name: 'ISO8859-3',
  },
  {
    id: 'ISO8859-4',
    name: 'ISO8859-4',
  },
  {
    id: 'ISO8859-5',
    name: 'ISO8859-5',
  },
  {
    id: 'ISO8859-6',
    name: 'ISO8859-6',
  },
  {
    id: 'ISO8859-7',
    name: 'ISO8859-7',
  },
  {
    id: 'ISO8859-8',
    name: 'ISO8859-8',
  },
  {
    id: 'ISO8859-9',
    name: 'ISO8859-9',
  },
  {
    id: 'ISO8859-10',
    name: 'ISO8859-10',
  },
  {
    id: 'ISO8859-13',
    name: 'ISO8859-13',
  },
  {
    id: 'ISO8859-14',
    name: 'ISO8859-14',
  },
  {
    id: 'ISO8859-15',
    name: 'ISO8859-15',
  },
  {
    id: 'ISO8859-16',
    name: 'ISO8859-16',
  },
  {
    id: 'CP437',
    name: 'CP437',
  },
  {
    id: 'CP850',
    name: 'CP850',
  },
  {
    id: 'CP852',
    name: 'CP852',
  },
  {
    id: 'CP855',
    name: 'CP855',
  },
  {
    id: 'CP858',
    name: 'CP858',
  },
  {
    id: 'CP860',
    name: 'CP860',
  },
  {
    id: 'CP862',
    name: 'CP862',
  },
  {
    id: 'CP863',
    name: 'CP863',
  },
  {
    id: 'CP865',
    name: 'CP865',
  },
  {
    id: 'CP866',
    name: 'CP866',
  },
  {
    id: 'EBCDIC-037',
    name: 'EBCDIC-037',
  },
  {
    id: 'EBCDIC-1040',
    name: 'EBCDIC-1040',
  },
  {
    id: 'EBCDIC-1047',
    name: 'EBCDIC-1047',
  },
  {
    id: 'KOI8R',
    name: 'KOI8R',
  },
  {
    id: 'KOI8U',
    name: 'KOI8U',
  },
  {
    id: 'MACINTOSH',
    name: 'MACINTOSH',
  },
  {
    id: 'MACINTOSH-CYRILLIC',
    name: 'MACINTOSH-CYRILLIC',
  },
  {
    id: 'WINDOWS1250',
    name: 'WINDOWS1250',
  },
  {
    id: 'WINDOWS1251',
    name: 'WINDOWS1251',
  },
  {
    id: 'WINDOWS1252',
    name: 'WINDOWS1252',
  },
  {
    id: 'WINDOWS1253',
    name: 'WINDOWS1253',
  },
  {
    id: 'WINDOWS1254',
    name: 'WINDOWS1254',
  },
  {
    id: 'WINDOWS1255',
    name: 'WINDOWS1255',
  },
  {
    id: 'WINDOWS1256',
    name: 'WINDOWS1256',
  },
  {
    id: 'WINDOWS1257',
    name: 'WINDOWS1257',
  },
  {
    id: 'WINDOWS1258',
    name: 'WINDOWS1258',
  },
  {
    id: 'WINDOWS874',
    name: 'WINDOWS874',
  },
  {
    id: 'UTF-16-BOM',
    name: 'UTF-16-BOM',
  },
  {
    id: 'UTF-16BE-BOM',
    name: 'UTF-16BE-BOM',
  },
  {
    id: 'UTF-16LE-BOM',
    name: 'UTF-16LE-BOM',
  },
];

export const CLUSTER_MAP = {
  INITIALIZATION: $i18n.t('generic.status.initializing'),
  DELETING: $i18n.t('generic.status.deleting'),
  'CREATE-FAILURE': $i18n.t('generic.status.createFailed'),
  'DELETE-FAILURE': $i18n.t('generic.status.deleteFailed'),
  'IMPORT-FAILURE': $i18n.t('cluster.status.importFailed'),
  RUNNING: $i18n.t('generic.status.ready'),
};

// 磁盘类型
export const diskEnum = [
  {
    id: 'CLOUD_PREMIUM',
    name: $i18n.t('cluster.ca.nodePool.create.instanceTypeConfig.diskType.premium'),
  },
  {
    id: 'CLOUD_SSD',
    name: $i18n.t('cluster.ca.nodePool.create.instanceTypeConfig.diskType.ssd'),
  },
  {
    id: 'CLOUD_HSSD',
    name: $i18n.t('cluster.ca.nodePool.create.instanceTypeConfig.diskType.hssd'),
  },
];
