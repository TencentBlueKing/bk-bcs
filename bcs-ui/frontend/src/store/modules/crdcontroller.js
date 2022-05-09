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
        crdInstanceList: [],
        nameSpaceList: []
    },
    mutations: {
        /**
         * 更新crdInstanceList
         * @param {Object} state store state
         * @param {Object} data data
         */
        updateCrdInstanceList (state, data) {
            state.crdInstanceList = data
        },

        /**
         * 更新namespace 列表
         *
         * @param {Object} state store state
         * @param {Object} data data
         */
        updateNameSpaceList (state, data) {
            state.nameSpaceList.splice(0, state.nameSpaceList.length, ...data)
        }
    },
    actions: {
        /**
         * 获取组件库
         *
         * @param {Object} context store 上下文对象
         * @param {Object} projectId, clusterId
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getCrdControllersByCluster (context, { projectId, clusterId }, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/bcs_crd/projects/${projectId}/clusters/${clusterId}/crd_controllers/`
            return http.get(url, {}, config)
        },

        /**
         * 获取组件库状态
         *
         * @param {Object} context store 上下文对象
         * @param {Object} projectId, clusterId
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getCrdcontrollerStatus (context, { projectId, clusterId, crdcontrollerId }, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/bcs_crd/projects/${projectId}/clusters/${clusterId}/crd_controllers/${crdcontrollerId}/instances/-/detail/`
            return http.get(url, {}, config)
        },

        /**
         * 启用组件
         *
         * @param {Object} context store 上下文对象
         * @param {Object} projectId, clusterId, name
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        enableCrdController (context, { projectId, clusterId, id, data }, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/bcs_crd/projects/${projectId}/clusters/${clusterId}/crd_controllers/${id}/instances/`
            return http.post(url, data, config)
        },

        /**
         * 查询crd列表 (老，仅DB授权使用)
         *
         * @param {Object} context store 上下文对象
         * @param {Object} projectId, clusterId, crdKind
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getCrdInstanceList (context, { projectId, clusterId, crdKind, params = {} }, config = {}) {
            context.commit('updateCrdInstanceList', [])
            const url = `${DEVOPS_BCS_API_URL}/api/bcs_crd/projects/${projectId}/clusters/${clusterId}/crds/${crdKind}/custom_objects/?${json2Query(params)}`
            return http.get(url, {}, config).then(res => {
                context.commit('updateCrdInstanceList', res.data)
                return res
            })
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
        getBcsCrdsList (context, { projectId, params = {} }, config = {}) {
            context.commit('updateCrdInstanceList', [])
            const url = `${DEVOPS_BCS_API_URL}/api/bcs_crd/projects/${projectId}/crds/?${json2Query(params)}`
            return http.get(url, {}, config).then(res => {
                context.commit('updateCrdInstanceList', res.data)
                return res
            })
        },

        /**
         * 查询单个crd (仅用于DB授权)
         *
         * @param {Object} context store 上下文对象
         * @param {Object} projectId, clusterId, crdId
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getCrdInstanceDetail (context, { projectId, clusterId, crdKind, crdId }, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/bcs_crd/projects/${projectId}/clusters/${clusterId}/crds/${crdKind}/custom_objects/-/detail/?id=${crdId}`
            return http.get(url, {}, config)
        },

        /**
         * 查询单个crd (日志采集)
         *
         * @param {Object} context store 上下文对象
         * @param {Object} projectId, crdKind, crdId
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getLogCrdInstanceDetail (context, { projectId, crdId, crdKind, clusterId }, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/bcs_crd/projects/${projectId}/clusters/${clusterId}/crds/${crdKind}/custom_objects/-/detail/?id=${crdId}`
            return http.get(url, {}, config)
        },

        /**
         * 查询单个crd (通用)
         *
         * @param {Object} context store 上下文对象
         * @param {Object} projectId, crdKind, crdId
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getCommonCrdInstanceDetail (context, { projectId, crdId, clusterId }, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/bcs_crd/projects/${projectId}/clusters/${clusterId}/crd_controllers/${crdId}/instances/-/detail/`
            return http.get(url, {}, config)
        },

        /**
         * 获取版本列表
         */
        getChartVersionsList (context, { projectId, chartName }, config = {}) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/helm_chart/projects/${projectId}/charts/${chartName}/versions/?is_public_repo=true`)
        },

        /**
         * 获取当前集群下的namespace 列表
         *
         * @param {Object} context store 上下文对象
         * @param {Object} projectId, clusterId
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getNameSpaceListByCluster (context, { projectId, clusterId }, config = {}) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/namespace/?cluster_id=${clusterId}`, {}, config).then(res => {
                context.commit('updateNameSpaceList', res.data)
                return res
            })
        },

        /**
         * 添加crdInstance
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含：projectId, data
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        addCrdInstance (context, { projectId, clusterId, crdKind, data }, config = {}) {
            return http.post(`${DEVOPS_BCS_API_URL}/api/bcs_crd/projects/${projectId}/clusters/${clusterId}/crds/${crdKind}/custom_objects/`, data, config)
        },

        /**
         * 更新crdInstance
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含：projectId, data
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        updateCrdInstance (context, { projectId, clusterId, crdKind, data }, config = {}) {
            return http.put(`${DEVOPS_BCS_API_URL}/api/bcs_crd/projects/${projectId}/clusters/${clusterId}/crds/${crdKind}/custom_objects/`, data, config)
        },

        /**
         * 删除单个crd
         *
         * @param {Object} context store 上下文对象
         * @param {Object} projectId, clusterId, crdId
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        deleteCrdInstance (context, { projectId, clusterId, crdKind, crdId }, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/bcs_crd/projects/${projectId}/clusters/${clusterId}/crds/${crdKind}/custom_objects/`
            return http.delete(url, { data: { id: crdId } }, config)
        },

        /**
         * 更新单个crd
         *
         * @param {Object} context store 上下文对象
         * @param {Object} projectId, clusterId, crdId
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        // clusters/(?P<cluster_id>[\w\-]+)/crd_controllers/(?P<crd_ctr_id>\d+)/instances/-/
        updateCommonCrdInstance (context, { projectId, clusterId, crdId, data }, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/bcs_crd/projects/${projectId}/clusters/${clusterId}/crd_controllers/${crdId}/instances/-/`
            return http.put(url, data, config)
        }
    }
}
