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
    state: {},
    mutations: {},
    actions: {
        /**
         * 获取 metric 集合
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getMetricList (context, { projectId }, config = {}) {
            // return http.get(`/app/metric?invoke=getMetricList`, {}, config)
            return http.get(`${DEVOPS_BCS_API_URL}/api/metric/${projectId}/`, {}, config)
        },

        /**
         * 创建 metric
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        createMetric (context, params, config = {}) {
            const projectId = params.projectId
            delete params.projectId
            // return http.post(`/api/metric/index?invoke=createMetric`).then(response => response.data)
            return http.post(`${DEVOPS_BCS_API_URL}/api/metric/${projectId}/`, params, config)
        },

        /**
         * 编辑 metric
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        editMetric (context, params, config = {}) {
            const { projectId, metricId } = params
            delete params.projectId
            delete params.metricId
            return http.put(`${DEVOPS_BCS_API_URL}/api/metric/${projectId}/${metricId}/`, params, config)
        },

        /**
         * 删除 metric
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        deleteMetric (context, { projectId, metricId }, config = {}) {
            return http.delete(`${DEVOPS_BCS_API_URL}/api/metric/${projectId}/${metricId}/`, {}, config)
        },

        /**
         * 暂停/恢复 metric
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        pauseAndResumeMetric (context, params, config = {}) {
            const opType = params.op_type
            const projectId = params.projectId
            const metricId = params.metricId

            if (opType === 'pause') {
                return http.delete(`${DEVOPS_BCS_API_URL}/api/metric/${projectId}/${metricId}/`, {
                    data: {
                        ns_id_list: params.ns_id_list,
                        op_type: opType
                    }
                }, config)
            }

            return http.post(`${DEVOPS_BCS_API_URL}/api/metric/${projectId}/${metricId}/`, {
                ns_id_list: params.ns_id_list,
                op_type: opType
            }, config)
        },

        /**
         * 查看 metric 实例
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        checkMetricInstance (context, { projectId, metricId }, config = {}) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/metric/${projectId}/instances/${metricId}/`, {}, config)
        },

        /**
         * 拉取service列表
         * list_services
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        listServices (context, { projectId, clusterId }, config = {}) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/metrics/projects/${projectId}/clusters/${clusterId}/services/`, {}, config)
        },

        /**
         * 获取当前项目下集群的ServiceMonitor
         * list_service_monitor
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        listServiceMonitor (context, { projectId, clusterId }, config = {}) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/metrics/projects/${projectId}/clusters/${clusterId}/service_monitors/`, {}, config)
        },

        /**
         * 拉取targets列表
         * list_targets
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        listTargets (context, { projectId, clusterId }, config = {}) {
            return http.get(`${DEVOPS_BCS_API_URL}/api/metrics/projects/${projectId}/clusters/${clusterId}/targets/`, {}, config)
        },

        /**
         * 获取当前service_monitor的targets
         * get_service_monitor_targets
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getServiceMonitorTargets (context, { projectId, clusterId, namespace, name }, config = {}) {
            return http.get(
                `${DEVOPS_BCS_API_URL}/api/metrics/projects/${projectId}/clusters/${clusterId}/`
                    + `service_monitors/${namespace}/${name}/targets/`,
                {},
                config
            )
        },

        /**
         * 获取当前service_monitor
         * get_service_monitor_targets
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getServiceMonitor (context, { projectId, clusterId, namespace, name }, config = {}) {
            return http.get(
                `${DEVOPS_BCS_API_URL}/api/metrics/projects/${projectId}/clusters/${clusterId}/`
                    + `service_monitors/${namespace}/${name}/`,
                {},
                config
            )
        },

        /**
         * 创建service_monitor
         * create_service_monitor
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        createServiceMonitor (context, params, config = {}) {
            const projectId = params.projectId
            delete params.projectId
            delete params.displayName
            return http.post(
                `${DEVOPS_BCS_API_URL}/api/metrics/projects/${projectId}/clusters/${params.cluster_id}/service_monitors/`,
                params,
                config
            )
        },

        /**
         * 删除 metric
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        deleteServiceMonitor (context, { projectId, clusterId, namespace, name }, config = {}) {
            return http.delete(
                `${DEVOPS_BCS_API_URL}/api/metrics/projects/${projectId}/clusters/${clusterId}/`
                    + `service_monitors/${namespace}/${name}/`,
                {},
                config
            )
        },

        /**
         * 批量删除 metric
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        batchDeleteServiceMonitor (context, { projectId, clusterId, data }, config = {}) {
            return http.delete(
                `${DEVOPS_BCS_API_URL}/api/metrics/projects/${projectId}/clusters/${clusterId}/service_monitors/batch/`,
                { data },
                config
            )
        },

        /**
         * 修改 metric
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        updateServiceMonitor (context, params, config = {}) {
            const projectId = params.projectId
            const clusterId = params.cluster_id
            const namespace = params.namespace
            const name = params.name
            delete params.projectId
            delete params.cluster_id
            delete params.namespace
            delete params.name
            return http.put(
                `${DEVOPS_BCS_API_URL}/api/metrics/projects/${projectId}/clusters/${clusterId}/`
                    + `service_monitors/${namespace}/${name}/`,
                params,
                config
            )
        },

        /**
         * 查看是否需要升级版本
         * get_prometheus_update
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        getPrometheusUpdate (context, { projectId, clusterId }, config = {}) {
            return http.get(
                `${DEVOPS_BCS_API_URL}/api/metrics/projects/${projectId}/clusters/${clusterId}/`
                    + `prometheus/update/`,
                {},
                config
            )
        },

        /**
         * 开始升级版本
         * get_prometheus_update
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
        startPrometheusUpdate (context, { projectId, clusterId }, config = {}) {
            return http.put(
                `${DEVOPS_BCS_API_URL}/api/metrics/projects/${projectId}/clusters/${clusterId}/`
                    + `prometheus/update/`,
                {},
                config
            )
        }
    }
}
