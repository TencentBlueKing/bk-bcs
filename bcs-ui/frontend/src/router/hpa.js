/**
 * @file hpa router 配置
 */

// Helm应用列表
const HPAIndex = () => import(/* webpackChunkName: 'helm' */'@open/views/hpa')

const childRoutes = [
    {
        path: ':projectCode/hpa',
        name: 'hpa',
        component: HPAIndex
    }
]

export default childRoutes
