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

import http from '@open/api'

export default {
    namespaced: true,
    state: {
    },
    mutations: {
    },
    actions: {
        /**
         * 获取 Storage list
         *
         * @param {Object} context store 上下文对象
         * @param {String} projectId, projectId
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getList (context, { projectId, clusterId, idx }, config = {}) {
            // idx: pv, pvc, sc
            return http.get(
                `${DEVOPS_BCS_API_URL}/api/projects/${projectId}/clusters/${clusterId}/storages/?res_kind=${idx}`,
                {},
                config
            )
        }
    }
}
