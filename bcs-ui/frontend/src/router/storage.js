/**
 * @file storage router 配置
 */

const Storage = () => import(/* webpackChunkName: 'storage' */'@open/views/storage')
const PV = () => import(/* webpackChunkName: 'storage' */'@open/views/storage/pv')
const PVC = () => import(/* webpackChunkName: 'storage' */'@open/views/storage/pvc')
const StorageClass = () => import(/* webpackChunkName: 'storage' */'@open/views/storage/storage-class')

const childRoutes = [
    // 这里没有把 depot 作为 cluster 的 children
    // 是因为如果把 depot 作为 cluster 的 children，那么必须要在 Cluster 的 component 中
    // 通过 router-view 来渲染子组件，但在业务逻辑中，depot 和 cluster 是平级的
    {
        path: ':projectCode/storage',
        name: 'storageMain',
        component: Storage,
        children: [
            {
                path: 'pv',
                component: PV,
                name: 'pv',
                alias: ''
            },
            {
                path: 'pvc',
                component: PVC,
                name: 'pvc'
            },
            {
                path: 'storage-class',
                name: 'storageClass',
                component: StorageClass
            }
        ]
    }
]

export default childRoutes
