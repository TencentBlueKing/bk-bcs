/**
 * @file 日志采集 router 配置
 */

const Index = () => import(/* webpackChunkName: 'network' */'@open/views/log/index.vue')
const List = () => import(/* webpackChunkName: 'network' */'@open/views/log/list.vue')

const childRoutes = [
    {
        path: ':projectCode/log-collection',
        name: 'logCollection',
        component: Index
    },

    {
        path: ':projectCode/cluster/:clusterId/log-collection/:crdKind/instances',
        name: 'logCollectionInstances',
        component: List
    }
]

export default childRoutes
