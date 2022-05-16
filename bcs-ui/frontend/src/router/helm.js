/**
 * @file helm router 配置
 */

// Helm应用列表
const helms = () => import(/* webpackChunkName: 'helm' */'@/views/helm/release/')

// Helm模板列表
const helmTplList = () => import(/* webpackChunkName: 'helm' */'@/views/helm/charts/index.vue')

// Helm模板详情
const helmTplDetail = () => import(/* webpackChunkName: 'helm' */'@/views/helm/charts/tpl-detail.vue')

// Helm实例化
const helmTplInstance = () => import(/* webpackChunkName: 'helm' */'@/views/helm/charts/tpl-deploy.vue')

// Helm app详情
const helmUpdateApp = () => import(/* webpackChunkName: 'helm' */'@/views/helm/release/update-app.vue')

// helm status详情
const helmAppStatus = () => import(/* webpackChunkName: 'helm' */'@/views/helm/release/detail.vue')

const childRoutes = [
    {
        path: ':projectCode/helm',
        name: 'helms',
        component: helms
    },
    {
        path: ':projectCode/helm/list',
        name: 'helmTplList',
        component: helmTplList
    },
    {
        path: ':projectCode/helm/tpl/:tplId',
        name: 'helmTplDetail',
        component: helmTplDetail,
        meta: {
            menuId: 'helmTplList'
        }
    },
    {
        path: ':projectCode/helm/instance/:chartName',
        name: 'helmTplInstance',
        component: helmTplInstance,
        meta: {
            menuId: 'helmTplList'
        }
    },
    {
        path: ':projectCode/helm/app/:clusterId/:namespace/:name',
        name: 'helmUpdateApp',
        component: helmUpdateApp,
        meta: {
            menuId: 'helms'
        }
    },
    {
        path: ':projectCode/helm/status/:clusterId/:namespace/:name',
        name: 'helmAppStatus',
        component: helmAppStatus,
        meta: {
            menuId: 'helms'
        }
    }
]

export default childRoutes
