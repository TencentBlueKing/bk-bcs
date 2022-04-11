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
        // 当前模板基础信息
        curTemplate: {
            name: '',
            desc: '',
            is_locked: false,
            locker: '',
            permissions: {
                create: true,
                delete: true,
                list: true,
                view: true,
                edit: true,
                use: true
            }
        },
        existConfigmapList: [],
        curVersion: 0, // 模板当前版本号
        curTemplateId: 0, // 当前模板ID
        curShowVersionId: 0,
        templateList: [], // 模板集列表
        certListUrl: '', // 证书列表访问url
        certList: [], // 证书列表
        applications: [],
        deployments: [],
        services: [],
        secrets: [],
        configmaps: [],
        daemonsets: [],
        jobs: [],
        statefulsets: [],
        ingresss: [],
        HPAs: [],
        versionList: [], // 模板集版本列表
        imageList: [], // 镜像列表
        isTemplateSaving: false,
        canTemplateBindVersion: false,
        linkApplications: [], // 关联应用列表
        linkServices: [], // 关联services列表
        metricList: [],
        linkApps: []
    },
    mutations: {
        clearCurTemplateData (state) {
            // 模板数据
            state.curTemplate = {
                name: '',
                desc: '',
                is_locked: false,
                locker: '',
                permissions: {
                    create: true,
                    delete: true,
                    list: true,
                    view: true,
                    edit: true,
                    use: true
                }
            }
            state.curTemplate.id = 0
            state.curVersion = 0
            state.curTemplateId = 0
            state.curShowVersionId = 0
            state.curTemplate.latest_version_id = 0

            // 资源数据
            state.deployments.splice(0, state.deployments.length, ...[])
            state.services.splice(0, state.services.length, ...[])
            state.secrets.splice(0, state.secrets.length, ...[])
            state.configmaps.splice(0, state.configmaps.length, ...[])
            state.statefulsets.splice(0, state.statefulsets.length, ...[])
            state.jobs.splice(0, state.jobs.length, ...[])
            state.daemonsets.splice(0, state.daemonsets.length, ...[])
            state.ingresss.splice(0, state.ingresss.length, ...[])
            state.HPAs.splice(0, state.HPAs.length, ...[])

            // 关联数据
            state.linkApplications.splice(0, state.linkApplications.length, ...[])
            state.certList.splice(0, state.certList.length, ...[])
            state.imageList.splice(0, state.imageList.length, ...[])
            state.linkServices.splice(0, state.linkServices.length, ...[])
            state.metricList.splice(0, state.metricList.length, ...[])

            // 状态数据
            state.isTemplateSaving = false
            state.canTemplateBindVersion = false
        },
        updateExistConfigmap (state, list) {
            state.existConfigmapList.splice(0, state.existConfigmapList.length, ...list)
        },
        updateMetricList (state, list) {
            state.metricList.splice(0, state.metricList.length, ...list)
        },
        updateCertList (state, list) {
            state.certList.splice(0, state.certList.length, ...list)
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
        updateLinkApplications (state, data) {
            const list = []
            if (data.Deployment) {
                const apps = {
                    deploy_name: 'Deployment',
                    children: []
                }
                data.Deployment.forEach(item => {
                    apps.children.push(item)
                })
                list.push(apps)
            }
            if (data.DaemonSet) {
                const apps = {
                    deploy_name: 'DaemonSet',
                    children: []
                }
                data.DaemonSet.forEach(item => {
                    apps.children.push(item)
                })
                list.push(apps)
            }
            // if (data.Job) {
            //     const apps = {
            //         deploy_name: 'Job',
            //         children: []
            //     }
            //     data.Job.forEach(item => {
            //         apps.children.push(item)
            //     })
            //     list.push(apps)
            // }
            if (data.StatefulSet) {
                const apps = {
                    deploy_name: 'StatefulSet',
                    children: []
                }
                data.StatefulSet.forEach(item => {
                    apps.children.push(item)
                })
                list.push(apps)
            }
            state.linkApplications.splice(0, state.linkApplications.length, ...list)
        },
        updateLinkServices (state, data) {
            state.linkServices.splice(0, state.linkServices.length, ...data)
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
            if (!data) {
                return false
            }
            if (!data.permissions) {
                data.permissions = {
                    create: true,
                    delete: true,
                    list: true,
                    view: true,
                    edit: true,
                    use: true
                }
            }
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
            function initApp (item) {
                if (String(item.id).indexOf('local_') > -1) {
                    item.isEdited = true
                } else {
                    item.isEdited = false
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
                if (!item.config.webCache.hostAliasesCache) {
                    item.config.webCache.hostAliasesCache = []
                }

                // 解决历史问题
                if (JSON.stringify(item.config.spec.template.spec.nodeSelector) === '{}') {
                    item.config.webCache.nodeSelectorList = [
                        {
                            key: '',
                            value: ''
                        }
                    ]
                }

                // 解决空数据不传key兼容问题
                if (!item.config.spec.template.spec.hasOwnProperty('serviceAccountName')) {
                    item.config.spec.template.spec.serviceAccountName = ''
                }

                if (!item.config.spec.selector) {
                    item.config.spec.selector = {
                        matchLabels: {}
                    }
                }

                if (!item.config.monitorLevel) {
                    item.config.monitorLevel = 'general'
                }

                // deployment兼容接口数据
                if (item.config.spec.hasOwnProperty('minReadySecond') && !item.config.spec.hasOwnProperty('minReadySeconds')) {
                    item.config.spec.minReadySeconds = item.config.spec.minReadySecond
                }

                const spec = item.config.spec.template.spec
                spec.allContainers = []
                spec.containers.forEach(container => {
                    // 初初始化imageVersion
                    if (container.image && !container.imageVersion) {
                        const arr = container.image.split(':')
                        if (arr.length > 1) {
                            container.imageVersion = arr[arr.length - 1]
                        }
                    }

                    // 初始化类型
                    if (container.webCache && !container.webCache.containerType) {
                        container.webCache.containerType = 'container'
                    }

                    // 兼容自定义镜像
                    if (container.webCache && !container.webCache.hasOwnProperty('isImageCustomed')) {
                        // 以DEVOPS_ARTIFACTORY_HOST开头都是通过下拉配置，包括变量形式，否则是自定义
                        container.webCache.isImageCustomed = !container.image.startsWith(`${DEVOPS_ARTIFACTORY_HOST}`)
                    }

                    // 命令下的workingDir兼容
                    if (!container.hasOwnProperty('workingDir')) {
                        container.workingDir = ''
                    }

                    // 兼容资源限制特权
                    if (!container.hasOwnProperty('securityContext')) {
                        container.securityContext = {
                            privileged: false
                        }
                    }

                    spec.allContainers.push(container)
                })

                if (spec.initContainers) {
                    spec.initContainers.forEach(container => {
                        // 初初始化imageVersion
                        if (container.image && !container.imageVersion) {
                            const arr = container.image.split(':')
                            if (arr.length > 1) {
                                container.imageVersion = arr[arr.length - 1]
                            }
                        }

                        // 初始化类型
                        if (container.webCache && !container.webCache.containerType) {
                            container.webCache.containerType = 'initContainer'
                        }

                        // 命令下的workingDir兼容
                        if (!container.hasOwnProperty('workingDir')) {
                            container.workingDir = ''
                        }

                        // initContainers类型，在保存时会删除livenessProbe，readinessProbe，lifecycle
                        if (!container.hasOwnProperty('livenessProbe')) {
                            container.livenessProbe = {
                                httpGet: {
                                    port: '',
                                    path: '',
                                    httpHeaders: []
                                },
                                tcpSocket: {
                                    port: ''
                                },
                                exec: {
                                    command: ''
                                },
                                initialDelaySeconds: 15,
                                periodSeconds: 10,
                                timeoutSeconds: 5,
                                failureThreshold: 3,
                                successThreshold: 1
                            }
                        }

                        if (!container.hasOwnProperty('readinessProbe')) {
                            container.readinessProbe = {
                                httpGet: {
                                    port: '',
                                    path: '',
                                    httpHeaders: []
                                },
                                tcpSocket: {
                                    port: ''
                                },
                                exec: {
                                    command: ''
                                },
                                initialDelaySeconds: 15,
                                periodSeconds: 10,
                                timeoutSeconds: 5,
                                failureThreshold: 3,
                                successThreshold: 1
                            }
                        }

                        if (!container.hasOwnProperty('lifecycle')) {
                            container.lifecycle = {
                                preStop: {
                                    exec: {
                                        command: ''
                                    }
                                },
                                postStart: {
                                    exec: {
                                        command: ''
                                    }
                                }
                            }
                        }

                        item.config.spec.template.spec.allContainers.push(container)
                    })
                } else {
                    spec.initContainers = []
                }
            }

            if (data.K8sDeployment) {
                data.K8sDeployment.forEach(item => {
                    initApp(item)
                })
                state.deployments.splice(0, state.deployments.length, ...data.K8sDeployment)
            } else {
                state.deployments.splice(0, state.deployments.length)
            }

            if (data.K8sDaemonSet) {
                data.K8sDaemonSet.forEach(item => {
                    initApp(item)
                })
                state.daemonsets.splice(0, state.daemonsets.length, ...data.K8sDaemonSet)
            } else {
                state.daemonsets.splice(0, state.daemonsets.length)
            }

            if (data.K8sJob) {
                data.K8sJob.forEach(item => {
                    initApp(item)
                })
                state.jobs.splice(0, state.jobs.length, ...data.K8sJob)
            } else {
                state.jobs.splice(0, state.jobs.length)
            }

            if (data.K8sStatefulSet) {
                data.K8sStatefulSet.forEach(item => {
                    initApp(item)
                })
                state.statefulsets.splice(0, state.statefulsets.length, ...data.K8sStatefulSet)
            } else {
                state.statefulsets.splice(0, state.statefulsets.length)
            }

            if (data.K8sIngress) {
                data.K8sIngress.forEach(item => {
                    if (String(item.id).indexOf('local_') > -1) {
                        item.isEdited = true
                    } else {
                        item.isEdited = false
                    }
                    // 兼容处理
                    if (!item.config.spec.tls) {
                        item.config.spec.tls = [
                            {
                                hosts: '',
                                secretName: ''
                            }
                        ]
                    }
                    item.config.spec.tls.forEach(computer => {
                        if (typeof computer.hosts === 'object') {
                            computer.hosts = computer.hosts.join(',')
                        }
                    })
                })
                state.ingresss.splice(0, state.ingresss.length, ...data.K8sIngress)
            } else {
                state.ingresss.splice(0, state.ingresss.length)
            }

            if (data.K8sHPA) {
                data.K8sHPA.forEach(item => {
                    if (String(item.id).indexOf('local_') > -1) {
                        item.isEdited = true
                    } else {
                        item.isEdited = false
                    }
                })
                state.HPAs.splice(0, state.HPAs.length, ...data.K8sHPA)
            } else {
                state.HPAs.splice(0, state.HPAs.length)
            }

            if (data.K8sService) {
                data.K8sService.forEach(item => {
                    if (String(item.id).indexOf('local_') > -1) {
                        item.isEdited = true
                    } else {
                        item.isEdited = false
                    }
                })
                state.services.splice(0, state.services.length, ...data.K8sService)
            } else {
                state.services.splice(0, state.services.length)
            }
            if (data.K8sConfigMap) {
                data.K8sConfigMap.forEach(item => {
                    if (String(item.id).indexOf('local_') > -1) {
                        item.isEdited = true
                    } else {
                        item.isEdited = false
                    }

                    const list = []
                    const keys = item.config.data
                    if (!item.configmapKeyList) {
                        item.configmapKeyList = []
                    }
                    for (const [key, value] of Object.entries(keys)) {
                        list.push({
                            key: key,
                            isEdit: false,
                            content: value
                        })
                    }

                    item.configmapKeyList.splice(0, item.configmapKeyList.length, ...list)
                })

                state.configmaps.splice(0, state.configmaps.length, ...data.K8sConfigMap)
            } else {
                state.configmaps.splice(0, state.configmaps.length)
            }

            if (data.K8sSecret) {
                data.K8sSecret.forEach(item => {
                    if (String(item.id).indexOf('local_') > -1) {
                        item.isEdited = true
                    } else {
                        item.isEdited = false
                    }
                    const list = []
                    const keys = item.config.data
                    if (!item.secretKeyList) {
                        item.secretKeyList = []
                    }
                    for (const [key, value] of Object.entries(keys)) {
                        list.push({
                            key: key,
                            isEdit: false,
                            content: value
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
                state.secrets.splice(0, state.secrets.length, ...data.K8sSecret)
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
        updateDeploymentById (state, { deployment, preId, targetData }) {
            for (const item of state.deployments) {
                if (item.id === preId) {
                    item.id = deployment.id
                    item.config.spec.template.metadata.labels = targetData.config.spec.template.metadata.labels
                    item.config.spec.selector.matchLabels = targetData.config.spec.selector.matchLabels
                    delete item.cache
                }
            }
            const list = JSON.parse(JSON.stringify(state.deployments))
            state.deployments.splice(0, state.deployments.length, ...list)
        },
        updateDaemonsetById (state, { daemonset, preId, targetData }) {
            for (const item of state.daemonsets) {
                if (item.id === preId) {
                    item.id = daemonset.id
                    item.config.spec.template.metadata.labels = targetData.config.spec.template.metadata.labels
                    item.config.spec.selector.matchLabels = targetData.config.spec.selector.matchLabels
                    delete item.cache
                }
            }
            const list = JSON.parse(JSON.stringify(state.daemonsets))
            state.daemonsets.splice(0, state.daemonsets.length, ...list)
        },
        updateJobById (state, { job, preId, targetData }) {
            for (const item of state.jobs) {
                if (item.id === preId) {
                    item.id = job.id
                    item.config.spec.template.metadata.labels = targetData.config.spec.template.metadata.labels
                    item.config.spec.selector.matchLabels = targetData.config.spec.selector.matchLabels
                    delete item.cache
                }
            }
            const list = JSON.parse(JSON.stringify(state.jobs))
            state.jobs.splice(0, state.jobs.length, ...list)
        },
        updateStatefulsetById (state, { statefulset, preId, targetData }) {
            for (const item of state.statefulsets) {
                if (item.id === preId) {
                    item.id = statefulset.id
                    item.deploy_tag = statefulset.resource_data.deploy_tag
                    item.config.spec.template.metadata.labels = targetData.config.spec.template.metadata.labels
                    item.config.spec.selector.matchLabels = targetData.config.spec.selector.matchLabels
                    delete item.cache
                }
            }
            const list = JSON.parse(JSON.stringify(state.statefulsets))
            state.statefulsets.splice(0, state.statefulsets.length, ...list)
        },
        updateServiceById (state, { service, preId }) {
            for (const item of state.services) {
                if (item.id === preId) {
                    item.id = service.id
                    item.service_tag = service.resource_data.service_tag
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
        updateIngressById (state, { ingress, preId }) {
            for (const item of state.ingresss) {
                if (item.id === preId) {
                    item.id = ingress.id
                }
            }
            const list = JSON.parse(JSON.stringify(state.ingresss))
            state.ingresss.splice(0, state.ingresss.length, ...list)
        },
        updateHPAById (state, { HPA, preId }) {
            for (const item of state.HPAs) {
                if (item.id === preId) {
                    item.id = HPA.id
                }
            }
            const list = JSON.parse(JSON.stringify(state.HPAs))
            state.HPAs.splice(0, state.HPAs.length, ...list)
        },
        updateDeployments (state, data) {
            state.deployments.splice(0, state.deployments.length, ...data)
        },
        updateServices (state, data) {
            state.services.splice(0, state.services.length, ...data)
        },
        updateSecrets (state, data) {
            state.secrets.splice(0, state.secrets.length, ...data)
        },
        updateConfigmaps (state, data) {
            state.configmaps.splice(0, state.configmaps.length, ...data)
        },
        updateDaemonsets (state, data) {
            state.daemonsets.splice(0, state.daemonsets.length, ...data)
        },
        updateJobs (state, data) {
            state.jobs.splice(0, state.jobs.length, ...data)
        },
        updateStatefulsets (state, data) {
            state.statefulsets.splice(0, state.statefulsets.length, ...data)
        },
        updateIngresss (state, data) {
            state.ingresss.splice(0, state.ingresss.length, ...data)
        },
        updateHPAs (state, data) {
            state.HPAs.splice(0, state.HPAs.length, ...data)
        }
    },
    actions: {
        /**
         * 更新templateData
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        updateTemplateDraft (context, { projectId, templateId, data }) {
            return http.post(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/template/${templateId}/draft/`, data)
        },

        /**
         * 获取deploymentLabelList
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        getDeploymentLabelList (context, { projectId, versionId }) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/K8sDeployment/labels/${versionId}/?deploy_tag_list=["1519876530"]`).then(res => {
                if (res && res.data) {
                    context.commit('updateVersionList', res.data)
                } else {
                    context.commit('updateVersionList', [])
                }
                return res
            }, res => {
                context.commit('updateVersionList', [])
            })
        },

        /**
         * 获取version list
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        getVersionList (context, { projectId, templateId }) {
            const url = `${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/template/${templateId}/show/version/`
            return http.get(url).then(res => {
                if (res && res.data) {
                    context.commit('updateVersionList', res.data)
                } else {
                    context.commit('updateVersionList', [])
                }
                return res
            }, res => {
                context.commit('updateVersionList', [])
            })
        },

        /**
         * 添加version
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        addVersion (context, { projectId, templateId, data }) {
            return http.post(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/show/versions/${templateId}/`, data).then(res => {
                return res
            })
        },

        /**
         * 删除version
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        removeVersion (context, { projectId, templateId, versionId }) {
            return http.delete(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/template/${templateId}/show/version/${versionId}/`).then(res => {
                return res
            })
        },

        /**
         * 根据版本号获取template内容
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        getTemplateByVersion (context, { projectId, templateId, versionId }) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/template/${templateId}/show/version/${versionId}/`).then(res => {
                if (res && res.data) {
                    context.commit('updateResources', res.data)
                    context.commit('updateCurShowVersionId', versionId)
                    context.commit('updateCurVersion', res.data.version)
                }
                return res
            })
        },

        /**
         * 获取templateSet详细内容
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        getTemplateSetDetail (context, { projectId, templateId, versionId }) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/template/${templateId}/show/version/${versionId}/`)
        },

        /**
         * 获取ports
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        getPortsByDeployments (context, { projectId, version, apps }) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/configuration/projects/${projectId}/versions/${version}/K8sContainerPorts/?deploy_tag_list=${JSON.stringify(apps)}`, {}, {
                requestId: 'getPortsByDeployments',
                cancelPrevious: true
            })
        },

        /**
         * 获取labels
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        getLabelsByDeployments (context, { projectId, version, apps }) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/K8sDeployment/labels/${version}/?deploy_tag_list=${JSON.stringify(apps)}`, {}, {
                requestId: 'getLabelsByDeployments',
                cancelPrevious: true
            })
        },

        /**
         * 获取模板集列表
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        getTemplateList (context, projectId) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/templates/`)
        },

        /**
         * 获取单个模板集
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        getTemplateById (context, { projectId, templateId }) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/template/${templateId}/`)
        },

        /**
         * 更新template
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        updateTemplate (context, { projectId, templateId, data }) {
            return http.post(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/template/${templateId}/`, data)
        },

        /**
         * 获取application list
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        getApplicationList (context, { projectId, templateId, version }) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/template/${templateId}/version/0/`)
        },

        /**
         * 获取所有资源
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        getTemplateResource (context, { projectId, templateId, version }) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/template/${templateId}/show/version/${version}/`).then(res => {
                context.commit('updateCurShowVersionId', version)
                return res
            })
        },

        /**
         * 删除模板集
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        removeTemplate (context, { templateId, projectId }) {
            return http.delete(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/template/${templateId}/`)
        },

        /**
         * 添加模板集
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        addTemplate (context, { templateParams, projectId }) {
            return http.post(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/templates/`, templateParams)
        },

        /**
         * 获取应用标签关联情况
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        getApplicationLinkLabels (context, { projectId, versionId }) {
            if (!versionId) {
                return []
            }
            const url = `${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/labels/${versionId}/`
            return http.get(url)
        },

        /**
         * 添加新application
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        addApplication (context, { data, version, projectId }) {
            return http.put(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/version/${version}/K8sDeployment/0/`, data)
        },

        /**
         * 更新application
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        updateApplication (context, { data, version, projectId, applicationId }) {
            return http.put(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/version/${version}/K8sDeployment/${applicationId}/`, data)
        },

        /**
         * 添加第一个application
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        addFirstApplication (context, { data, templateId, projectId }) {
            return http.post(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/template/${templateId}/K8sDeployment/`, data)
        },

        /**
         * 删除application
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        removeApplication (context, { applicationId, version, projectId }) {
            return http.delete(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/version/${version}/application/${applicationId}/`)
        },

        /**
         * 拉取镜像列表
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        getImageList (context, { projectId }) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/depot/available/images/${projectId}/`)
        },

        /**
         * 拉取某个镜像版本
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        getImageVertionList (context, { projectId, imageId, isPub }) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/depot/available/tags/${projectId}/?repo=${imageId}&is_pub=${isPub}`)
        },

        /**
         * 获取对应version 的deployment  list
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        getDeploymentsByVersion (context, { projectId, version }) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/K8sDeployment/${version}/`).then(res => {
                if (res.data.length) {
                    context.commit('updateLinkApps', res.data)
                }
                return res
            })
        },

        getAppsByVersion (context, { projectId, version }) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/pods/${version}/`).then(res => {
                if (res.data) {
                    context.commit('updateLinkApplications', res.data)
                }
                return res
            })
        },

        /**
         * 获取对应version 的service list
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        getServicesByVersion (context, { projectId, version }) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/K8sService/${version}/`).then(res => {
                if (res.data.length) {
                    context.commit('updateLinkServices', res.data)
                }
                return res
            })
        },

        /**
         * 获取对应version 的daemonset list
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        getDaemonsetsByVersion (context, { projectId, version }) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/K8sDaemonSet/${version}/`).then(res => {
                if (res.data.length) {
                    context.commit('updateLinkApps', res.data)
                }
                return res
            })
        },

        /**
         * 获取对应version 的job list
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        getJobsByVersion (context, { projectId, version }) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/K8sJob/${version}/`).then(res => {
                if (res.data.length) {
                    context.commit('updateLinkApps', res.data)
                }
                return res
            })
        },

        /**
         * 获取对应version 的statefulset  list
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        getStatefulsetsByVersion (context, { projectId, version }) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/K8sStatefulSet/${version}/`).then(res => {
                if (res.data.length) {
                    context.commit('updateLinkApps', res.data)
                }
                return res
            })
        },

        /**
         * 创建deployment
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        addDeployment (context, { data, version, projectId }) {
            return http.put(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/version/${version}/K8sDeployment/0/`, data)
        },

        /**
         * 添加第一个Deployment
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        addFirstDeployment (context, { data, templateId, projectId }) {
            return http.post(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/template/${templateId}/K8sDeployment/`, data)
        },

        /**
         * 更新deployment
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        updateDeployment (context, { data, version, projectId, id }) {
            return http.put(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/version/${version}/K8sDeployment/${id}/`, data)
        },

        /**
         * 删除deployment
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        removeDeployment (context, { id, version, projectId }) {
            return http.delete(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/version/${version}/K8sDeployment/${id}/`)
        },

        /**
         * 创建Daemonset
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        addDaemonset (context, { data, version, projectId }) {
            return http.put(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/version/${version}/K8sDaemonSet/0/`, data)
        },

        /**
         * 添加第一个Daemonset
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        addFirstDaemonset (context, { data, templateId, projectId }) {
            return http.post(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/template/${templateId}/K8sDaemonSet/`, data)
        },

        /**
         * 更新Daemonset
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        updateDaemonset (context, { data, version, projectId, id }) {
            return http.put(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/version/${version}/K8sDaemonSet/${id}/`, data)
        },

        /**
         * 删除Daemonset
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        removeDaemonset (context, { id, version, projectId }) {
            return http.delete(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/version/${version}/K8sDaemonSet/${id}/`)
        },

        /**
         * 创建Job
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        addJob (context, { data, version, projectId }) {
            return http.put(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/version/${version}/K8sJob/0/`, data)
        },

        /**
         * 添加第一个Job
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        addFirstJob (context, { data, templateId, projectId }) {
            return http.post(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/template/${templateId}/K8sJob/`, data)
        },

        /**
         * 更新Job
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        updateJob (context, { data, version, projectId, id }) {
            return http.put(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/version/${version}/K8sJob/${id}/`, data)
        },

        /**
         * 删除Job
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        removeJob (context, { id, version, projectId }) {
            return http.delete(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/version/${version}/K8sJob/${id}/`)
        },

        /**
         * 创建Statefulset
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        addStatefulset (context, { data, version, projectId }) {
            return http.put(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/version/${version}/K8sStatefulSet/0/`, data)
        },

        /**
         * 添加第一个Statefulset
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        addFirstStatefulset (context, { data, templateId, projectId }) {
            return http.post(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/template/${templateId}/K8sStatefulSet/`, data)
        },

        /**
         * 更新Statefulset
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        updateStatefulset (context, { data, version, projectId, id }) {
            return http.put(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/version/${version}/K8sStatefulSet/${id}/`, data)
        },

        /**
         * 删除Statefulset
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        removeStatefulset (context, { id, version, projectId }) {
            return http.delete(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/version/${version}/K8sStatefulSet/${id}/`)
        },

        /**
         * 创建service
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        addService (context, { data, version, projectId }) {
            return http.put(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/version/${version}/K8sService/0/`, data)
        },

        /**
         * 更新service
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        updateService (context, { data, version, projectId, serviceId }) {
            return http.put(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/version/${version}/K8sService/${serviceId}/`, data)
        },

        /**
         * 删除Service
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        removeService (context, { serviceId, version, projectId }) {
            return http.delete(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/version/${version}/K8sService/${serviceId}/`)
        },

        /**
         * 添加第一个service
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        addFirstService (context, { data, templateId, projectId }) {
            return http.post(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/template/${templateId}/K8sService/`, data)
        },

        /**
         * 创建ingress
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        addIngress (context, { data, version, projectId }) {
            return http.put(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/version/${version}/K8sIngress/0/`, data)
        },

        /**
         * 更新ingress
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        updateIngress (context, { data, version, projectId, ingressId }) {
            return http.put(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/version/${version}/K8sIngress/${ingressId}/`, data)
        },

        /**
         * 删除ingress
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        removeIngress (context, { ingressId, version, projectId }) {
            return http.delete(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/version/${version}/K8sIngress/${ingressId}/`)
        },

        /**
         * 添加第一个ingress
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        addFirstIngress (context, { data, templateId, projectId }) {
            return http.post(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/template/${templateId}/K8sIngress/`, data)
        },

        /**
         * 获取HPA metric类型
         *
         * @param {Object} context store 上下文对象
         * @param {Number} projectId project id
         *
         * @return {Promise} promise 对象
         */
        getHPAMetric (context, projectId) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/hpa/projects/${projectId}/metrics/`)
        },

        /**
         * 创建HPA
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 包括data, version, projectId
         *
         * @return {Promise} promise 对象
         */
        addHPA (context, { data, version, projectId }) {
            return http.put(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/version/${version}/K8sHPA/0/`, data)
        },

        /**
         * 更新HPA
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 包括data, version, projectId, HPAId
         *
         * @return {Promise} promise 对象
         */
        updateHPA (context, { data, version, projectId, HPAId }) {
            return http.put(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/version/${version}/K8sHPA/${HPAId}/`, data)
        },

        /**
         * 删除HPA
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 包括version, projectId, HPAId
         *
         * @return {Promise} promise 对象
         */
        removeHPA (context, { HPAId, version, projectId }) {
            return http.delete(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/version/${version}/K8sHPA/${HPAId}/`)
        },

        /**
         * 添加第一个HPA
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 包括data, templateId, HPAId
         *
         * @return {Promise} promise 对象
         */
        addFirstHPA (context, { data, templateId, projectId }) {
            return http.post(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/template/${templateId}/K8sHPA/`, data)
        },

        /**
         * 创建configmap
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        addConfigmap (context, { data, version, projectId }) {
            return http.put(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/version/${version}/K8sConfigMap/0/`, data)
        },

        /**
         * 更新configmap
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        updateConfigmap (context, { data, version, projectId, configmapId }) {
            return http.put(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/version/${version}/K8sConfigMap/${configmapId}/`, data)
        },

        /**
         * 删除Configmap
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        removeConfigmap (context, { configmapId, version, projectId }) {
            return http.delete(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/version/${version}/K8sConfigMap/${configmapId}/`)
        },

        /**
         * 添加第一个configmap
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        addFirstConfigmap (context, { data, templateId, projectId }) {
            return http.post(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/template/${templateId}/K8sConfigMap/`, data)
        },

        /**
         * 创建secret
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        addSecret (context, { data, version, projectId }) {
            return http.put(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/version/${version}/K8sSecret/0/`, data)
        },

        /**
         * 更新secret
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        updateSecret (context, { data, version, projectId, secretId }) {
            return http.put(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/version/${version}/K8sSecret/${secretId}/`, data)
        },

        /**
         * 删除Secret
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        removeSecret (context, { secretId, version, projectId }) {
            return http.delete(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/version/${version}/K8sSecret/${secretId}/`)
        },

        /**
         * 添加第一个secret
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        addFirstSecret (context, { data, templateId, projectId }) {
            return http.post(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/template/${templateId}/K8sSecret/`, data)
        },

        /**
         * 拉取挂载卷/环境变量configmaps
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        getConfigmaps (context, { projectId, version }) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/K8sConfigMap/${version}/`)
        },

        /**
         * 保存版本
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        saveVersion (context, { projectId, templateId, params }) {
            return http.post(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/template/${templateId}/show/version/`, params)
        },

        /**
         * 拉取挂载卷/环境变量secrets
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        getSecrets (context, { projectId, version }) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/K8sSecret/${version}/`)
        },

        /**
         * check 端口是否已经关联
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        checkPortIsLink (context, { projectId, version, portId }) {
            if (projectId && version && portId) {
                return http.get(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/K8sDeployment/check/version/${version}/port/${portId}/`)
            }
        },

        /**
         * 拉取metric 列表
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        getMetricList (context, projectId) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/metric/${projectId}/`).then(res => {
                context.commit('updateMetricList', res.data)
                return res
            })
        },

        /**
         * 获取模板集版本ss
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        getTemplatesetVerList (context, { projectId, templateId, hasFilter }) {
            // return http.get(`/api/configuration/configuration?invoke=getTemplatesetVerList`).then(
            let urlPrefix = 'show/versions'
            // 删除实例使用独立的url
            if (hasFilter) {
                urlPrefix = 'exist/show_version_name'
            }
            return http.get(
                `${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/${urlPrefix}/${templateId}/`
            )
        },

        /**
         * 根据模板集 版本下已经实例化过的资源
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        getTemplateInsResourceById (context, { projectId, templateId, showVerName }) {
            // return http.get(`/api/configuration/configuration?invoke=getTemplateListById`).then(
            return http.get(
                `${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/exist/resource/${templateId}/?show_version_name=${showVerName}`
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
        getTemplateListById (context, { projectId, tplVerId, hasFilter }) {
            // return http.get(`/api/configuration/configuration?invoke=getTemplateListById`).then(
            return http.get(
                `${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/resource/${tplVerId}/`
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
        getTemplateListByIdCategoryTmplName (context, { projectId, tplVerId, tmplAppName, category }) {
            // return http.get(`/api/configuration/configuration?invoke=getTemplateListByIdCategoryTmplName`
            return http.get(
                `${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/resource/${tplVerId}/`
                    + `?category=${category}&tmpl_app_name=${tmplAppName}`
            )
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
        getNamespaceList (context, { projectId, isGroupBy }) {
            // return http.get(`/api/configuration/configuration?invoke=${isGroupBy
            //     ? 'getNamespaceList' : 'getNamespaceListNoGroupBy'}`
            return http.get(
                `${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/namespace/${isGroupBy ? '?group_by=env_type' : ''}`
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
        previewNamespace (context, params) {
            const { projectId } = params
            delete params.projectId
            // return http.get(`/api/configuration/configuration?invoke=previewNamespace`).then(
            return http.post(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/preview/`, params)
        },

        /**
         * 预览配置
         * 用于应用列表跳转到实例化页面
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         *
         * @return {Promise} promise 对象
         */
        previewNamespace1 (context, params) {
            const { projectId } = params
            delete params.projectId
            return http.put(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/app/instantiation/`, params)
        },

        /**
         * 模板实例化，创建实例化
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         *
         * @return {Promise} promise 对象
         */
        createInstance (context, params) {
            const { projectId } = params
            delete params.projectId
            return http.post(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/instances/`, params)
        },

        /**
         * 模板实例化，创建实例化
         * 用于应用列表跳转到实例化页面
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         *
         * @return {Promise} promise 对象
         */
        createInstance1 (context, params) {
            const { projectId } = params
            delete params.projectId
            return http.post(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/app/instantiation/`, params)
        },

        /**
         * 预查询 命名空间 下的 loadbalance 信息览配置
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         *
         * @return {Promise} promise 对象
         */
        getLbInfo (context, { projectId, namespaceId }) {
            // return http.get(`/api/configuration/configuration?invoke=getLbInfo`).then(
            return http.get(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/loadbalance/${namespaceId}/`)
        },

        /**
         * 添加命名空间
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         *
         * @return {Promise} promise 对象
         */
        addNamespace (context, params) {
            const { projectId } = params
            delete params.projectId
            return http.post(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/namespace/`, params)
        },

        /**
         * 修改命名空间
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         *
         * @return {Promise} promise 对象
         */
        editNamespace (context, params) {
            const { projectId, namespaceId } = params
            delete params.projectId
            delete params.namespaceId
            return http.put(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/namespace/${namespaceId}/`, params)
        },

        /**
         * 删除命名空间
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         *
         * @return {Promise} promise 对象
         */
        delNamespace (context, params) {
            const { projectId, namespaceId } = params
            delete params.projectId
            delete params.namespaceId
            // return http.delete(`/api/configuration/configuration?invoke=delNamespace`, params).then(
            //     response => response.data
            // )
            return http.delete(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/namespace/${namespaceId}/`, params)
        },

        /**
         * 删除实例时的获取命名空间集合接口
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        getNamespaceList4DelInstance (context, { projectId, tplMusterId, tplsetVerId, tplId, category }) {
            // return http.get(`/api/configuration/configuration?invoke=getNamespaceList4DelInstance`
            return http.get(`${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/musters/${tplMusterId}/instances/`
                    + `namespaces/?group_by=env_type&category=${category}&`
                    + `show_version_name=${tplsetVerId}&res_name=${tplId}`
            )
        },

        /**
         * 反实例化删除命名空间
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         *
         * @return {Promise} promise 对象
         */
        delNamespaceInDelInstance (context, params) {
            const { projectId, tplMusterId, tplsetVerId, tplId } = params
            delete params.projectId
            delete params.tplMusterId
            delete params.tplsetVerId
            delete params.tplId

            // return http.delete(
            //     `/api/configuration/configuration?invoke=delNamespaceInDelInstance`,
            //     {data: params}
            // ).then(
            return http.delete(`${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/musters/${tplMusterId}/`
                    + `instances/resources/?show_version_name=${tplsetVerId}&res_name=${tplId}`, { data: params })
        },

        /**
         * 复制模板集
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         *
         * @return {Promise} promise 对象
         */
        copyTemplate (context, params) {
            const { projectId, templateId } = params
            delete params.projectId
            delete params.templateId
            // return http.get(`/api/configuration/configuration?invoke=copyTemplate`, params).then(
            return http.put(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/template/${templateId}/`, params)
        },

        /**
         * 删除模板集时获取 exist_version
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
        getExistVersion (context, { projectId, templateId }) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/template/exist_versions/${templateId}/`)
        },

        /**
         * 加锁
         *
         * @param {Object} context store 上下文对象
         * @param {Object} 请求参数，包括projectId, templateId
         */
        lockTemplateset (context, { projectId, templateId }) {
            return http.post(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/template/lock/${templateId}/`)
        },

        /**
         * 解锁
         *
         * @param {Object} context store 上下文对象
         * @param {Object} 请求参数，包括projectId, templateId
         */
        unlockTemplateset (context, { projectId, templateId }) {
            return http.post(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/template/unlock/${templateId}/`)
        },

        /**
         * 获取已经存在的configmap信息(已排除[kube-system, kube-public, thanos])
         *
         * @param {Object} context store 上下文对象
         * @param {Object} 请求参数，包括projectId
         */
        getExistConfigmap (context, { projectId }) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/resource/projects/${projectId}/configmap/exist/list/`)
        },

        /**
         * statefulset关联service
         *
         * @param {Object} context store 上下文对象
         * @param {Object} 请求参数，包括projectId
         */
        bindServiceForStatefulset (context, { projectId, versionId, statefulsetId, data }) {
            return http.put(`${DEVOPS_BCS_API_URL}/api/configuration/projects/${projectId}/versions/${versionId}/K8sStatefulSet/${statefulsetId}/service-tag/`, data, { cancelWhenRouteChange: false })
        },

        /**
         * 创建yaml模板资源
         *
         * @param {Object} context store 上下文对象
         * @param {Object} 请求参数，包括projectId及发到接口数据
         */
        createYamlTemplate (context, { projectId, data }) {
            return http.post(`${DEVOPS_BCS_API_URL}/api/projects/${projectId}/configuration/yaml_templates/`, data)
        },

        /**
         * 更新yaml模板资源
         *
         * @param {Object} context store 上下文对象
         * @param {Object} 请求参数，包括projectId及发到接口数据
         */
        updateYamlTemplate (context, { projectId, templateId, data }) {
            return http.put(`${DEVOPS_BCS_API_URL}/api/projects/${projectId}/configuration/yaml_templates/${templateId}/`, data, {
                cancelWhenRouteChange: false
            })
        },

        /**
         * 获取yaml模板资源
         *
         * @param {Object} context store 上下文对象
         * @param {Object} 请求参数，包括projectId、templaetId
         */
        getYamlTemplateDetail (context, { projectId, templateId }) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/projects/${projectId}/configuration/yaml_templates/${templateId}/`)
        },

        /**
         * 获取yaml模板资源
         *
         * @param {Object} context store 上下文对象
         * @param {Object} 请求参数，包括projectId、templaetId
         */
        getYamlTemplateDetailByVersion (context, { projectId, templateId, versionId, withFileContent = true }) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/projects/${projectId}/configuration/yaml_templates/${templateId}/show_versions/${versionId}/?with_file_content=${withFileContent}`)
        },

        /**
         * 获取yaml模板 release, 用于实例化时作preview效果
         *
         * @param {Object} context store 上下文对象
         * @param {Object} 请求参数，包括projectId、templaetId
         */
        createYamlTemplateReleases (context, { projectId, templateId, versionId, data }) {
            return http.post(`${DEVOPS_BCS_API_URL}/api/projects/${projectId}/configuration/yaml_templates/${templateId}/show_versions/${versionId}/releases/`, data)
        },

        /**
         * 获取yaml模板支持的资源
         *
         * @param {Object} context store 上下文对象
         * @param {Object} 请求参数，projectId
         */
        getYamlResources (context, { projectId }) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/projects/${projectId}/configuration/yaml_templates/initial_templates/`)
        }
    }
}
