/**
 * @file network router 配置
 */

const Service = () => import(/* webpackChunkName: 'network' */'@open/views/network/service')
const LoadBalance = () => import(/* webpackChunkName: 'network' */'@open/views/network/loadbalance')
const LoadBalanceDetail = () => import(/* webpackChunkName: 'network' */'@open/views/network/loadbalance-detail')
const childRoutes = [
    {
        path: ':projectCode/service',
        name: 'service',
        component: Service
    },
    {
        path: ':projectCode/loadbalance',
        name: 'loadBalance',
        component: LoadBalance
    },
    {
        path: ':projectCode/cluster/:clusterId/namespace/:namespace/loadbalance/:lbId/detail',
        name: 'loadBalanceDetail',
        component: LoadBalanceDetail
    }
]

export default childRoutes
