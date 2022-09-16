/* eslint-disable @typescript-eslint/camelcase */
/* eslint-disable camelcase */
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
    },
    mutations: {
    },
    actions: {
        /**
         * 根据项目 id 获取项目下 模板集 集合
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getMusters (context, params, config = {}) {
            const projectId = params.projectId
            delete params.projectId
            return http.get(
                `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/musters/?${json2Query(params)}`,
                {},
                config
            )
        },

        /**
         * 根据 muster id 获取模板集下 模板 集合
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {string} templateId 模板 id
         * @param {string} category 模板 category 属性
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getTemplateList (context, params, config = {}) {
            const projectId = params.projectId
            const tmplMusterId = params.tmplMusterId
            delete params.projectId
            delete params.tmplMusterId
            const url = `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/musters/${tmplMusterId}/templates/`
                + `?${json2Query(params)}`
            return http.get(url, {}, config)
        },

        /**
         * 根据 id 获取 instance 详情
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {string} instanceId instance id
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getInstanceInfo (context, params, config = {}) {
            const projectId = params.projectId
            const instanceId = params.instanceId
            delete params.projectId
            delete params.instanceId
            return http.get(
                `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/instances/${instanceId}/info/?${json2Query(params)}`,
                {},
                config
            )
        },

        /**
         * 重新创建
         * backend_status 为 BackendError 时，重试按钮操作即重新创建
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        reCreate (context, { projectId, instanceId, category }, config = {}) {
            const params = {}
            if (category) {
                params.category = category
            }
            return http.post(
                `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/instances/${instanceId}/retry/`, params, config
            )
        },

        /**
         * 滚动升级获取模板集版本列表
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getTmplsetVerList (context, { projectId, instanceId, category }, config = {}) {
            // return http.get(`/api/app/app?invoke=getTmplsetVerList`
            let url = `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/instances/${instanceId}/versions/`
            if (category) {
                url += `?category=${category}`
            }
            return http.get(url, {}, config)
        },

        /**
         * 暂停滚动升级
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        pauseRollingUpdate (context, params, config = {}) {
            const projectId = params.projectId
            const instanceId = params.instanceId
            const instanceName = params.name

            delete params.projectId
            delete params.instanceId

            return http.put(
                `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/instances/${instanceId}/${instanceName}/pause/`
                    + `?${json2Query(params)}`,
                params,
                config
            )
        },

        /**
         * 取消滚动升级
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        cancelRollingUpdate (context, params, config = {}) {
            const projectId = params.projectId
            const instanceId = params.instanceId
            const instanceName = params.name

            delete params.projectId
            delete params.instanceId

            return http.put(
                `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/instances/${instanceId}/${instanceName}/cancel/`
                    + `?${json2Query(params)}`,
                params,
                config
            )
        },

        /**
         * 恢复滚动升级
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        resumeRollingUpdate (context, params, config = {}) {
            const projectId = params.projectId
            const instanceId = params.instanceId
            const instanceName = params.name

            delete params.projectId
            delete params.instanceId

            return http.put(
                `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/instances/${instanceId}/${instanceName}/resume/`
                    + `?${json2Query(params)}`,
                params,
                config
            )
        },

        /**
         * 扩缩容
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        scaleInstanceNum (context, params, config = {}) {
            // return http.put(`/api/app/app?invoke=scaleInstanceNum`
            const { projectId, instanceId, instanceNum } = params
            const instanceName = params.name

            delete params.projectId
            delete params.instanceId
            delete params.instanceNum

            return http.put(
                `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/instances/${instanceId}/${instanceName}/scale/`
                    + `?instance_num=${instanceNum}&${json2Query(params)}`,
                params,
                config
            )
        },

        /**
         * 删除
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        deleteInstance (context, params, config = {}) {
            const projectId = params.projectId
            const instanceId = params.instanceId
            const instanceName = params.name

            delete params.projectId
            delete params.instanceId

            return http.delete(
                `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/instances/${instanceId}/${instanceName}/delete/?`
                    + `${json2Query(params)}`,
                { data: params },
                config
            )
        },

        /**
         * 强制删除
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        forceDeleteInstance (context, params, config = {}) {
            const projectId = params.projectId
            const instanceId = params.instanceId
            const instanceName = params.name

            delete params.projectId
            delete params.instanceId

            return http.delete(
                `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/instances/${instanceId}/${instanceName}/delete/?`
                    + `enforce=1&${json2Query(params)}`,
                { data: params },
                config
            )
        },

        /**
         * 实例详情页面获取 taskgroup 信息
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getTaskgroupList (context, params, config = {}) {
            const projectId = params.projectId
            const instanceId = params.instanceId
            delete params.projectId
            delete params.instanceId
            return http.get(
                `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/instances/${instanceId}/taskgroups/`
                    + `?${json2Query(params)}`,
                {},
                config
            )
        },

        /**
         * 实例详情页面获取 taskgroup 下的 container 集合
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getContainterList (context, params, config = {}) {
            const projectId = params.projectId
            const instanceId = params.instanceId
            const taskgroupName = params.taskgroupName
            delete params.projectId
            delete params.instanceId
            delete params.taskgroupName
            return http.get(
                `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/instances/${instanceId}/`
                    + `taskgroups/${taskgroupName}/containers/?${json2Query(params)}`,
                {},
                config
            )
        },

        /**
         * 实例详情页面获取 taskgroup 下的 container 集合的日志链接
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getContaintersLogLinks (context, params, config = {}) {
            const projectId = params.projectId
            delete params.projectId

            return http.post(
                `${DEVOPS_BCS_API_URL}/api/datalog/projects/${projectId}/log_links/`,
                params,
                config
            )
        },

        /**
         * 实例详情页面获取 containerIds
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getContainerIds (context, params, config = {}) {
            const projectId = params.projectId
            const instanceId = params.instanceId
            delete params.projectId
            delete params.instanceId
            return http.get(
                `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/instances/${instanceId}/containers/?${json2Query(params)}`,
                {},
                config
            )
        },

        /**
         * 实例详情页面获取 taskgroup 详情
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getTaskgroupInfo (context, params, config = {}) {
            const projectId = params.projectId
            const instanceId = params.instanceId
            const taskgroupName = params.taskgroupName
            delete params.projectId
            delete params.instanceId
            delete params.taskgroupName
            return http.get(
                `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/instances/`
                    + `${instanceId}/taskgroups/${taskgroupName}/info/?${json2Query(params)}`,
                {},
                config
            )
        },

        /**
         * 实例详情页面 taskgroup 重新调度
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        reschedulerTaskgroup (context, params, config = {}) {
            const projectId = params.projectId
            const instanceId = params.instanceId
            const taskgroupName = params.taskgroupName
            delete params.projectId
            delete params.instanceId
            delete params.taskgroupName
            return http.put(
                `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/instances/${instanceId}/rescheduler/?${json2Query(params)}`,
                {
                    taskgroup: taskgroupName
                },
                config
            )
        },

        /**
         * 实例详情页面 taskgroup 里获取容器日志
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getContainterLog (context, params, config = {}) {
            const projectId = params.projectId
            const containerId = params.containerId
            delete params.projectId
            delete params.containerId
            return http.get(
                `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/containers/${containerId}/logs/?${json2Query(params)}`,
                {},
                config
            )
        },

        /**
         * 实例详情页面获取标签信息
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getLabelList (context, params, config = {}) {
            const projectId = params.projectId
            const instanceId = params.instanceId
            const instanceName = params.instanceName
            delete params.projectId
            delete params.instanceId
            delete params.instanceName

            return http.get(
                `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/instances/${instanceId}/${instanceName}/labels/?${json2Query(params)}`,
                {},
                config
            )
        },

        /**
         * 实例详情页面获取备注信息
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getAnnotationList (context, params, config = {}) {
            const projectId = params.projectId
            const instanceId = params.instanceId
            const instanceName = params.instanceName
            delete params.projectId
            delete params.instanceId
            delete params.instanceName
            return http.get(
                `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/instances/`
                    + `${instanceId}/${instanceName}/annotations/?${json2Query(params)}`,
                {},
                config
            )
        },

        /**
         * 实例详情页面获取 metric 信息
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getMetricList (context, params, config = {}) {
            const projectId = params.projectId
            const instanceId = params.instanceId
            delete params.projectId
            delete params.instanceId

            return http.get(
                `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/instances/${instanceId}/metric/?${json2Query(params)}`,
                {},
                config
            )
        },

        /**
         * 实例详情页面获取事件信息
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         *
         * @return {Promise} promise 对象
         */
        getEventList (context, params, config = {}) {
            const projectId = params.projectId
            const instanceId = params.instanceId
            delete params.projectId
            delete params.instanceId
            return http.get(
                `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/instances/${instanceId}/events/?${json2Query(params)}`,
                {},
                config
            )
        },

        /**
         * 实例详情页面获取容器图表信息，所有容器
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         *
         * @return {Promise} promise 对象
         */
        getAllContainerMetrics (context, params, config = {}) {
            const projectId = params.projectId
            delete params.projectId

            const category = params.category
            delete params.category
            let url = `${DEVOPS_BCS_API_URL}/api/projects/${projectId}/metrics/docker/`
            if (category) {
                url += `?category=${category}`
            }

            return http.post(url, params, config)
        },

        /**
         * 容器详情页面获取信息 k8s
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         *
         * @return {Promise} promise 对象
         */
        getContainerInfoK8s (context, params, config = {}) {
            const projectId = params.projectId
            const instanceId = params.instanceId
            const taskgroupName = params.taskgroupName
            const containerId = params.containerId
            delete params.projectId
            delete params.instanceId
            delete params.taskgroupName
            delete params.containerId
            return http.post(
                `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/instances/${instanceId}/`
                    + `taskgroups/${taskgroupName}/containers/info/?${json2Query(params)}`,
                {
                    container_id: containerId,
                    category: params.category
                },
                config
            )
        },

        /**
         * 容器详情页面获取信息 mesos
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         *
         * @return {Promise} promise 对象
         */
        getContainerInfoMesos (context, params, config = {}) {
            const projectId = params.projectId
            const instanceId = params.instanceId
            const taskgroupName = params.taskgroupName
            const containerId = params.containerId
            delete params.projectId
            delete params.instanceId
            delete params.taskgroupName
            delete params.containerId
            return http.get(
                `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/instances/${instanceId}/`
                    + `taskgroups/${taskgroupName}/containers/${containerId}/info/?${json2Query(params)}`,
                {},
                config
            )
        },

        /**
         * 容器详情页面获取图表信息，单个容器
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getContainerMetrics (context, { projectId, containerId, metric, category }, config = {}) {
            let url = `${DEVOPS_BCS_API_URL}/api/projects/${projectId}/metrics/docker/`
                + `?res_id=${containerId}&metric=${metric}`
            if (category) {
                url += `&category=${category}`
            }

            return http.get(url, {}, config)
        },

        /**
         * 实例详情页面获取 strategy 数据
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         *
         * @return {Promise} promise 对象
         */
        getStrategy (context, { projectId, instanceId, strategy }, config = {}) {
            // return http.get(`/api/app/app?invoke=getStrategy&strategy=${strategy}`).then(
            return http.get(
                `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/instances/${instanceId}/strategys/`
                    + `?strategy=${strategy}`,
                {},
                config
            )
        },

        /**
         * 实例详情页面 to json
         * mesos tojson
         * k8s toyaml
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         *
         * @return {Promise} promise 对象
         */
        toJson (context, params, config = {}) {
            const projectId = params.projectId
            const instanceId = params.instanceId
            delete params.projectId
            delete params.instanceId
            // return http.get(`/api/app/app?invoke=toJson&${json2Query(params)}`
            return http.get(
                `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/instances/${instanceId}/configs/?${json2Query(params)}`,
                {},
                config
            )
        },

        /**
         * 应用列表命名空间视图获取命名空间数据
         * {{host}}/api/app/projects/b37778ec757544868a01e1f01f07037f/namespaces/?category=deployment&exist_app=0
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getNamespaces (context, params, config = {}) {
            const projectId = params.projectId
            delete params.projectId

            // return http.get(`/api/app/app?invoke=getNamespaces&${json2Query(params)}`)
            return http.get(
                `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/namespaces/?${json2Query(params)}`,
                {},
                config
            )
        },

        /**
         * 应用列表命名空间视图展开命名空间获取命名空间下的应用列表
         * "^api/app/projects/(?P<project_id>[\w\-]+)/namespaces/(?P<ns_id>\d+)/instances/$"
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getAppListInNamespaceViewMode (context, params, config = {}) {
            const projectId = params.projectId
            const namespaceId = params.namespaceId
            delete params.projectId
            delete params.namespaceId

            const url = `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/namespaces/${namespaceId}/instances/`
                + `?with_count=1&${json2Query(params)}`
            return http.get(url, {}, config)
        },

        /**
         * 根据项目 id 获取项目下 app 集合
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {string} templateId 模板 id
         * @param {string} category 模板 category 属性
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getInstanceList (context, params, config = {}) {
            const projectId = params.projectId
            const tmplMusterId = params.tmplMusterId
            const templateId = params.templateId
            delete params.projectId
            delete params.tmplMusterId
            // return http.get(`/api/app/app?invoke=getInstanceList&${templateId}`
            return http.get(
                `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/musters/${tmplMusterId}/templates/${templateId}/`
                    + `instances/?with_count=1&${json2Query(params)}`,
                {},
                config
            )
        },

        /**
         * 滚动更新获取版本信息
         * {{host}}/api/app/projects/01b6ad17aafc49dcb5cb1aa3d6ee6e01/instances/314/version_conf/?show_version_id=71
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getVersionInRollingUpdate (context, { projectId, instanceId, showVersionId }, config = {}) {
            let url = `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/instances/${instanceId}/version_conf/`
            // 如果存在 showVersionId，表示查询指定的版本
            // 如果不存在，表示查询老版本
            if (showVersionId) {
                url += `?show_version_id=${showVersionId}`
            }
            // url = `/api/app/app?invoke=getVersionInRollingUpdate`
            return http.get(url, {}, config)
        },

        /**
         * 非 template 创建的滚动更新获取版本信息
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getVersionInRollingUpdateInNotPlatform (context, params, config = {}) {
            const projectId = params.projectId
            const instanceId = params.instanceId

            delete params.projectId
            delete params.instanceId

            return http.get(
                `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/instances/${instanceId}/version_conf/?${json2Query(params)}`,
                {},
                config
            )
        },

        /**
         * 滚动更新获取所有版本信息
         * {{host}}/api/app/projects/01b6ad17aafc49dcb5cb1aa3d6ee6e01/instances/314/all_versions/
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getAllVersionInRollingUpdate (context, { projectId, instanceId, category }, config = {}) {
            let url = `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/instances/${instanceId}/all_versions/`
            if (category) {
                url += `?category=${category}`
            }
            return http.get(url, {}, config)
        },

        /**
         * 滚动升级
         * "^api/app/projects/(?P<project_id>[\w\-]+)/instances/(?P<instance_id>[\w\-]+)/update/$"
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        rollingUpdate (context, params, config = {}) {
            const projectId = params.projectId
            const instanceId = params.instanceId
            const variable = params.variable

            delete params.projectId
            delete params.instanceId
            delete params.variable

            return http.put(
                `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/instances/${instanceId}/new_update/`
                    + `?${json2Query(params)}`,
                { variable: variable },
                config
            )
        },

        /**
         * 非模板集创建的应用滚动升级
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        rollingUpdateInNotPlatform (context, params, config = {}) {
            const projectId = params.projectId
            const instanceId = params.instanceId
            const conf = params.conf

            delete params.projectId
            delete params.instanceId
            delete params.conf

            return http.put(
                `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/instances/${instanceId}/new_update/`
                    + `?${json2Query(params)}`,
                {
                    conf: conf
                },
                config
            )
        },

        /**
         * mesos application 更新
         * api/app/projects/(?P<project_id>[\w\-]+)/instances/(?P<instance_id>[\w\-]+)/application_update/
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        update4Application (context, { projectId, instanceId, versionId, category, variable }, config = {}) {
            let url = `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/instances/${instanceId}/application_update/`
                + `?version_id=${versionId}`
            if (category) {
                url += `&category=${category}`
            }

            return http.put(url, { variable: variable }, config)
        },

        /**
         * mesos application 更新
         * api/app/projects/(?P<project_id>[\w\-]+)/instances/(?P<instance_id>[\w\-]+)/application_update/
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        update4ApplicationInNotPlatform (context, params, config = {}) {
            const projectId = params.projectId
            const instanceId = params.instanceId
            const conf = params.conf

            delete params.projectId
            delete params.instanceId
            delete params.conf

            return http.put(
                `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/instances/${instanceId}/application_update/`
                    + `?${json2Query(params)}`,
                {
                    conf: conf
                },
                config
            )
        },

        /**
         * 删除
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        batchDeleteInstance (context, params, config = {}) {
            const projectId = params.projectId
            delete params.projectId
            return http.delete(
                `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/instances/batch/`, {
                    data: params
                },
                config
            )
        },

        /**
         * 应用搜索 命名空间名称 数据源
         * {{host}}/api/app/projects/b37778ec757544868a01e1f01f07037f/all_namespaces/?cluster_type=2&category=deployment
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getAllNamespace4AppSearch (context, params, config = {}) {
            // const projectId = params.projectId
            // const clusterType = params.cluster_type
            // const category = params.category
            // let url = `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/all_namespaces/?cluster_type=${clusterType}`
            // if (category) {
            //     url += `&category=${category}`
            // }
            // return http.get(url, {}, config)

            const projectId = params.projectId
            delete params.projectId

            return http.get(
                // `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/all_namespaces/?${json2Query(params)}`,
                `${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/namespace/?${json2Query(params)}`,
                {},
                config
            )
        },

        /**
         * 应用搜索 应用名称 数据源
         * {{host}}/api/app/projects/b37778ec757544868a01e1f01f07037f/all_instances/?cluster_type=2&category=deployment
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getAllInstance4AppSearch (context, params, config = {}) {
            // const projectId = params.projectId
            // const clusterType = params.cluster_type
            // const category = params.category
            // let url = `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/all_instances/?cluster_type=${clusterType}`
            // if (category) {
            //     url += `&category=${category}`
            // }
            // return http.get(url, {}, config)

            const projectId = params.projectId
            delete params.projectId

            return http.get(
                `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/all_instances/?${json2Query(params)}`,
                {},
                config
            )
        },

        /**
         * 应用搜索 模板集名称 数据源
         * {{host}}/api/app/projects/b37778ec757544868a01e1f01f07037f/all_musters/?cluster_type=2&category=deployment
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getAllMuster4AppSearch (context, params, config = {}) {
            // const projectId = params.projectId
            // const clusterType = params.cluster_type
            // const category = params.category
            // let url = `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/all_musters/?cluster_type=${clusterType}`
            // if (category) {
            //     url += `&category=${category}`
            // }
            // return http.get(url, {}, config)

            const projectId = params.projectId
            delete params.projectId

            return http.get(
                `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/all_musters/?${json2Query(params)}`,
                {},
                config
            )
        },

        /**
         * k8s 系统特殊引用方式的环境变量获取
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getEnvInfo (context, params, config = {}) {
            const projectId = params.projectId
            const instanceId = params.instanceId
            const taskgroupName = params.taskgroupName
            const containerId = params.containerId
            delete params.projectId
            delete params.instanceId
            delete params.taskgroupName
            delete params.containerId
            return http.post(
                `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/instances/${instanceId}/`
                    + `pods/${taskgroupName}/containers/env_info/?${json2Query(params)}`,
                {
                    container_id: containerId,
                    category: params.category
                },
                config
            )
        },

        // ------------------------------------------------------------------------------------------------ //

        /**
         * POD CPU使用率
         * /api/metrics/projects/{project_id}/clusters/{cluster_id}/pods/cpu_usage/
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        podCpuUsage (context, params, config = {}) {
            const { projectId, clusterId } = params
            delete params.projectId
            delete params.clusterId

            return http.post(
                `${DEVOPS_BCS_API_URL}/api/metrics/projects/${projectId}/clusters/${clusterId}/pods/cpu_usage/`,
                params.data,
                config
            )
        },

        /**
         * POD 内存使用量
         * /api/metrics/projects/{project_id}/clusters/{cluster_id}/pods/memory_usage/
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        podMemUsage (context, params, config = {}) {
            const { projectId, clusterId } = params
            delete params.projectId
            delete params.clusterId

            return http.post(
                `${DEVOPS_BCS_API_URL}/api/metrics/projects/${projectId}/clusters/${clusterId}/pods/memory_usage/`,
                params.data,
                config
            )
        },

        /**
         * POD网路接收
         * /api/metrics/projects/{project_id}/clusters/{cluster_id}/pods/network_receive/
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        podNetReceive (context, params, config = {}) {
            const { projectId, clusterId } = params
            delete params.projectId
            delete params.clusterId

            return http.post(
                `${DEVOPS_BCS_API_URL}/api/metrics/projects/${projectId}/clusters/${clusterId}/pods/network_receive/`,
                params.data,
                config
            )
        },

        /**
         * POD网路发送
         * /api/metrics/projects/{project_id}/clusters/{cluster_id}/pods/network_transmit/
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        podNetTransmit (context, params, config = {}) {
            const { projectId, clusterId } = params
            delete params.projectId
            delete params.clusterId

            return http.post(
                `${DEVOPS_BCS_API_URL}/api/metrics/projects/${projectId}/clusters/${clusterId}/pods/network_transmit/`,
                params.data,
                config
            )
        },

        /**
         * 容器CPU使用率
         * /api/projects/{project_id}/clusters/{cluster_id}/metrics/container/cpu_usage/?res_id_list={container_id_list}
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        containerCpuUsage (context, params, config = {}) {
            const { projectId, clusterId, pod_name } = params
            delete params.projectId
            delete params.clusterId
            delete params.pod_name

            return http.post(
                `${DEVOPS_BCS_API_URL}/api/metrics/projects/${projectId}/clusters/${clusterId}/pods/${pod_name}/containers/cpu_usage/`,
                params,
                config
            )
        },

        /**
         * POD CPU使用率 容器视图
         * /api/metrics/projects/{project_id}/clusters/{cluster_id}/pods/{pod_name}/containers/cpu_usage/
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        podCpuUsageContainerView (context, params, config = {}) {
            const { projectId, clusterId, pod_name } = params
            delete params.projectId
            delete params.clusterId
            delete params.pod_name

            return http.post(
                `${DEVOPS_BCS_API_URL}/api/metrics/projects/${projectId}/clusters/${clusterId}/pods/${pod_name}/containers/cpu_usage/`,
                params,
                config
            )
        },

        /**
         * POD 内存使用量 容器视图
         * /api/metrics/projects/{project_id}/clusters/{cluster_id}/pods/{pod_name}/containers/memory_usage/
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        podMemUsageContainerView (context, params, config = {}) {
            const { projectId, clusterId, pod_name } = params
            delete params.projectId
            delete params.clusterId
            delete params.pod_name

            return http.post(
                `${DEVOPS_BCS_API_URL}/api/metrics/projects/${projectId}/clusters/${clusterId}/pods/${pod_name}/containers/memory_usage/`,
                params,
                config
            )
        },

        /**
         * 容器磁盘写 容器视图
         * /api/metrics/projects/{project_id}/clusters/{cluster_id}/pods/{pod_name}/containers/disk_write/
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        podDiskWriteContainerView (context, params, config = {}) {
            const { projectId, clusterId, pod_name } = params
            delete params.projectId
            delete params.clusterId
            delete params.pod_name

            return http.post(
                `${DEVOPS_BCS_API_URL}/api/metrics/projects/${projectId}/clusters/${clusterId}/pods/${pod_name}/containers/disk_write/`,
                params,
                config
            )
        },

        /**
         * 容器磁盘读 容器视图
         * /api/metrics/projects/{project_id}/clusters/{cluster_id}/pods/{pod_name}/containers/disk_read/
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        podDiskReadContainerView (context, params, config = {}) {
            const { projectId, clusterId, pod_name } = params
            delete params.projectId
            delete params.clusterId
            delete params.pod_name

            return http.post(
                `${DEVOPS_BCS_API_URL}/api/metrics/projects/${projectId}/clusters/${clusterId}/pods/${pod_name}/containers/disk_read/`,
                params,
                config
            )
        },

        /**
         * 容器CPU使用率限制
         * /api/metrics/projects/{project_id}/clusters/{cluster_id}/pods/-/containers/cpu_limit/
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        containerCpuLimit (context, params, config = {}) {
            const { projectId, clusterId, pod_name } = params
            delete params.projectId
            delete params.clusterId
            delete params.pod_name

            return http.post(
                `${DEVOPS_BCS_API_URL}/api/metrics/projects/${projectId}/clusters/${clusterId}/pods/${pod_name}/containers/cpu_limit/`,
                params,
                config
            )
        },

        /**
         * 容器内存使用量
         * /api/metrics/projects/{project_id}/clusters/{cluster_id}/pods/-/containers/memory_usage/
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        containerMemUsage (context, params, config = {}) {
            const { projectId, clusterId, pod_name } = params
            delete params.projectId
            delete params.clusterId
            delete params.pod_name

            return http.post(
                `${DEVOPS_BCS_API_URL}/api/metrics/projects/${projectId}/clusters/${clusterId}/pods/${pod_name}/containers/memory_usage/`,
                params,
                config
            )
        },

        /**
         * 容器内存限制
         * /api/metrics/projects/{project_id}/clusters/{cluster_id}/pods/0/containers/memory_limit/
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        containerMemLimit (context, params, config = {}) {
            const { projectId, clusterId, pod_name } = params
            delete params.projectId
            delete params.clusterId
            delete params.pod_name

            return http.post(
                `${DEVOPS_BCS_API_URL}/api/metrics/projects/${projectId}/clusters/${clusterId}/pods/${pod_name}/containers/memory_limit/`,
                params,
                config
            )
        },

        /**
         * 容器网路接收
         * /api/projects/{project_id}/clusters/{cluster_id}/metrics/container/network_receive/?res_id_list={container_id_list}
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        containerNetReceive (context, params, config = {}) {
            const { projectId, clusterId } = params
            delete params.projectId
            delete params.clusterId

            return http.get(
                `${DEVOPS_BCS_API_URL}/api/projects/${projectId}/clusters/${clusterId}/metrics/container/network_receive/?${json2Query(params)}`,
                {},
                config
            )
        },

        /**
         * 容器网路发送
         * /api/projects/{project_id}/clusters/{cluster_id}/metrics/container/network_transmit/?res_id_list={container_id_list}
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        containerNetTransmit (context, params, config = {}) {
            const { projectId, clusterId } = params
            delete params.projectId
            delete params.clusterId

            return http.get(
                `${DEVOPS_BCS_API_URL}/api/projects/${projectId}/clusters/${clusterId}/metrics/container/network_transmit/?${json2Query(params)}`,
                {},
                config
            )
        },

        /**
         * 容器磁盘写
         * /api/projects/{project_id}/clusters/{cluster_id}/metrics/container/disk_write/?res_id_list={container_id_list}
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        containerDiskWrite (context, params, config = {}) {
            const { projectId, clusterId, pod_name } = params
            delete params.projectId
            delete params.clusterId
            delete params.pod_name

            return http.post(
                `${DEVOPS_BCS_API_URL}/api/metrics/projects/${projectId}/clusters/${clusterId}/pods/${pod_name}/containers/disk_write/`,
                params,
                config
            )
        },

        /**
         * 容器磁盘读
         * /api/projects/{project_id}/clusters/{cluster_id}/metrics/container/disk_read/?res_id_list={container_id_list}
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        containerDiskRead (context, params, config = {}) {
            const { projectId, clusterId, pod_name } = params
            delete params.projectId
            delete params.clusterId
            delete params.pod_name

            return http.post(
                `${DEVOPS_BCS_API_URL}/api/metrics/projects/${projectId}/clusters/${clusterId}/pods/${pod_name}/containers/disk_read/`,
                params,
                config
            )
        },

        /**
         * 应用回滚上一版本获取当前版本和上一版本信息接口
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getInstanceConfig4RollbackPrevious (context, params, config = {}) {
            const { projectId, instanceId } = params
            delete params.projectId
            delete params.instanceId

            return http.get(
                `${DEVOPS_BCS_API_URL}/api/projects/${projectId}/instances/${instanceId}/all_configs/`,
                {},
                config
            )
        },

        /**
         * 应用回滚上一版本
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        rollbackPrevious (context, params, config = {}) {
            const { projectId, instanceId } = params
            delete params.projectId
            delete params.instanceId

            return http.put(
                `${DEVOPS_BCS_API_URL}/api/projects/${projectId}/instances/${instanceId}/rollback/`,
                params,
                config
            )
        },

        /**
         * 批量重建
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {string} clusterId 集群 id
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        batchRebuild (context, { projectId, data }, config = {}) {
            return http.put(
                `${DEVOPS_BCS_API_URL}/api/projects/${projectId}/pods/reschedule/`,
                data,
                config
            )
        },

        /**
         * 获取 gamestatefulset 集合
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getGameStatefulsetList (context, params, config = {}) {
            return http.get(
                `${DEVOPS_BCS_API_URL}/api/dashboard/projects/${params.projectId}/clusters/${params.clusterId}/crds/${params.gamestatefulsets}/custom_objects/?${json2Query(params.data)}`,
                {},
                config
            )
        },

        /**
         * 获取 gamestatefulset 详细信息
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getGameStatefulsetInfo (context, params, config = {}) {
            return http.get(
                `${DEVOPS_BCS_API_URL}/api/dashboard/projects/${params.projectId}/clusters/${params.clusterId}/crds/${params.gamestatefulsets}/custom_objects/${params.name}/?${json2Query(params.data)}`,
                {},
                config
            )
        },

        /**
         * 删除 gamestatefulset 详细信息
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        deleteGameStatefulsetInfo (context, params, config = {}) {
            return http.delete(
                `${DEVOPS_BCS_API_URL}/api/dashboard/projects/${params.projectId}/clusters/${params.clusterId}/crds/${params.gamestatefulsets}/custom_objects/${params.name}/?${json2Query(params.data)}`,
                {},
                config
            )
        },

        /**
         * 更新 gamestatefulset 详细信息
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        updateGameStatefulsetInfo (context, params, config = {}) {
            return http.patch(
                `${DEVOPS_BCS_API_URL}/api/dashboard/projects/${params.projectId}/clusters/${params.clusterId}/crds/${params.gamestatefulsets}/custom_objects/${params.name}/`,
                params.data,
                config
            )
        },

        /**
         * gamestatefulset 扩缩容
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        scaleGameStatefulsetInfo (context, params, config = {}) {
            return http.patch(
                `${DEVOPS_BCS_API_URL}/api/dashboard/projects/${params.projectId}/clusters/${params.clusterId}/crds/${params.gamestatefulsets}/custom_objects/${params.name}/scale/`,
                params.data,
                config
            )
        },

        /**
         * 删除
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        batchDeleteGameStatefulset (context, params, config = {}) {
            return http.delete(
                `${DEVOPS_BCS_API_URL}/api/dashboard/projects/${params.projectId}/clusters/${params.clusterId}/crds/${params.gamestatefulsets}/custom_objects/`, {
                    data: params.data
                },
                config
            )
        },

        /**
         * 获取 crd 集合
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getCRDList (context, params, config = {}) {
            return http.get(
                `${DEVOPS_BCS_API_URL}/api/dashboard/projects/${params.projectId}/clusters/${params.clusterId}/crds/`,
                {},
                config
            )
        },

        /**
         * 获取版本日志列表
         */
        getVersionsLogList () {
            return http.get(`${DEVOPS_BCS_API_URL}/change_log/`)
        }
    }
}
