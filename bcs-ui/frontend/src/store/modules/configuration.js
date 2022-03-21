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
        curTemplate: {
            name: '',
            desc: '',
            permissions: {
                create: true,
                delete: true,
                list: true,
                view: true,
                edit: true,
                use: true
            }
        },
        curVersion: 0,
        curTemplateId: 0,
        curShowVersionId: 0,
        templateList: [],
        applications: [],
        deployments: [],
        services: [],
        secrets: [],
        configmaps: [],
        versionList: [],
        imageList: [],
        isTemplateSaving: false,
        canTemplateBindVersion: false,
        linkApps: [],
        metricList: []
    },
    mutations: {
        clearCurTemplateData (state) {
            // 模板数据
            state.curTemplate = {
                name: '',
                desc: '',
                permissions: {
                    create: true,
                    delete: true,
                    list: true,
                    view: true,
                    edit: true,
                    use: true
                }
            }
            state.curVersion = 0
            state.curTemplateId = 0
            state.curTemplate.id = 0
            state.curTemplate.latest_version_id = 0
            state.curShowVersionId = 0

            // 资源数据
            state.applications.splice(0, state.applications.length, ...[])
            state.deployments.splice(0, state.deployments.length, ...[])
            state.services.splice(0, state.services.length, ...[])
            state.secrets.splice(0, state.secrets.length, ...[])
            state.configmaps.splice(0, state.configmaps.length, ...[])

            // 关联数据
            state.linkApps.splice(0, state.linkApps.length, ...[])
            state.metricList.splice(0, state.metricList.length, ...[])

            // 状态数据
            state.isTemplateSaving = false
            state.canTemplateBindVersion = false
        },
        updateMetricList (state, list) {
            state.metricList.splice(0, state.metricList.length, ...list)
        },
        updateTemplateList (state, list) {
            state.templateList.splice(0, state.templateList.length, ...list)
        },
        updateBindVersion (state, data) {
            state.canTemplateBindVersion = data
        },
        updateLinkApps (state, data) {
            state.linkApps.splice(0, state.linkApps.length, ...data)
        },
        updateIsTemplateSaving (state, data) {
            state.isTemplateSaving = data
        },
        updateCurShowVersionId (state, data) {
            state.curShowVersionId = data
        },
        updateCurTemplateId (state, data) {
            state.curTemplateId = data
            state.curTemplate.id = data
        },
        updateCurTemplate (state, data) {
            state.curTemplate = data
        },
        updateImageList (state, data) {
            state.imageList = data
        },
        updateCurVersion (state, data) {
            state.curVersion = data
            state.curTemplate.latest_version_id = data
        },
        updateVersionList (state, data) {
            if (data && data.forEach) {
                data.forEach(item => {
                    item.isSelected = false
                })
                state.versionList = data
            }
        },
        addVersion (state, data) {
            data.isSelected = false
            state.versionList.unshift(data)
        },
        updateResources (state, data) {
            if (data.application) {
                data.application.forEach(item => {
                    item.isEdited = false
                    if (String(item.id).indexOf('local_') > -1) {
                        item.isEdited = true
                    } else {
                        item.isEdited = false
                    }

                    // 旧数据兼容
                    if (!item.config.webCache) {
                        item.config.webCache = {}
                    }
                    if (!item.config.webCache.metricIdList) {
                        item.config.webCache.isMetric = false
                        item.config.webCache.metricIdList = []
                    }

                    if (!item.config.webCache.logLabelListCache) {
                        item.config.customLogLabel = {}
                        item.config.webCache.logLabelListCache = [
                            {
                                key: '',
                                value: ''
                            }
                        ]
                    }

                    if (!item.config.monitorLevel) {
                        item.config.monitorLevel = 'general'
                    }

                    item.config.spec.template.spec.containers.forEach(container => {
                        if (!container.logListCache) {
                            container.logListCache = [
                                {
                                    value: ''
                                }
                            ]
                        }
                        // 初初始化imageVersion
                        if (container.image && !container.imageVersion) {
                            const arr = container.image.split(':')
                            if (arr.length > 1) {
                                container.imageVersion = arr[arr.length - 1]
                            }
                        }
                    })
                })

                state.applications.splice(0, state.applications.length, ...data.application)
            } else {
                state.applications.splice(0, state.applications.length)
            }

            if (data.deployment) {
                data.deployment.forEach(item => {
                    if (String(item.id).indexOf('local_') > -1) {
                        item.isEdited = true
                    } else {
                        item.isEdited = false
                    }
                    const rollingupdate = item.config.strategy.rollingupdate
                    if (!rollingupdate.hasOwnProperty('rollingManually')) {
                        rollingupdate.rollingManually = false
                    }
                })
                state.deployments.splice(0, state.deployments.length, ...data.deployment)
            } else {
                state.deployments.splice(0, state.deployments.length)
            }

            if (data.service) {
                data.service.forEach(item => {
                    item.serviceIPs = item.config.spec.clusterIP.join(',')
                    if (String(item.id).indexOf('local_') > -1) {
                        item.isEdited = true
                    } else {
                        item.isEdited = false
                    }
                    // 对旧数据兼容
                    if (!item.config.webCache) {
                        item.config.webCache = {}
                    }
                    if (!item.config.webCache.link_app) {
                        item.config.webCache.link_app = []
                        item.config.webCache.link_app_weight = []
                    }
                })
                state.services.splice(0, state.services.length, ...data.service)
            } else {
                state.services.splice(0, state.services.length)
            }

            if (data.configmap) {
                data.configmap.forEach(item => {
                    if (String(item.id).indexOf('local_') > -1) {
                        item.isEdited = true
                    } else {
                        item.isEdited = false
                    }

                    const list = []
                    const keys = item.config.datas
                    if (!item.configmapKeyList) {
                        item.configmapKeyList = []
                    }
                    for (const [key, value] of Object.entries(keys)) {
                        list.push({
                            key: key,
                            type: value.type,
                            isEdit: false,
                            content: value.content
                        })
                    }

                    item.configmapKeyList.splice(0, item.configmapKeyList.length, ...list)
                })

                state.configmaps.splice(0, state.configmaps.length, ...data.configmap)
            } else {
                state.configmaps.splice(0, state.configmaps.length)
            }

            if (data.secret) {
                data.secret.forEach(item => {
                    if (String(item.id).indexOf('local_') > -1) {
                        item.isEdited = true
                    } else {
                        item.isEdited = false
                    }
                    const list = []
                    const keys = item.config.datas
                    if (!item.secretKeyList) {
                        item.secretKeyList = []
                    }
                    for (const [key, value] of Object.entries(keys)) {
                        list.push({
                            key: key,
                            isEdit: false,
                            content: value.content
                        })
                    }
                    this.curKeyIndex = 0
                    if (list.length) {
                        this.curKeyParams = list[0]
                    } else {
                        this.curKeyParams = null
                    }

                    item.secretKeyList.splice(0, item.secretKeyList.length, ...list)
                })
                state.secrets.splice(0, state.secrets.length, ...data.secret)
            } else {
                state.secrets.splice(0, state.secrets.length)
            }
        },
        updateApplications (state, data) {
            state.applications.splice(0, state.applications.length, ...data)
        },
        updateApplicationById (state, { application, preId }) {
            for (const item of state.applications) {
                if (item.id === preId) {
                    item.id = application.id
                }
            }

            const list = JSON.parse(JSON.stringify(state.applications))
            state.applications.splice(0, state.applications.length, ...list)
        },
        updateDeploymentById (state, { deployment, preId }) {
            for (const item of state.deployments) {
                if (item.id === preId) {
                    item.id = deployment.id
                }
            }
            const list = JSON.parse(JSON.stringify(state.deployments))
            state.deployments.splice(0, state.deployments.length, ...list)
        },
        updateServiceById (state, { service, preId }) {
            for (const item of state.services) {
                if (item.id === preId) {
                    item.id = service.id
                }
            }
            const list = JSON.parse(JSON.stringify(state.services))
            state.services.splice(0, state.services.length, ...list)
        },
        updateConfigmapById (state, { configmap, targetData, preId }) {
            for (const item of state.configmaps) {
                if (item.id === preId) {
                    item.config = targetData.config
                    item.id = configmap.id
                }
            }
            const list = JSON.parse(JSON.stringify(state.configmaps))
            state.configmaps.splice(0, state.configmaps.length, ...list)
        },
        updateSecretById (state, { secret, preId }) {
            for (const item of state.secrets) {
                if (item.id === preId) {
                    item.id = secret.id
                }
            }
            const list = JSON.parse(JSON.stringify(state.secrets))
            state.secrets.splice(0, state.secrets.length, ...list)
        },
        updateDeployments (state, data) {
            state.deployments.splice(0, state.deployments.length, ...data)
        },
        updateServices (state, data) {
            if (data.length) {
                data.forEach(item => {
                    // 对旧数据兼容
                    if (!item.config.webCache) {
                        item.config.webCache = {}
                    }
                    if (!item.config.webCache.link_app) {
                        item.config.webCache.link_app = []
                        item.config.webCache.link_app_weight = []
                    }
                })
            }
            state.services.splice(0, state.services.length, ...data)
        },
        updateSecrets (state, data) {
            state.secrets.splice(0, state.secrets.length, ...data)
        },
        updateConfigmaps (state, data) {
            state.configmaps.splice(0, state.configmaps.length, ...data)
        }
    },
    actions: {
        // 获取单个模板集
        getTemplateById (context, { projectId, templateId }, config = {}) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/template/${templateId}/`, {}, config)
        },

        /**
         * 获取所有命名空间
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         *
         * @return {Promise} promise 对象
         */
        getAllNamespaceList (context, params, config = {}) {
            const projectId = params.projectId
            delete params.projectId

            // return http.get(`/app/configuration?invoke=getAllNamespaceList&${json2Query(params)}`, {}, config)

            return http.get(
                `${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/namespace/?${json2Query(params)}`,
                {},
                config
            )
        },

        /**
         * 获取模板集版本
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {string} templateId template id
         *
         * @return {Promise} promise 对象
         */
        getTemplatesetVerList (context, { projectId, templateId, hasFilter }, config = {}) {
            // return http.get(`/api/configuration/configuration?invoke=getTemplatesetVerList`).then(
            let urlPrefix = 'show/versions'
            // 删除实例使用独立的url
            if (hasFilter) {
                urlPrefix = 'exist/show_version_name'
            }
            return http.get(
                `${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/${urlPrefix}/${templateId}/`,
                {},
                config
            )
        },

        /**
         * 根据模板集 id, category, tmpl_app_name 获取模板
         * 用于应用列表跳转到实例化页面
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {string} tplVerId template version id
         * @param {string} tmplAppName template appName
         * @param {string} category category category
         *
         * @return {Promise} promise 对象
         */
        getTemplateListByIdCategoryTmplName (context, { projectId, tplVerId, tmplAppName, category }, config = {}) {
            // return http.get(`/api/configuration/configuration?invoke=getTemplateListByIdCategoryTmplName`
            return http.get(
                `${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/resource/${tplVerId}/`
                    + `?category=${category}&tmpl_app_name=${tmplAppName}`,
                {},
                config
            )
        },

        /**
         * 根据模板集 id 获取模板
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {string} tplVerId template version id
         *
         * @return {Promise} promise 对象
         */
        getTemplateListById (context, { projectId, tplVerId }, config = {}) {
            // return ajax.get(`/api/configuration/configuration?invoke=getTemplateListById`).then(
            return http.get(
                `${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/resource/${tplVerId}/`,
                {},
                config
            )
        },

        /**
         * 根据模板集 id 获取已经被使用过的 namespace
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {string} tplVerId template version id
         *
         * @return {Promise} promise 对象
         */
        getExistNamespace (context, { projectId, tplVerId, instanceEntity }, config = {}) {
            // return http.get(`/api/configuration/configuration?invoke=getExistNamespace`).then(
            return http.post(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/instance/ns/${tplVerId}/`, {
                instance_entity: instanceEntity
            }, config)
        },

        /**
         * 查询 lb 和 variable
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        getLbVariable (context, { projectId, tplVerId, namespaces, instanceEntity }, config = {}) {
            return http.post(
                `${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/variable/resource/${tplVerId}/`,
                {
                    namespaces: namespaces,
                    instance_entity: instanceEntity
                },
                config
            )
        },

        /**
         * 预查询 命名空间 下的 loadbalance 信息览配置
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         *
         * @return {Promise} promise 对象
         */
        getLbInfo (context, { projectId, clusterId }, config = {}) {
            return http.get(
                `${DEVOPS_BCS_API_URL}/api/configuration/projects/${projectId}/clusters/${clusterId}/lbs/`,
                {},
                config
            )
        },

        /**
         * 预览配置
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         *
         * @return {Promise} promise 对象
         */
        previewNamespace (context, params, config = {}) {
            const { projectId } = params
            delete params.projectId
            // return http.get(`/api/configuration/configuration?invoke=previewNamespace`).then(
            return http.post(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/preview/`, params, config)
        },

        /**
         * 模板实例化，创建实例化
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         *
         * @return {Promise} promise 对象
         */
        createInstance (context, params, config = {}) {
            const { projectId } = params
            delete params.projectId
            return http.post(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/instances/`, params, config)
        },

        /**
         * 获取命名空间
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {boolean} isGroupBy 是否需要 group_by
         *
         * @return {Promise} promise 对象
         */
        getNamespaceList (context, { projectId, isGroupBy }, config = {}) {
            return http.get(
                // `${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/namespace/${isGroupBy ? '?group_by=env_type' : ''}`
                `${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/namespace/${isGroupBy ? '?group_by=cluster_name' : ''}`,
                {},
                config
            )
        },

        getNamespaceListByClusterId (context, { projectId, clusterId }, config = {}) {
            return http.get(
                `${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/namespace/${clusterId ? `?cluster_id=${clusterId}` : ''}`,
                {},
                config
            )
        },

        /**
         * 查询命名空间的变量信息
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        getNamespaceVariable (context, { projectId, namespaceId }, config = {}) {
            return http.get(
                `${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/variable/namespace/${namespaceId}/`,
                {},
                config
            )
        },

        /**
         * 添加命名空间
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         *
         * @return {Promise} promise 对象
         */
        addNamespace (context, params, config = {}) {
            const { projectId } = params
            delete params.projectId
            return http.post(
                `${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/namespace/`,
                params,
                config
            )
        },

        /**
         * 修改命名空间
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         *
         * @return {Promise} promise 对象
         */
        editNamespace (context, params, config = {}) {
            const { projectId, namespaceId } = params
            delete params.projectId
            delete params.namespaceId
            return http.put(
                `${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/namespace/${namespaceId}/`,
                params,
                config
            )
        },

        /**
         * 删除命名空间
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         *
         * @return {Promise} promise 对象
         */
        delNamespace (context, params, config = {}) {
            const { projectId, namespaceId } = params
            delete params.projectId
            delete params.namespaceId
            return http.delete(
                `${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/namespace/${namespaceId}/`,
                params,
                config
            )
        },

        /**
         * 同步命名空间
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         *
         * @return {Promise} promise 对象
         */
        syncNamespace (context, params, config = {}) {
            const { projectId } = params
            delete params.projectId
            return http.post(
                `${DEVOPS_BCS_API_URL}/api/projects/${projectId}/configuration/namespaces/sync/`,
                params,
                config
            )
        },

        /**
         * 获取配额数据
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {string} tplVerId template version id
         *
         * @return {Promise} promise 对象
         */
        getQuota (context, { projectId, clusterId, namespaceName }, config = {}) {
            return http.get(
                `${DEVOPS_BCS_API_URL}/api/resources/projects/${projectId}/clusters/${clusterId}/namespaces/${namespaceName}/`,
                {},
                config
            )
        },

        /**
         * 获取配额数据
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {string} tplVerId template version id
         *
         * @return {Promise} promise 对象
         */
        editQuota (context, params, config = {}) {
            const { projectId, clusterId, namespaceName, data } = params
            return http.put(
                `${DEVOPS_BCS_API_URL}/api/resources/projects/${projectId}/clusters/${clusterId}/namespaces/${namespaceName}/`,
                data,
                config
            )
        },

        /**
         * 删除配额
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         *
         * @return {Promise} promise 对象
         */
        delQuota (context, { projectId, clusterId, namespaceName }, config = {}) {
            return http.delete(
                `${DEVOPS_BCS_API_URL}/api/resources/projects/${projectId}/clusters/${clusterId}/namespaces/${namespaceName}/`,
                {},
                config
            )
        }
    }
}
