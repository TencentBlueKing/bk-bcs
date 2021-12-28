/**
 * @file cluster router 配置
 */

// 集群首页
const Cluster = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster')
// 创建集群
const ClusterCreate = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster/create-cluster/create-cluster')
// 表单模式
const CreateFormCluster = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster/create-cluster/create-form-cluster')
// import模式
const CreateImportCluster = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster/create-cluster/create-import-cluster')
// 集群模板创建
const CreateClusterTemplate = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster/create-cluster/create-template-cluster')
// 集群总览
const ClusterOverview = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster/overview')
// 节点详情
const ClusterNode = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster/node')
// 集群信息
const ClusterInfo = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster/info')
const ClusterNodeOverview = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster/node-overview')
const ContainerDetailForNode = () => import(/* webpackChunkName: 'cluster' */'@/views/cluster/container')

const childRoutes = [
    {
        path: ':projectCode',
        redirect: {
            name: 'clusterMain'
        }
    },
    {
        path: ':projectCode/',
        redirect: {
            name: 'clusterMain'
        }
    },
    // domain/bcs/projectId => domain/bcs/projectCode/cluster 容器服务的首页是具体项目的集群页面
    // 集群首页
    {
        path: ':projectCode/cluster',
        name: 'clusterMain',
        component: Cluster
    },
    // 创建集群
    {
        path: ':projectCode/cluster/create',
        name: 'clusterCreate',
        component: ClusterCreate,
        meta: {
            menuId: 'CLUSTER',
            title: window.i18n.t('创建容器集群')
        }
    },
    // 创建集群 - 表单模式
    {
        path: ':projectCode/cluster/create/form',
        name: 'createFormCluster',
        component: CreateFormCluster,
        meta: {
            title: window.i18n.t('新增集群')
        }
    },
    // 创建集群 - import导入模式
    {
        path: ':projectCode/cluster/create/import',
        name: 'createImportCluster',
        component: CreateImportCluster,
        meta: {
            title: window.i18n.t('导入集群')
        }
    },
    // 创建集群模板
    {
        path: ':projectCode/cluster/create/template',
        name: 'createClusterTemplate',
        component: CreateClusterTemplate,
        meta: {
            title: window.i18n.t('新建集群模板')
        }
    },
    // 集群总览
    {
        path: ':projectCode/cluster/:clusterId/overview',
        name: 'clusterOverview',
        component: ClusterOverview,
        alias: ':projectCode/cluster/:clusterId'
    },
    // 集群里的节点列表
    {
        path: ':projectCode/cluster/:clusterId/node',
        name: 'clusterNode',
        component: ClusterNode,
        meta: {
            menuId: 'OVERVIEW'
        }
    },
    // 集群里的集群信息
    {
        path: ':projectCode/cluster/:clusterId/info',
        name: 'clusterInfo',
        component: ClusterInfo,
        meta: {
            menuId: 'OVERVIEW'
        }
    },
    // 集群里的具体节点
    {
        path: ':projectCode/cluster/:clusterId/node/:nodeId',
        name: 'clusterNodeOverview',
        component: ClusterNodeOverview,
        meta: {
            menuId: 'OVERVIEW'
        }
    },
    // 节点详情页面跳转的容器详情页面
    {
        path: ':projectCode/cluster/:clusterId/node/:nodeId/container/:containerId',
        name: 'containerDetailForNode',
        component: ContainerDetailForNode,
        meta: {
            menuId: 'OVERVIEW'
        }
    }
]

export default childRoutes
