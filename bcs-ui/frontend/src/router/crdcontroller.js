/**
 * @file crdcontroller router 配置
 */

const Index = () => import(/* webpackChunkName: 'network' */'@/views/crdcontroller/index.vue')
const DBList = () => import(/* webpackChunkName: 'network' */'@/views/crdcontroller/db_list.vue')
const LogList = () => import(/* webpackChunkName: 'network' */'@/views/crdcontroller/log_list.vue')
const Detail = () => import(/* webpackChunkName: 'network' */'@/views/crdcontroller/detail.vue')
const BcsPolaris = () => import(/* webpackChunkName: 'network' */'@/views/crdcontroller/polaris_list.vue')

const childRoutes = [
    {
        path: ':projectCode/crdcontroller/DbPrivilege',
        name: 'dbCrdcontroller',
        component: Index,
        meta: {
            crdKind: 'DbPrivilege'
        }
    },

    {
        path: ':projectCode/crdcontroller/BcsLog',
        name: 'logCrdcontroller',
        component: Index,
        meta: {
            crdKind: 'BcsLog'
        }
    },

    {
        path: ':projectCode/cluster/:clusterId/crdcontroller/DbPrivilege/instances',
        name: 'crdcontrollerDBInstances',
        component: DBList,
        meta: {
            menuId: 'COMPONENTS'
        }
    },

    {
        path: ':projectCode/cluster/:clusterId/crdcontroller/BcsPolaris/instances',
        name: 'crdcontrollerPolarisInstances',
        component: BcsPolaris
    },

    {
        path: ':projectCode/cluster/:clusterId/crdcontroller/BcsLog/instances',
        name: 'crdcontrollerLogInstances',
        component: LogList
    },

    {
        path: ':projectCode/cluster/:clusterId/crdcontroller/:chartName/:name/instances/:id',
        name: 'crdcontrollerInstanceDetail',
        component: Detail,
        meta: {
            menuId: 'COMPONENTS'
        }
    }
]

export default childRoutes
