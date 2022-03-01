/**
 * @file main store
 */

import Vue from 'vue'
import Vuex from 'vuex'
import cookie from 'cookie'

import http from '@/api'
import { unifyObjectStyle, json2Query } from '@/common/util'

import depot from '@/store/modules/depot'
import metric from '@/store/modules/metric'
import mc from '@/store/modules/mc'
import cluster from '@/store/modules/cluster'
import resource from '@/store/modules/resource'
import app from '@/store/modules/app'
import variable from '@/store/modules/variable'
import configuration from '@/store/modules/configuration'
import templateset from '@/store/modules/templateset'
import network from '@/store/modules/network'
import k8sTemplate from '@/store/modules/k8s-template'
import helm from '@/store/modules/helm'
import crdcontroller from '@/store/modules/crdcontroller'
import log from '@/store/modules/log'
import hpa from '@/store/modules/hpa'
import storage from '@/store/modules/storage'
import dashboard from '@/store/modules/dashboard'
import clustermanager from '@/store/modules/clustermanager'
import token from '@/store/modules/token'
import { projectFeatureFlag } from '@/api/base'

Vue.use(Vuex)
Vue.config.devtools = NODE_ENV === 'development'
// cookie 中 zh-cn / en
let lang = cookie.parse(document.cookie).blueking_language || 'zh-cn'
if (['zh-CN', 'zh-cn', 'cn', 'zhCN', 'zhcn'].indexOf(lang) > -1) {
    lang = 'zh-CN'
} else {
    lang = 'en-US'
}

const store = new Vuex.Store({
    // 模块
    modules: {
        depot,
        metric,
        mc,
        cluster,
        resource,
        app,
        variable,
        configuration,
        templateset,
        network,
        k8sTemplate,
        helm,
        hpa,
        crdcontroller,
        storage,
        dashboard,
        log,
        clustermanager,
        token
    },
    // 公共 store
    state: {
        curProject: {},
        curProjectCode: '', // 项目代码
        curProjectId: '', // 项目ID
        curClusterId: null,
        mainContentLoading: false,
        // 系统当前登录用户
        user: {},
        // 左侧导航
        sideMenu: {
            // 在线的 project
            onlineProjectList: []
        },

        // 当前语言环境
        lang: lang,
        isEn: lang === 'en-US',

        // 是否允许路由跳转
        allowRouterChange: true,

        crdInstanceList: [],
        // 功能开关
        featureFlag: {},
        viewMode: '',
        curMenuId: '',
        menuList: []
    },
    // 公共 getters
    getters: {
        mainContentLoading: state => state.mainContentLoading,
        user: state => state.user,
        lang: state => state.lang,
        featureFlag: state => state.featureFlag,
        curNavName: state => {
            let navName = ''
            state.menuList.find(menu => {
                if (menu?.id === state.curMenuId) {
                    navName = menu?.routeName
                    return true
                } else if (menu.children) {
                    const child = menu.children.find(child => child.id === state.curMenuId)
                    navName = child?.routeName
                    return !!navName
                }
                return false
            })
            return navName
        },
        curProjectCode: state => state.curProjectCode,
        curProjectId: state => state.curProjectId,
        curClusterId: state => state.curClusterId,
        menuList: state => state.menuList
    },
    // 公共 mutations
    mutations: {
        updateProjectCode (state, code) {
            state.curProjectCode = code
        },
        updateProjectId (state, id) {
            state.curProjectId = id
        },
        /**
         * 设置内容区的 loading 是否显示
         *
         * @param {Object} state store state
         * @param {boolean} loading 是否显示 loading
         */
        setMainContentLoading (state, loading) {
            state.mainContentLoading = loading
        },

        /**
         * 更新当前用户 user
         *
         * @param {Object} state store state
         * @param {Object} user user 对象
         */
        updateUser (state, user) {
            state.user = Object.assign({}, user)
        },

        /**
         * 更改当前项目信息
         *
         * @param {Object} state store state
         * @param {String} projectId
         */
        updateCurProject (state, project) {
            state.curProject = project || {}
        },

        /**
         * 修改state.curClusterId
         *
         * @param {Object} state store state
         * @param {boolean} val 值
         */
        updateCurClusterId (state, val) {
            state.curClusterId = val
        },

        /**
         * 修改 state.allowRouterChange
         *
         * @param {Object} state store state
         * @param {boolean} val 值
         */
        updateAllowRouterChange (state, val) {
            state.allowRouterChange = val
        },

        /**
         * 更新 store 中的 onlineProjectList
         *
         * @param {Object} state store state
         * @param {list} list 项目列表
         */
        forceUpdateOnlineProjectList (state, list) {
            state.sideMenu.onlineProjectList.splice(0, state.sideMenu.onlineProjectList.length, ...list)
        },

        /**
         * 更新 store 中的 menuList
         *
         * @param {Object} state store state
         * @param {list} list menu 列表
         * @param {boolean} isDashboard 是否是 dashboard 路由
         */
        forceUpdateMenuList (state, data) {
            const { list, isDashboard } = data
            if (isDashboard) {
                state.sideMenu.dashboardMenuList.splice(0, state.sideMenu.dashboardMenuList.length, ...list)
            } else if (Boolean(state.curClusterId) && (state.curProject.kind === PROJECT_K8S || state.curProject.kind === PROJECT_TKE)) {
                state.sideMenu.clusterk8sMenuList.splice(0, state.sideMenu.clusterk8sMenuList.length, ...list)
            } else if (state.curProject && (state.curProject.kind === PROJECT_K8S || state.curProject.kind === PROJECT_TKE)) {
                state.sideMenu.k8sMenuList.splice(0, state.sideMenu.k8sMenuList.length, ...list)
            } else {
                state.sideMenu.menuList.splice(0, state.sideMenu.menuList.length, ...list)
            }
        },

        /**
         * 更新 store 中的 menuList
         *
         * @param {Object} state store state
         * @param {list} list menu 列表
         */
        forceUpdateDevOpsMenuList (state, list) {
            state.sideMenu.devOpsMenuList.splice(0, state.sideMenu.devOpsMenuList.length, ...list)
        },

        /**
         * 更新crdInstanceList
         * @param {Object} state store state
         * @param {Object} data data
         */
        updateCrdInstanceList (state, data) {
            state.crdInstanceList = data
        },

        /**
         * 功能开关
         * @param {*} state
         * @param {*} data
         */
        setFeatureFlag (state, data) {
            state.featureFlag = data || {}
        },
        updateViewMode (state, mode) {
            state.viewMode = mode
        },
        updateCurMenuId (state, id) {
            state.curMenuId = id
        },
        updateMenuList (state, menu = []) {
            state.menuList = menu
        }
    },
    actions: {
        /**
         * 获取用户信息
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        userInfo (context, params, config = {}) {
            // return http.get(`/app/index?invoke=userInfo`, {}, config)
            return http.get(`${DEVOPS_BCS_API_URL}/api/user/`, params, config).then(response => {
                const userData = response.data || {}
                context.commit('updateUser', userData)
                return userData
            })
        },

        /**
         * 根据 user bg info
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getUserBgInfo (context, params, config = {}) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/accounts/user_bg_info/`)
        },

        /**
         * 获取项目列表
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getProjectList (context, params, config = {}) {
            return http.get(DEVOPS_BCS_API_URL + '/api/authorized_projects/', params, config).then(response => {
                const data = response.data || []
                context.commit('forceUpdateOnlineProjectList', data)
                return data
            })
        },

        /**
         * 根据项目 id 查询项目的权限
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getProjectPerm (context, { projectCode }, config = {}) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/projects/${projectCode}/`)
        },

        /**
         * 获取关联 CC 的列表
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getCCList (context, params = {}, config = {}) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/cc/?${json2Query(params)}`, params, config)
        },

        /**
         * 停用/启用屏蔽
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        editProject (context, params, config = {}) {
            const projectId = params.project_id
            return http.put(`${DEVOPS_BCS_API_URL}/api/projects/${projectId}/`, params, config)
        },

        /**
         * 获取项目信息
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getProject (context, params, config = {}) {
            const projectId = params.projectId
            return http.get(`${DEVOPS_BCS_API_URL}/api/projects/${projectId}/`, params, config)
        },

        /**
         * 项目启用日志采集功能
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        enableLogPlans (context, projectId, config = {}) {
            return http.post(
                `${DEVOPS_BCS_API_URL}/api/datalog/projects/${projectId}/log_plans/`,
                {},
                config
            )
        },

        /**
         * 获取项目日志采集信息
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getLogPlans (context, projectId, config = {}) {
            return http.get(
                `${DEVOPS_BCS_API_URL}/api/datalog/projects/${projectId}/log_plans/`,
                {},
                config
            )
        },

        /**
         * 查询crd列表 (新)
         *
         * @param {Object} context store 上下文对象
         * @param {Object} projectId, clusterId, crdKind
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getBcsCrdsList (context, { projectId, clusterId, crdKind, params = {} }, config = {}) {
            context.commit('updateCrdInstanceList', [])
            const url = `${DEVOPS_BCS_API_URL}/api/bcs_crd/projects/${projectId}/clusters/${clusterId}/crds/${crdKind}/custom_objects/?${json2Query(params)}`
            return http.get(url, {}, config).then(res => {
                context.commit('updateCrdInstanceList', res.data)
                return res
            })
        },

        async getFeatureFlag (context) {
            const params = {}
            const curCluster = context.state.cluster.curCluster
            // eslint-disable-next-line camelcase
            if (curCluster?.cluster_id) {
                params.cluster_type = curCluster.is_shared ? 'SHARED' : 'SINGLE'
            } else {
                params.$clusterId = '-'
            }

            if (context.state.viewMode === 'dashboard') {
                params.view_mode = 'ResourceDashboard'
            }
            const data = await projectFeatureFlag(params, {
                cancelWhenRouteChange: false
            }).catch(() => ({}))
            context.commit('setFeatureFlag', data)
            return data
        }
    }
})

/**
 * hack vuex dispatch, add third parameter `config` to the dispatch method
 *
 * 需要对单独的请求做配置的话，无论是 get 还是 post，store.dispatch 都需要三个参数，例如：
 * store.dispatch('example/btn1', {btn: 'btn1'}, {fromCache: true})
 * 其中第二个参数指的是请求本身的参数，第三个参数指的是请求的配置，如果请求本身没有参数，那么
 * 第二个参数也必须占位，store.dispatch('example/btn1', {}, {fromCache: true})
 * 在 store 中需要如下写法：
 * btn1 ({commit, state, dispatch}, params, config) {
 *     return http.get(`/app/index?invoke=btn1`, params, config)
 * }
 *
 * @param {Object|string} _type vuex type
 * @param {Object} _payload vuex payload
 * @param {Object} config config 参数，主要指 http 的参数，详见 src/api/index initConfig
 *
 * @return {Promise} 执行请求的 promise
 */
store.dispatch = function (_type, _payload, config = {}) {
    const { type, payload } = unifyObjectStyle(_type, _payload)
    const action = { type, payload, config }
    const entry = store._actions[type]

    if (!entry) {
        if (NODE_ENV !== 'production') {
            console.error(`[vuex] unknown action type: ${type}`)
        }
        return
    }

    store._actionSubscribers.slice().filter(sub => sub.before).forEach(sub => {
        return sub.before(action, store.state)
    })

    return entry.length > 1
        ? Promise.all(entry.map(handler => handler(payload, config)))
        : entry[0](payload, config)
}

export default store
