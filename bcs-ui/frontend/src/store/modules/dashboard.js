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
import {
    dashbordList,
    retrieveDetail,
    podMetric,
    listWorkloadPods,
    listStoragePods,
    listContainers,
    retrieveContainerDetail,
    containerMetric,
    fetchContainerEnvInfo,
    resourceDelete,
    resourceCreate,
    resourceUpdate,
    exampleManifests,
    subscribeList,
    namespaceList,
    customResourceList,
    retrieveCustomResourceDetail,
    customResourceCreate,
    customResourceUpdate,
    customResourceDelete,
    reschedulePod,
    logLinks,
    dashbordListWithoutNamespace,
    crdList
} from '@/api/base'

export default {
    namespaced: true,
    state: {
    },
    mutations: {
    },
    actions: {
        // 获取表格数据通用方法
        async getTableData (context, params) {
            const res = await dashbordList(params, { needRes: true }).catch(() => ({
                data: {
                    manifest: {},
                    manifest_ext: {}
                }
            }))
            return res
        },

        // 获取表格数据通用方法（无命名空间）
        async getTableDataWithoutNamespace (context, params) {
            const res = await dashbordListWithoutNamespace(params, { needRes: true }).catch(() => ({
                data: {
                    manifest: {},
                    manifest_ext: {}
                }
            }))
            return res
        },

        // 订阅接口
        async subscribeList (context, params, config = { needRes: true }) {
            if (!context.rootState?.curClusterId) return { events: [], latest_rv: null }
            const res = await subscribeList(params, config).catch((err) => {
                if (err.code === 4005005) { // resourceVersion 重载当前窗口（也可以在每个界面重新调用获取列表详情的接口，目前这样快速处理）
                    location.reload()
                }
                return {
                    data: { events: [], latest_rv: null }
                }
            })
            return res.data
        },
        // 获取命名空间
        async getNamespaceList (context, params, config = {}) {
            const data = await namespaceList(params).catch(() => ({
                manifest: {},
                manifest_ext: {}
            }))
            return data
        },

        /**
         * 获取工作负载详情
         * @param {*} context
         * @param {*} params
         * @param {*} config
         * @returns
         */
        async getResourceDetail (context, params) {
            const res = await retrieveDetail(params, { needRes: true }).catch(() => ({
                data: {
                    manifest: {},
                    manifest_ext: {}
                }
            }))
            return res
        },

        /**
         * 获取pod指标项
         * @param {*} context
         * @param {*} params
         * @param {*} config
         * @returns
         */
        async podMetric (context, params, config = {}) {
            const data = await podMetric(params, config).catch(() => ({
                result: [],
                resultType: ''
            }))
            return data
        },

        /**
         * 容器指标
         * @param {*} context
         * @param {*} params
         * @param {*} config
         * @returns
         */
        async containerMetric (context, params, config = {}) {
            const data = await containerMetric(params, config).catch(() => ({
                result: [],
                resultType: ''
            }))
            return data
        },

        /**
         * 获取工作负载下属的pod
         * @param {*} context
         * @param {*} params
         * @param {*} config
         * @returns
         */
        async listWorkloadPods (context, params, config = {}) {
            const data = await listWorkloadPods(params, config = {}).catch(() => ({
                manifest: {},
                manifest_ext: {}
            }))
            return data
        },

        /**
         * 获取pod下的存储信息
         */
        async listStoragePods (context, params, config = {}) {
            const data = await listStoragePods(params, config = {}).catch(() => ({
                manifest: {},
                manifest_ext: {}
            }))
            return data
        },

        /**
         * 获取指定 pod 下 container 列表
         * @param {*} context
         * @param {*} params
         * @param {*} config
         * @returns
         */
        async listContainers (context, params, config = {}) {
            const data = await listContainers(params, config = {}).catch(() => ([]))
            return data
        },

        /**
         * 获取pod 下Container 详情
         * @param {*} context
         * @param {*} params
         * @param {*} config
         */
        async retrieveContainerDetail (context, params, config = {}) {
            const data = await retrieveContainerDetail(params, config = {}).catch(() => ({}))
            return data
        },

        /**
         * 容器的环境变量
         * @param {*} context
         * @param {*} params
         * @param {*} config
         * @returns
         */
        async fetchContainerEnvInfo (context, params, config = {}) {
            const data = await fetchContainerEnvInfo(params, config = {}).catch(() => ([]))
            return data
        },
        // 资源删除
        async resourceDelete (context, params, config = {}) {
            const data = await resourceDelete(params, config = {}).catch(() => false)
            return data
        },
        // 资源创建
        async resourceCreate (context, params, config = {}) {
            // 需要单独处理错误信息
            const data = await resourceCreate(params, config = {})
            return data
        },
        // 资源更新
        async resourceUpdate (context, params, config = {}) {
            // 需要单独处理错误信息
            const data = await resourceUpdate(params, config = {})
            return data
        },
        // yaml实例
        async exampleManifests (context, params, config = {}) {
            const data = await exampleManifests(params, config = {}).catch(() => ({
                kind: '',
                reference: '',
                items: []
            }))
            return data
        },
        // 获取CRD列表
        async crdList () {
            const res = await crdList({}, { needRes: true }).catch(() => ({
                data: {
                    manifest: {},
                    manifest_ext: {}
                }
            }))
            return res
        },
        // 自定义资源列表
        async customResourceList (context, params) {
            const res = await customResourceList(params, { needRes: true }).catch(() => ({
                data: {
                    manifest: {},
                    manifest_ext: {}
                }
            }))
            return res
        },
        // 自定义资源详情
        async retrieveCustomResourceDetail (context, params) {
            const res = retrieveCustomResourceDetail(params, { needRes: true }).catch(() => ({
                data: {
                    manifest: {},
                    manifest_ext: {}
                }
            }))
            return res
        },
        // 自定义资源创建（需要单独处理错误信息）
        async customResourceCreate (context, params) {
            const data = await customResourceCreate(params)
            return data
        },
        // 自定义资源更新（需要单独处理错误信息）
        async customResourceUpdate (context, params) {
            const data = await customResourceUpdate(params)
            return data
        },
        // 自定义资源删除
        async customResourceDelete (context, params) {
            const data = await customResourceDelete(params).catch(() => false)
            return data
        },
        // 重新调度
        async reschedulePod (context, params) {
            const data = await reschedulePod(params).catch(() => false)
            return data
        },
        // 容器日志链接
        async logLinks (context, params) {
            const data = await logLinks(params).catch(() => ({}))
            return data
        }
    }
}
