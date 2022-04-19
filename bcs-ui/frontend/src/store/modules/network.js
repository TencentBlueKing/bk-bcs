/**
 * Tencent is pleased to support the open source community by making 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition) available.
 * Copyright (C) 2017-2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

import http from '@/api'
import { json2Query } from '@/common/util'

export default {
    namespaced: true,
    state: {
        loadBalanceList: [],
        cloudLoadBalanceList: [],
        nameSpaceList: [],
        nameSpaceClusterList: [],
        clusterList: [],
        serviceList: [],
        endpoints: [],
        curLoadBalance: null
    },
    mutations: {
        /**
         * 更新loadbalance 列表
         *
         * @param {Object} state store state
         * @param {Object} data data
         */
        updateLoadBalanceList (state, data) {
            state.loadBalanceList.splice(0, state.loadBalanceList.length, ...data)
        },

        /**
         * 更新cloudLoadBalanceList 列表
         *
         * @param {Object} state store state
         * @param {Object} data data
         */
        updateCloudLoadBalanceList (state, data) {
            state.cloudLoadBalanceList.splice(0, state.cloudLoadBalanceList.length, ...data)
        },

        /**
         * 更新namespace 列表
         *
         * @param {Object} state store state
         * @param {Object} data data
         */
        updateNameSpaceList (state, data) {
            state.nameSpaceList.splice(0, state.nameSpaceList.length, ...data)
        },

        /**
         * 更新nameSpaceClusterList列表
         *
         * @param {Object} state store state
         * @param {Object} data data
         */
        updateNameSpaceClusterList (state, data) {
            state.nameSpaceClusterList.splice(0, state.nameSpaceClusterList.length, ...data)
        },

        /**
         * 更新集群列表
         *
         * @param {Object} state store state
         * @param {Object} data data
         */
        updateClusterList (state, data) {
            state.clusterList.splice(0, state.clusterList.length, ...data)
        },

        /**
         * 更新serviceList
         *
         * @param {Object} state store state
         * @param {Object} data data
         */
        updateServiceList (state, data) {
            data.forEach(item => {
                const internal = item.access_info ? item.access_info.internal : {}
                const internalKeys = Object.keys(internal)
                const accessInternal = []
                if (internalKeys.length) {
                    internalKeys.forEach(key => {
                        const arr = []
                        const val = internal[key]
                        val.forEach(v => {
                            arr.push(`${v}`)
                        })
                        accessInternal.push(...arr)
                    })
                } else {
                    accessInternal.push('--')
                }
                item.accessInternal = accessInternal

                const external = item.access_info ? item.access_info.external : {}
                const externalKeys = Object.keys(external)
                const accessExternal = []
                if (externalKeys.length) {
                    externalKeys.forEach(key => {
                        const arr = []
                        const val = external[key]
                        val.forEach(v => {
                            arr.push(`${key} ${v}`)
                        })
                        accessExternal.push(...arr)
                    })
                } else {
                    accessExternal.push('--')
                }
                item.accessExternal = accessExternal
            })
            state.serviceList.splice(0, state.serviceList.length, ...data)
        },

        /**
         * 更新单个service
         *
         * @param {Object} state store state
         * @param {Object} data data
         */
        updateServiceInList (state, data) {
            const results = []
            for (let item of state.serviceList) {
                if (item._id === data._id) {
                    item = data
                }
                results.push(item)
            }
            state.serviceList.splice(0, state.serviceList.length, ...results)
        },

        /**
         * 更新endpoints
         *
         * @param {Object} state store state
         * @param {Object} data data
         */
        updateEndpoints (state, data) {
            state.endpoints.splice(0, state.endpoints.length, ...data)
        },

        /**
         * 更新当前lb
         *
         * @param {Object} state store state
         * @param {Object} data data
         */
        updateCurLoadBalance (state, data) {
            state.curLoadBalance = data
        },

        /**
         * 更新单个lb
         *
         * @param {Object} state store state
         * @param {Object} data data
         */
        updateLoadBalanceInList (state, data) {
            const results = []
            for (let item of state.loadBalanceList) {
                if (item.id === data.id) {
                    item = data
                }
                results.push(item)
            }
            state.loadBalanceList.splice(0, state.loadBalanceList.length, ...results)
        },

        /**
         * 更新单个clb
         *
         * @param {Object} state store state
         * @param {Object} data data
         */
        updateCloudLoadBalanceInList (state, data) {
            const results = []
            for (let item of state.cloudLoadBalanceList) {
                if (item.id === data.id) {
                    item = data
                }
                results.push(item)
            }
            state.cloudLoadBalanceList.splice(0, state.cloudLoadBalanceList.length, ...results)
        }
    },
    actions: {
        /**
         * 获取loadbalance 列表
         *
         * @param {Object} context store 上下文对象
         * @param {Object} curProject 当前项目
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getLoadBalanceList (context, { project, params }, config = {}) {
            if (!project) {
                return false
            }
            const projectId = project.project_id
            // let url = `${DEVOPS_BCS_API_URL}/api/network/${projectId}/lb/`
            // // k8s
            // if (project.kind === PROJECT_K8S) {
            //     url = `${DEVOPS_BCS_API_URL}/api/network/${projectId}/k8s/lb/`
            // }

            const url = `${DEVOPS_BCS_API_URL}/api/network/${projectId}/k8s/lb/?${json2Query(params)}`

            return http.get(url).then(res => {
                context.commit('updateLoadBalanceList', res.data)
                return res
            })
        },

        /**
         * 获取ports
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含projectId, version, apps
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getPortsByApps (context, { projectId, version, apps }, config = {}) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/ports/${version}/?app_ids=${apps.join(',')}`, {}, config)
        },

        /**
         * 获取lb
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含projectId, clusterId
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getLoadBalanceByNamespace (context, { projectId, clusterId }, config = {}) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/configuration/projects/${projectId}/clusters/${clusterId}/lbs/`, {}, config)
        },

        /**
         * 获取第一页面loadbalance 列表
         *
         * @param {Object} context store 上下文对象
         * @param {Object} curProject 当前项目
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getLoadBalanceListByPage (context, { project, params }, config = {}) {
            if (!project) {
                return false
            }
            const projectId = project.project_id
            const url = `${DEVOPS_BCS_API_URL}/api/network/${projectId}/k8s/lb/?${json2Query(params)}`
            return http.get(url).then(res => {
                // mesos和k8s的接口格式不一样，处理兼容
                if (res.data && res.data.results) {
                    res.results = res.data.results
                }
                context.commit('updateLoadBalanceList', res.results)
                return res
            })
        },

        /**
         * 删除 loadBalance
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含：projectId, loadBalanceId, projectKind
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        removeLoadBalance (context, { projectId, loadBalanceId, projectKind }, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/network/${projectId}/k8s/lb/${loadBalanceId}/`

            return http.delete(url)
        },

        /**
         * 停止 loadBalance
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含：projectId, loadBalanceId
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        stopBcsLoadBalance (context, { projectId, loadBalanceId }, config = {}) {
            return http.delete(`${DEVOPS_BCS_API_URL}/api/network/${projectId}/lb/bcs/${loadBalanceId}/`, {}, config)
        },

        /**
         * 获取loadBalance状态
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含：projectId, loadBalanceId
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getLoadBalanceStatus (context, { projectId, loadBalanceId }, config = {}) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/network/${projectId}/lb/detail/${loadBalanceId}/`).then(res => {
                context.commit('updateLoadBalanceInList', res.data)
                return res
            })
        },

        /**
         * 后台创建 loadBalance
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        createBcsLoadBalance (context, { projectId, loadBalanceId }, config = {}) {
            return http.post(`${DEVOPS_BCS_API_URL}/api/network/${projectId}/lb/${loadBalanceId}/`, {}, config)
        },

        /**
         * 添加 loadBalance
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含：projectId, data
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        addLoadBalance (context, { projectId, data }, config = {}) {
            return http.post(`${DEVOPS_BCS_API_URL}/api/network/${projectId}/lb/`, data, config)
        },

        /**
         * 添加 loadK8sBalance
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含：projectId, data
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        addK8sLoadBalance (context, { projectId, data }, config = {}) {
            return http.post(`${DEVOPS_BCS_API_URL}/api/network//${projectId}/k8s/lb/`, data, config)
        },

        /**
         * 更新loadBalance
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含：projectId, data, loadBalanceId, projectKind
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        updateLoadBalance (context, { projectId, data, loadBalanceId, projectKind }, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/network/${projectId}/k8s/lb/${loadBalanceId}/`

            return http.put(url, data, config)
        },

        /**
         * 获取taskgroup
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含：projectId, loadBalanceId
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getTaskGroup (context, { projectId, loadBalanceId }, config = {}) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/network/${projectId}/lb/taskgroup/${loadBalanceId}/`, {}, config)
        },

        /**
         * 获取loadbalance 详情
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含：projectId, loadBalanceId, projectKind
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getLoadBalanceDetail (context, { projectId, loadBalanceId, projectKind }, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/network/${projectId}/k8s/lb/${loadBalanceId}/`

            return http.get(url, {}, config)
        },

        /**
         * 获取namespace 列表
         *
         * @param {Object} context store 上下文对象
         * @param {Object} projectId 项目ID
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getNameSpaceList (context, projectId, config = {}) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/namespace/`, {}, config).then(res => {
                context.commit('updateNameSpaceList', res.data)
                return res
            })
        },

        /**
         * 获取集群对应的namespace列表
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含：projectId, clusterId
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getNameSpaceClusterList (context, { projectId, clusterId }, config = {}) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/namespace/?cluster_id=${clusterId}&perm_can_use=1`, {}, config).then(res => {
                context.commit('updateNameSpaceClusterList', res.data.results || [])
                return res
            })
        },

        /**
         * 获取serviceList
         *
         * @param {Object} context store 上下文对象
         * @param {Object} projectId 项目ID
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getServiceList (context, { projectId, params = {} }, config = {}) {
            // return http.get('/app/network?invoke=getServiceList', {}, config).then(res => {
            //     commit('updateServiceList', res.data.data)
            //     return res
            // })
            return http.get(`${DEVOPS_BCS_API_URL}/api/network/${projectId}/services/?${json2Query(params)}`, {}, config).then(res => {
                context.commit('updateServiceList', res.data)
                return res
            })
        },

        /**
         * 删除service
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含：projectId, clusterId, namespace, serviceId
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        deleteService (context, { projectId, clusterId, namespace, serviceId }, config = {}) {
            return http.delete(`${DEVOPS_BCS_API_URL}/api/network/${projectId}/services/clusters/${clusterId}/namespaces/${namespace}/endpoints/${serviceId}/`, {}, config)
        },

        /**
         * 删除services
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含：projectId, data
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        deleteServices (context, { projectId, data }, config = {}) {
            const params = {
                data: data
            }
            return http.post(`${DEVOPS_BCS_API_URL}/api/network/${projectId}/services/batch/`, params, config)
        },

        /**
         * 获取service状态
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含：projectId, namespace, name
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getServiceStatus (context, { projectId, namespace, name }, config = {}) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/network/${projectId}/services/?namespace=${namespace}&name=${name}`, {}, config).then(res => {
                context.commit('updateServiceInList', res.data.data[0])
                return res
            })
        },

        /**
         * 获取applicationlist
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含：projectId, version
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getApplicationsByVersion (context, { projectId, version }, config = {}) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/apps/${version}/`, {}, config)
        },

        /**
         * 获取service详情
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含：projectId, clusterId, namespace, serviceId
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getServiceDetail (context, { projectId, clusterId, namespace, serviceId }, config = {}) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/network/${projectId}/services/clusters/${clusterId}/namespaces/${namespace}/endpoints/${serviceId}/`, {}, config)
        },

        /**
         * 获取endpoint
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含：projectId, clusterId, namespace, name
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getEndpoints (context, { projectId, clusterId, namespace, name }, config = {}) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/resource/${projectId}/clusters/${clusterId}/namespaces/${namespace}/endpoints/${name}/`).then(res => {
                let eps = []
                if (res.data.length && res.data[0].data.eps) {
                    // mesos
                    eps = res.data[0].data.eps
                } else if (res.data.length && res.data[0].data.subsets) {
                    // k8s
                    res.data[0].data.subsets.forEach(item => {
                        if (item.addresses) {
                            item.addresses.forEach(item => {
                                eps.push(item)
                            })
                        }
                    })
                }
                context.commit('updateEndpoints', eps)
                return res
            })
        },

        /**
         * 保存service
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含：projectId, clusterId, namespace, serviceId, data
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        saveServiceDetail (context, { projectId, clusterId, namespace, serviceId, data }, config = {}) {
            return http.post(`${DEVOPS_BCS_API_URL}/api/network/${projectId}/services/clusters/${clusterId}/namespaces/${namespace}/endpoints/${serviceId}/`, data)
        },

        /**
         *  查询clb名称
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含：projectId, region
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getCloudLoadBalanceNames (context, { projectId, region }, config = {}) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/projects/${projectId}/network/clb/names/?region=${region}`, {}, config)
        },

        /**
         * 查询clb controller列表接口
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含：projectId
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getCloudLoadBalanceList (context, { projectId, params }, config = {}) {
            // const url = '/app/network?invoke=getCloudLoadBalanceList'
            const url = `${DEVOPS_BCS_API_URL}/api/projects/${projectId}/network/clbs/?${json2Query(params)}`
            return http.get(url, {}, config).then(res => {
                context.commit('updateCloudLoadBalanceList', res.data)
                return res
            })
        },

        /**
         * 查询单个clb controller
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含：projectId
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getCloudLoadBalanceDetail (context, { projectId, loadBalanceId }, config = {}) {
            // const url = '/app/network?invoke=getCloudLoadBalanceDetail'
            const url = `${DEVOPS_BCS_API_URL}/api/projects/${projectId}/network/clbs/${loadBalanceId}/`
            return http.get(url, {}, config).then(res => {
                context.commit('updateCloudLoadBalanceInList', res.data)
                return res
            })
        },

        /**
         * 查询单个clb 监听器详情
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含：projectId, loadBalanceId
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getCloudLoadBalanceListener (context, { projectId, loadBalanceId }, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/projects/${projectId}/network/clbs/${loadBalanceId}/status/`
            return http.get(url, {}, config)
        },

        /**
         * 创建clb controller接口
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含：projectId, data
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        addCloudLoadBalance (context, { projectId, data }, config = {}) {
            // const url = '/app/network?invoke=updateCloudLoadBalance'
            const url = `${DEVOPS_BCS_API_URL}/api/projects/${projectId}/network/clbs/`
            return http.post(url, data, config)
        },

        /**
         * 编辑clb contoller
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含：projectId, loadBalanceId, data
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        updateCloudLoadBalance (context, { projectId, loadBalanceId, data }, config = {}) {
            // const url = '/app/network?invoke=updateCloudLoadBalance'
            const url = `${DEVOPS_BCS_API_URL}/api/projects/${projectId}/network/clbs/${loadBalanceId}/`
            return http.put(url, data, config)
        },

        /**
         * 删除clb
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含：projectId, loadBalanceId
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        removeCloudLoadBalance (context, { projectId, loadBalanceId }, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/projects/${projectId}/network/clbs/${loadBalanceId}/`
            return http.delete(url, {}, config)
        },

        /**
         * 创建 CL5 路由
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        createCl5 (context, params, config = {}) {
            const { clusterId, projectId, crdKind } = params
            delete params.projectId
            delete params.clusterId
            delete params.crdKind
            const url = `${DEVOPS_BCS_API_URL}/api/bcs_crd/projects/${projectId}/clusters/${clusterId}/crds/${crdKind}/custom_objects/`
            return http.post(url, params, config)
        },

        /**
         * 获取区域列表
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含：projectId
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getRegions (context, { projectId }, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/projects/${projectId}/network/clb/regions/`
            return http.get(url, {}, config)
        },

        /**
         * 查询chart 版本记录
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getChartVersions (context, { projectId, params = {} }, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/network/projects/${projectId}/chart/versions/?${json2Query(params)}`
            return http.get(url, {}, config)
        },

        /**
         * 查询对应版本values
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getChartDetails (context, { projectId, params }, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/network/projects/${projectId}/chart/versions/-/detail/`
            return http.post(url, params, config)
        }
    }
}
