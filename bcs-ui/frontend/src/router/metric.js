/**
 * @file metric router 配置
 */

const MetricManage = () => import(/* webpackChunkName: 'metric' */'@open/views/metric')

const childRoutes = [
    // domain/bcs/projectCode//metric-manage Metric 管理页面
    {
        path: ':projectCode/metric-manage',
        name: 'metricManage',
        component: MetricManage
    }
]

export default childRoutes
