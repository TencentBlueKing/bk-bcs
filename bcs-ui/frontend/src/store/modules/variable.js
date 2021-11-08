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
        varList: []
    },
    mutations: {
        updateVarList (state, data) {
            state.varList.splice(0, state.varList.length, ...data)
        }
    },
    actions: {
        getVarList (context, projectId, config = {}) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/variables/?limit=10000&offset=0`).then(response => {
                if (response.code === 0) {
                    context.commit('updateVarList', response.data.results)
                }
                return response.data
            })
        },
        getVarListByPage (context, { projectId, offset = 0, limit = 100000, scope = '', keyword = '' }, config = {}) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/variables/?limit=${limit}&offset=${offset}&search_key=${keyword}&scope=${scope}`).then(response => {
                return response.data
            })
        },
        getNamespaceBatchVarList (context, { projectId, variableId }, config = {}) {
            return http.get(
                `${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/variable/batch/namespace/${variableId}/`,
                {},
                config
            )
        },
        getClusterBatchVarList (context, { projectId, variableId }, config = {}) {
            return http.get(
                `${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/variable/batch/cluster/${variableId}/`,
                {},
                config
            )
        },

        getBaseVarList (context, projectId, config = {}) {
            return http.get(
                `${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/variables/?type=base&limit=10000&offset=0`,
                {},
                config
            ).then(response => {
                if (response.code === 0) {
                    context.commit('updateVarList', response.data.results)
                }
                return response.data
            })
        },

        getQuoteDetail (context, { projectId, varId }, config = {}) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/variable/quotes/${varId}/`, {}, config)
        },
        addVar (context, { projectId, data }, config = {}) {
            return http.post(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/variables/`, data, config)
        },
        updateVar (context, { projectId, varId, data }, config = {}) {
            return http.post(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/variable/${varId}/`, data, config)
        },
        updateNamespaceBatchVar (context, { projectId, varId, data }, config = {}) {
            return http.post(
                `${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/variable/batch/namespace/${varId}/`,
                data,
                config
            )
        },
        updateClusterBatchVar (context, { projectId, varId, data }, config = {}) {
            return http.post(
                `${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/variable/batch/cluster/${varId}/`,
                data,
                config
            )
        },
        deleteVar (context, { projectId, data }, config = {}) {
            return http.delete(
                `${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/variables/batch/?id_list=${data.id_list}`,
                {},
                config
            )
        },
        importVars (context, { projectId, data }, config = {}) {
            return http.post(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/variables/batch/`, data, config)
        }
    }
}
