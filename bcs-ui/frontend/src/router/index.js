/**
 * @file router 配置
 */

import Vue from 'vue'
import VueRouter from 'vue-router'
import store from '@/store'
import http from '@/api'
import resourceRoutes from '@/router/resource'
import nodeRoutes from '@/router/node'
import mcRoutes from '@/router/mc'
import depotRoutes from './depot'
import metricRoutes from './metric'
import clusterRoutes from './cluster'
import appRoutes from './app'
import configurationRoutes from './configuration'
import networkRoutes from './network'
import helmRoutes from './helm'
import HPARoutes from './hpa'
import crdController from './crdcontroller.js'
import storageRoutes from './storage'
import dashboardRoutes from './dashboard'
import menuConfig from '@/store/menu'

const originalPush = VueRouter.prototype.push
const originalReplace = VueRouter.prototype.replace
// push
VueRouter.prototype.push = function push (location, onResolve, onReject) {
    if (onResolve || onReject) return originalPush.call(this, location, onResolve, onReject)
    return originalPush.call(this, location).catch(err => err)
}
// replace
VueRouter.prototype.replace = function push (location, onResolve, onReject) {
    if (onResolve || onReject) return originalReplace.call(this, location, onResolve, onReject)
    return originalReplace.call(this, location).catch(err => err)
}
Vue.use(VueRouter)

const Entry = () => import(/* webpackChunkName: entry */'@/views/index')
const NotFound = () => import(/* webpackChunkName: 'none' */'@/components/exception')
const ProjectManage = () => import(/* webpackChunkName: 'projectmanage' */'@/views/project/project.vue')
const userToken = () => import(/* webpackChunkName: 'token' */'@/views/token/token.vue')
const Forbidden = () => import(/* webpackChunkName: 'none' */'@/components/exception/403.vue')

const router = new VueRouter({
    mode: 'history',
    routes: [
        {
            path: `${SITE_URL}`,
            name: 'entry',
            component: Entry,
            children: [
                ...clusterRoutes,
                ...nodeRoutes,
                ...appRoutes,
                ...configurationRoutes,
                ...networkRoutes,
                ...resourceRoutes,
                ...depotRoutes,
                ...metricRoutes,
                ...mcRoutes,
                ...helmRoutes,
                ...HPARoutes,
                ...crdController,
                ...storageRoutes,
                ...dashboardRoutes
            ]
        },
        {
            path: '/api_key',
            name: 'token',
            component: userToken
        },
        {
            path: '/projectManage',
            name: 'projectManage',
            component: ProjectManage
        },
        {
            path: '/403',
            name: '403',
            props: (route) => ({ ...route.params, ...route.query }),
            component: Forbidden
        },
        // 404
        {
            path: '*',
            name: '404',
            component: NotFound
        }
    ]
})

const cancelRequest = async () => {
    const allRequest = http.queue.get()
    const requestQueue = allRequest.filter(request => request.cancelWhenRouteChange)
    await http.cancel(requestQueue.map(request => request.requestId))
}

router.beforeEach(async (to, from, next) => {
    // 设置必填路由参数
    if (!to.params.projectId && store.state.curProjectId) {
        to.params.projectId = store.state.curProjectId
    }
    if (!to.params.projectCode && store.state.curProjectCode) {
        to.params.projectCode = store.state.curProjectCode
    }
    if (!to.params.clusterId && store.state.cluster.curCluster) {
        to.params.clusterId = store.state.cluster.curCluster.cluster_id
    }

    await cancelRequest()
    next()
})

let containerEle = null
router.afterEach((to, from) => {
    if (!containerEle) {
        containerEle = document.getElementsByClassName('container-content')
    }
    if (containerEle && containerEle[0] && containerEle[0].scrollTop !== 0) {
        containerEle[0].scrollTop = 0
    }

    // 设置左侧菜单栏选中项
    let activeMenuId = to.meta?.menuId // 1. 是否指定了菜单ID
    if (!activeMenuId) { // 2. 在菜单配置中查找当前路由对应的菜单ID
        const menuList = to.meta.isDashboard ? menuConfig.dashboardMenuList : menuConfig.k8sMenuList
        menuList.find(menu => {
            if (menu?.routeName === to.name) {
                activeMenuId = menu?.id
                return true
            } else if (menu.children) {
                const child = menu.children.find(child => child.routeName === to.name)
                activeMenuId = child?.id
                return !!activeMenuId
            }
            return false
        })
    }
    if (activeMenuId) {
        store.commit('updateCurMenuId', activeMenuId)
    } else {
        console.warn('找不到当前路由对应的菜单项，请检查', to)
    }
})

export default router
