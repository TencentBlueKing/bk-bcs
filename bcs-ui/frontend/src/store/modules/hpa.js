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

export default {
    namespaced: true,
    state: {
        HPAList: []
    },
    mutations: {
        /**
         * 更新HPA list
         * @param {Object} state store state
         * @param {Object} data data
         */
        updateHPAList (state, data) {
            state.HPAList = data
        }
    },
    actions: {
        /**
         * 获取HPA list
         *
         * @param {Object} context store 上下文对象
         * @param {String} projectId, projectId
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getHPAList (context, { projectId, clusterId }, config = {}) {
            // 清空上次数据
            context.commit('updateHPAList', [])
            const url = `${DEVOPS_BCS_API_URL}/api/hpa/projects/${projectId}/?cluster_id=${clusterId}`
            return http.get(url, {}, { cancelWhenRouteChange: true }).then(res => {
                const list = res.data || []
                list.forEach(item => {
                    const conditions = item.conditions || []
                    const conditionsLen = conditions.length
                    for (let i = 0; i < conditionsLen; i++) {
                        if (conditions[i].status.toLowerCase() === 'false') {
                            item.needShowConditions = true
                            break
                        }
                    }
                })
                context.commit('updateHPAList', list)
                return res
            })
        },

        /**
         * 批量删除HPA
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数对象，包括projectId，hpa列表
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        batchDeleteHPA (context, { projectId, params }, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/hpa/projects/${projectId}/`
            return http.delete(url, { data: params }, config)
        },

        /**
         * 批量HPA
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数对象，包括projectId，clusterId, namespace, name
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        deleteHPA (context, { projectId, clusterId, namespace, name }, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/hpa/projects/${projectId}/clusters/${clusterId}/namespaces/${namespace}/${name}/`
            return http.delete(url, {}, config)
        }
    }
}
