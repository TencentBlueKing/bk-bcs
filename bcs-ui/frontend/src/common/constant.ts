export const BCS_CLUSTER = 'bcs-cluster'
export const CUR_SELECT_NAMESPACE = 'CUR_SELECT_NAMESPACE'
export const CUR_SELECT_CRD = 'CUR_SELECT_CRD'
// 节点管理表格列展示配置
export const CLUSTER_NODE_TABLE_COL = 'CLUSTER_NODE_TABLE_COL'
export const nodeStatusColorMap = {
    initialization: 'blue',
    running: 'green',
    deleting: 'blue',
    'add-failure': 'red',
    'remove-failure': 'red',
    removable: '',
    notready: 'red',
    unknown: ''
}
export const nodeStatusMap = {
    initialization: window.i18n.t('初始化中'),
    running: window.i18n.t('正常'),
    deleting: window.i18n.t('删除中'),
    'add-failure': window.i18n.t('上架失败'),
    'remove-failure': window.i18n.t('下架失败'),
    removable: window.i18n.t('不可调度'),
    notready: window.i18n.t('不正常'),
    unknown: window.i18n.t('未知状态')
}
export const taskStatusTextMap = {
    initialzing: window.i18n.t('初始化中'),
    running: window.i18n.t('运行中'),
    success: window.i18n.t('成功'),
    failure: window.i18n.t('失败'),
    timeout: window.i18n.t('超时'),
    notstarted: window.i18n.t('未执行')
}
export const taskStatusColorMap = {
    initialzing: 'blue',
    running: 'blue',
    success: 'green',
    failure: 'red',
    timeout: 'red',
    notstarted: 'blue'
}
