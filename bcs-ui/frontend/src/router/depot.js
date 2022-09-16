/**
 * @file depot router 配置
 */

const Depot = () => import(/* webpackChunkName: 'depot' */'@/views/depot')
const ImageLibrary = () => import(/* webpackChunkName: 'depot' */'@/views/depot/image-library')
const ImageDetail = () => import(/* webpackChunkName: 'depot' */'@/views/depot/image-detail')
const ProjectImage = () => import(/* webpackChunkName: 'depot' */'@/views/depot/project-image')

const childRoutes = [
    // 这里没有把 depot 作为 cluster 的 children
    // 是因为如果把 depot 作为 cluster 的 children，那么必须要在 Cluster 的 component 中
    // 通过 router-view 来渲染子组件，但在业务逻辑中，depot 和 cluster 是平级的
    {
        path: ':projectCode/depot',
        name: 'depotMain',
        component: Depot,
        children: [
            // domain/bcs/projectCode/depot => domain/bcs/projectCode/depot/image-library
            {
                path: 'image-library',
                component: ImageLibrary,
                name: 'imageLibrary',
                alias: ''
            },
            {
                path: 'image-detail/:imageRepo',
                component: ImageDetail,
                name: 'imageDetail',
                alias: '',
                props: true,
                meta: {
                    menuId: 'imageLibrary'
                }
            },
            {
                path: 'project-image',
                name: 'projectImage',
                component: ProjectImage
            }
        ]
    }
]

export default childRoutes
