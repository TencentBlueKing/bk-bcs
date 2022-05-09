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
        appList: [],
        tplList: [],
        publicTplList: [],
        privateTplList: []
    },
    mutations: {
        /**
         * 更新App
         *
         * @param {Object} state store state
         * @param {Object} data data
         */
        updateApp (state, data) {
            const appList = state.appList
            appList.forEach((app, index) => {
                if (data.id === app.id) {
                    appList[index] = data
                }
            })
        },
        /**
         * 更新App列表
         *
         * @param {Object} state store state
         * @param {Object} data data
         */
        updateAppList (state, data) {
            state.appList.splice(0, state.appList.length, ...data.results)
        },

        /**
         * 更新所有Html模板列表
         *
         * @param {Object} state store state
         * @param {Object} data data
         */
        updateTplList (state, data) {
            state.tplList.splice(0, state.tplList.length, ...data)
        },

        /**
         * 更新公共Helm模板列表
         *
         * @param {Object} state store state
         * @param {Object} data data
         */
        updatePublicTplList (state, data) {
            state.publicTplList.splice(0, state.publicTplList.length, ...data)
        },

        /**
         * 更新私有Helm模板列表
         *
         * @param {Object} state store state
         * @param {Object} data data
         */
        updatePrivateTplList (state, data) {
            state.privateTplList.splice(0, state.privateTplList.length, ...data)
        }
    },
    actions: {
        /**
         * 获取App状态
         *
         * @param {Object} context store 上下文对象
         * @param {object} params, 包含：projectId, appId
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        checkAppStatus (context, { projectId, appId }, config = {}) {
            // const url = `/app/helm?invoke=checkAppStatus`
            const url = `${DEVOPS_BCS_API_URL}/api/bcs/k8s/configuration/${projectId}/apps/${appId}/transitioning/`
            return http.get(url, {}, { cancelWhenRouteChange: true }).then(res => {
                context.commit('updateApp', res.data)
                return res
            })
        },
        /**
         * 获取App列表
         *
         * @param {Object} context store 上下文对象
         * @param {number} projectId 项目ID
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getAppList (context, { projectId, params }, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/bcs/k8s/configuration/${projectId}/apps/?${json2Query(params)}`
            return http.get(url, {}, config).then(res => {
                context.commit('updateAppList', res.data)
                return res
            })
        },

        /**
         * 获取回滚版本列表
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含projectId, appId
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getRebackList (context, { projectId, appId }, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/bcs/k8s/configuration/${projectId}/apps/${appId}/rollback_selections/`
            return http.get(url, {}, config)
        },

        /**
         * 获取更新版本列表
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含projectId, appId
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getUpdateVersions (context, { projectId, appId }, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/bcs/k8s/configuration/${projectId}/apps/${appId}/upgrade_versions/`
            return http.get(url, {}, config)
        },

        /**
         * 获取更新版本列表 (新)
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含projectId, appId, clusterId
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getUpdateVersionList (context, { projectId, appId, clusterId, namespace }, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/helm_release/projects/${projectId}/clusters/${clusterId}/namespaces/${namespace}/releases/${appId}/versions/`
            return http.get(url, {}, config)
        },

        /**
         * 获取App
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含projectId, appId
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getAppById (context, { projectId, appId }, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/bcs/k8s/configuration/${projectId}/apps/${appId}/`
            return http.get(url, {}, config)
        },

        /**
         * 获取版本对应的app
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含projectId, appId, version
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getUpdateChartByVersion (context, { projectId, appId, version }, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/bcs/k8s/configuration/${projectId}/apps/${appId}/update_chart_versions/${version}/`
            return http.get(url, {}, config)
        },

        /**
         * 获取版本对应的app (新)
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含projectId, appId, clusterId
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getUpdateChartVersionDetail (context, { projectId, appId, clusterId, version, namespace }, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/helm_release/projects/${projectId}/clusters/${clusterId}/namespaces/${namespace}/releases/${appId}/versions/`
            return http.post(url, { version: version }, config)
        },

        /**
         * 获取版本对应的chart
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含projectId, chartId，version
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getChartByVersion (context, { projectId, chartId, version }, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/bcs/k8s/configuration/${projectId}/helm/charts/${chartId}/versions/${version}/`
            return http.get(url, {}, config)
        },

        /**
         * 获取版本对应的chart (新)
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含projectId, chartId, version
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getChartVersionDetail (context, { projectId, chartId, version, isPublic }, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/helm_chart/projects/${projectId}/charts/${chartId}/versions/${version}/`
            return http.post(url, { is_public_repo: isPublic }, config)
        },

        /**
         * 更新App
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含projectId, params
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        updateApp (context, { projectId, appId, params }, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/bcs/k8s/configuration/${projectId}/apps/${appId}/`
            return http.put(url, params, config)
        },

        /**
         * 删除App
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含projectId, appId
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        deleteApp (context, { projectId, appId }, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/bcs/k8s/configuration/${projectId}/apps/${appId}/`
            return http.delete(url, {}, config)
        },

        /**
         * 预览App
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含projectId, appId, params
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        previewApp (context, { projectId, appId, params }, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/bcs/k8s/configuration/${projectId}/apps/${appId}/release_preview/`
            return http.post(url, params)
        },

        /**
         * 回滚应用
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含projectId, appId, params
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        reback (context, { projectId, appId, params }, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/bcs/k8s/configuration/${projectId}/apps/${appId}/rollback/`
            return http.put(url, params, config)
        },

        /**
         * 获取模板列表
         *
         * @param {Object} context store 上下文对象
         * @param {number} projectId 项目ID
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getTplList (context, projectId, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/bcs/k8s/configuration/${projectId}/helm/charts/`
            return http.get(url, {}, config).then(res => {
                const publicRepo = []
                const privateRepo = []

                // 区分是私有还是公有的模板
                res.data.forEach(item => {
                    if (item.repository.name === 'public-repo') {
                        publicRepo.push(item)
                    } else {
                        privateRepo.push(item)
                    }
                })
                context.commit('updateTplList', res.data)
                context.commit('updatePublicTplList', publicRepo)
                context.commit('updatePrivateTplList', privateRepo)
                return res
            })
        },

        /**
         * 获取模板列表
         *
         * @param {Object} context store 上下文对象
         * @param {number} projectId 项目ID
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        asyncGetTplList (context, projectId, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/bcs/k8s/configuration/${projectId}/helm/charts/`
            return http.get(url, {}, config)
        },

        /**
         * 获取模板版本列表
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含projectId, tplId
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getTplVersions (context, { projectId, tplId }, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/bcs/k8s/configuration/${projectId}/helm/charts/${tplId}/versions/`
            return http.get(url, {}, config)
        },

        /**
         * 获取模板版本列表（新）
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含projectId, tplId
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getTplVersionList (context, { projectId, tplId, isPublic }, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/helm_chart/projects/${projectId}/charts/${tplId}/versions/?is_public_repo=${isPublic}`
            return http.get(url, {}, config)
        },

        /**
         * 获取命名空间列表
         *
         * @param {Object} context store 上下文对象
         * @param {object} 包括：projectId, params
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getNamespaceList (context, { projectId, params = {} }, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/bcs/k8s/configuration/${projectId}/namespaces/?${json2Query(params)}`
            return http.get(url, {}, config)
        },

        /**
         * 创建应用
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含projectId, data
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        createApp (context, { projectId, data }, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/bcs/k8s/configuration/${projectId}/apps/`
            return http.post(url, data, config)
        },

        /**
         * 回滚预览
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含projectId, appId, params
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        previewReback (context, { projectId, appId, params }, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/bcs/k8s/configuration/${projectId}/apps/${appId}/rollback_preview/`
            return http.post(url, params, config)
        },

        /**
         * 创建应用预览
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含projectId, params
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        previewCreateApp (context, { projectId, data }, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/bcs/k8s/configuration/${projectId}/apps/create_preview/`
            return http.post(url, data, config)
        },

        /**
         * 通过接口将json转yaml
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含json, yaml
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        syncJsonToYaml (context, { json, yaml }, config = {}) {
            const data = {
                dict: json,
                yaml: yaml
            }
            const url = `${DEVOPS_BCS_API_URL}/api/bcs/k8s/tools/sync_dict2yaml/`
            return http.post(url, data, config)
        },

        /**
         * 同步helm模板仓库
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含projectId
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        syncHelmTpl (context, { projectId }, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/bcs/k8s/configuration/${projectId}/helm/repositories/sync/`
            return http.post(url, {}, config)
        },

        /**
         * 获取markdown文档
         *
         * @param {Object} context store 上下文对象
         * @param {number} projectId 项目ID
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getQuestionsMD (context, projectId, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/bcs/k8s/configuration/${projectId}/how-to-push-helm-chart/`
            return http.get(url, {}, config)
        },

        /**
         * 获取仓库信息
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含projectId
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getHelmDeops (context, { projectId }, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/bcs/k8s/configuration/${projectId}/helm/repositories/lists/detailed`
            return http.get(url, {}, config)
        },

        /**
         * 获取集群信息
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含：projectId, clusterId
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getClusterInfo (context, { projectId, clusterId }, config = {}) {
            const url = `${DEVOPS_BCS_API_URL}/api/bcs/k8s/configuration/${projectId}/container/registry/domian/?cluster_id=${clusterId}`
            return http.get(url, {}, config)
        },

        /**
         * 获取应用信息
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含：projectId, appId
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getAppInfo (context, { projectId, appId }, config = {}) {
            // const url = `/app/helm?invoke=getAppInfo`
            const url = `${DEVOPS_BCS_API_URL}/api/bcs/k8s/configuration/${projectId}/apps/${appId}/status/?format=json`
            return http.get(url, {}, config)
        },

        /**
         * 删除Chart时获取 release
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        getExistReleases (context, { projectId, chartName, versions }, config = {}) {
            return http.post(`${DEVOPS_BCS_API_URL}/api/projects/${projectId}/helm/charts/${chartName}/releases/`, { version_list: versions })
        },

        /**
         * 删除Chart
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {string} chartName
         *
         * @return {Promise} promise 对象
         */
        removeTemplate (context, { chartName, projectId, versions }, config = {}) {
            return http.delete(`${DEVOPS_BCS_API_URL}/api/projects/${projectId}/helm/charts/${chartName}/`, {
                data: {
                    version_list: versions
                }
            }, config)
        },

        /**
         * 获取 app notes
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {string} clusterId
         * @param {string} namespaceName
         * @param {string} releaseName
         *
         * @return {Promise} promise 对象
         */
        getNotes (context, { clusterId, projectId, namespaceName, releaseName }, config = {}) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/helm/projects/${projectId}/clusters/${clusterId}/namespaces/${namespaceName}/releases/${releaseName}/notes/`, {}, config)
        }

    }
}
