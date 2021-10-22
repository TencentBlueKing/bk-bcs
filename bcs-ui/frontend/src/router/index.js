/**
 * @file router 配置
 */

import Vue from 'vue'
import VueRouter from 'vue-router'

import http from '@open/api'
import { bus } from '@open/common/bus'
import preload from '@open/common/preload'
import resourceRoutes from '@/router/resource'
import nodeRoutes from '@/router/node'
import mcRoutes from '@/router/mc'

import store from '@open/store'
import depotRoutes from './depot'
import metricRoutes from './metric'
import clusterRoutes from './cluster'
import appRoutes from './app'
import configurationRoutes from './configuration'
import networkRoutes from './network'
import helmRoutes from './helm'
import HPARoutes from './hpa'
import Crdcontroller from './crdcontroller.js'
import storageRoutes from './storage'
import dashboardRoutes from './dashboard'

Vue.use(VueRouter)

const ContainerServiceEntry = () => import(/* webpackChunkName: 'containerserviceentry' */'@/views')
const None = () => import(/* webpackChunkName: 'none' */'@open/views/none')

const children = clusterRoutes.concat(
    nodeRoutes,
    appRoutes,
    configurationRoutes,
    networkRoutes,
    resourceRoutes,
    depotRoutes,
    metricRoutes,
    mcRoutes,
    helmRoutes,
    HPARoutes,
    Crdcontroller,
    storageRoutes,
    dashboardRoutes
)

const routes = [
    {
        // domain/bcs
        // path: '/bcs',
        path: `${SITE_URL}`,
        name: 'containerServiceMain',
        component: ContainerServiceEntry,
        children: children
    },
    // 404
    {
        path: '*',
        name: '404',
        component: None
    }
]

const router = new VueRouter({
    mode: 'history',
    routes: routes
})

const cancelRequest = async () => {
    const allRequest = http.queue.get()
    const requestQueue = allRequest.filter(request => request.cancelWhenRouteChange)
    await http.cancel(requestQueue.map(request => request.requestId))
}

let preloading = true
let canceling = true
let pageMethodExecuting = true

router.beforeEach(async (to, from, next) => {
    bus.$emit('close-apply-perm-modal')

    canceling = true
    await cancelRequest()
    canceling = false
    next()

    // 解决 sideslider 组件跳转后导致滚动条失效
    const node = document.documentElement
    const className = 'bk-sideslider-show has-sideslider-padding'
    const classNames = className.split(' ')
    const rtrim = /^\s+|\s+$/
    let setClass = ' ' + node.className + ' '

    classNames.forEach(cl => {
        setClass = setClass.replace(' ' + cl + ' ', ' ')
    })
    node.className = setClass.replace(rtrim, '')

    setTimeout(() => {
        window.scroll(0, 0)
    }, 100)
})

router.afterEach(async (to, from) => {
    const bodySrcollTop = document.body.scrollTop
    if (bodySrcollTop !== 0) {
        document.body.scrollTop = 0
        return
    }
    const docSrcollTop = document.documentElement.scrollTop
    if (docSrcollTop !== 0) {
        document.documentElement.scrollTop = 0
    }

    store.commit('setMainContentLoading', true)
    preloading = true
    await preload()
    preloading = false

    const pageDataMethods = []
    const routerList = to.matched
    routerList.forEach(r => {
        const fetchPageData = r.instances.default && r.instances.default.fetchPageData
        if (fetchPageData && typeof fetchPageData === 'function') {
            pageDataMethods.push(r.instances.default.fetchPageData())
        }
    })

    pageMethodExecuting = true
    await Promise.all(pageDataMethods)
    pageMethodExecuting = false

    if (!preloading && !canceling && !pageMethodExecuting) {
        store.commit('setMainContentLoading', false)
    }
})

export default router
