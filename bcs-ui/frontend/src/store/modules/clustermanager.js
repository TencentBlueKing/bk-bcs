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

import { cloudList, createCluster, cloudVpc, cloudRegion, vpccidrList, deleteCluster, retryCluster, taskList, taskDetail, clusterNode } from '@/api/base'

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
            const data = await deleteCluster(params).catch(() => false)
            return data
        },
        async retryCluster (ctx, params) {
            const data = await retryCluster(params).catch(() => null)
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
        }
    }
}
