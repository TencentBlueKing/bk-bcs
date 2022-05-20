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
    cloudList,
    createCluster,
    cloudVpc,
    cloudRegion,
    vpccidrList,
    deleteCluster,
    retryCluster,
    taskList,
    taskDetail,
    clusterNode,
    addClusterNode,
    deleteClusterNode,
    clusterDetail,
    modifyCluster,
    importCluster,
    kubeConfig,
    nodeAvailable,
    cloudAccounts,
    createCloudAccounts,
    deleteCloudAccounts,
    cloudRegionByAccount,
    cloudClusterList
} from '@/api/base'

export default {
    namespaced: true,
    actions: {
        async fetchCloudList (ctx, params) {
            const data = await cloudList(params).catch(() => [])
            return data
        },
        async createCluster (ctx, params) {
            const data = await createCluster(params).catch(() => false)
            return data
        },
        async fetchCloudVpc (ctx, params) {
            const data = await cloudVpc(params).catch(() => [])
            return data
        },
        async fetchCloudRegion (ctx, params) {
            const data = await cloudRegion(params).catch(() => [])
            return data
        },
        async fetchVpccidrList (ctx, params) {
            const data = await vpccidrList(params).catch(() => [])
            return data
        },
        async deleteCluster (ctx, params) {
            const data = await deleteCluster({
                ...params,
                operator: ctx.rootState.user?.username
            }).then(() => true).catch(() => false)
            return data
        },
        async retryCluster (ctx, params) {
            const data = await retryCluster({
                ...params,
                operator: ctx.rootState.user?.username
            }).catch(() => null)
            return data
        },
        async taskList (ctx, params) {
            const res = await taskList(params, { needRes: true }).catch(() => ({
                data: [],
                latestTask: null
            }))
            return res
        },
        async taskDetail (ctx, params) {
            const data = await taskDetail(params).catch(() => ({}))
            return data
        },
        async clusterNode (ctx, params) {
            const data = await clusterNode(params).catch(() => [])
            return data
        },
        async addClusterNode (ctx, params) {
            const data = await addClusterNode({
                ...params,
                operator: ctx.rootState.user?.username
            }).catch(() => false)
            return data
        },
        async deleteClusterNode (ctx, params) {
            const data = await deleteClusterNode({
                ...params,
                operator: ctx.rootState.user?.username
            }).then(() => true).catch(() => false)
            return data
        },
        async clusterDetail (ctx, params) {
            const data = await clusterDetail(params, { needRes: true }).catch(() => ({}))
            return data
        },
        async modifyCluster (ctx, params) {
            const data = await modifyCluster(params).then(() => true).catch(() => false)
            return data
        },
        async importCluster (ctx, params) {
            const data = await importCluster(params).then(() => true).catch(() => false)
            return data
        },
        // 可用性测试
        async kubeConfig (ctx, params) {
            const data = await kubeConfig(params).then(() => true).catch(() => false)
            return data
        },
        // 节点是否可用
        async nodeAvailable (ctx, params) {
            const data = await nodeAvailable(params).catch(() => ({}))
            return data
        },
        // 云账户信息
        async cloudAccounts (ctx, params) {
            const data = await cloudAccounts(params).catch(() => [])
            return data
        },
        async createCloudAccounts (ctx, params) {
            const result = await createCloudAccounts(params).then(() => true).catch(() => false)
            return result
        },
        async deleteCloudAccounts (ctx, params) {
            const result = await deleteCloudAccounts(params).then(() => true).catch(() => false)
            return result
        },
        async cloudRegionByAccount (ctx, params) {
            const data = await cloudRegionByAccount(params).catch(() => [])
            return data
        },
        async cloudClusterList (ctx, params) {
            const data = await cloudClusterList(params).catch(() => [])
            return data
        }
    }
}
