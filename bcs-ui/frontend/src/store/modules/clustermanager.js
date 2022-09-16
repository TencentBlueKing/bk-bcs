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
    cloudClusterList,
    nodeGroup,
    resourceSchema,
    cloudOsImage,
    cloudInstanceTypes,
    cloudSecurityGroups,
    cloudSubnets,
    nodeGroupDetail,
    createNodeGroup,
    updateNodeGroup,
    clusterAutoScaling,
    toggleClusterAutoScalingStatus,
    updateClusterAutoScaling,
    deleteNodeGroup,
    disableNodeGroupAutoScale,
    enableNodeGroupAutoScale,
    clusterAutoScalingLogs,
    nodeGroupNodeList,
    clusterNodeDrain,
    deleteNodeGroupNode,
    nodeCordon,
    nodeUnCordon,
    cloudDetail
} from '@/api/base'

export default {
    namespaced: true,
    actions: {
        async fetchCloudList (ctx, params) {
            const data = await cloudList(params).catch(() => [])
            return data
        },
        async cloudDetail (ctx, params) {
            const data = await cloudDetail(params).catch(() => ({}))
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
            const data = await cloudAccounts(params, { needRes: true }).catch(() => [])
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
        },
        // 节点池列表
        async nodeGroup (ctx, params) {
            const data = await nodeGroup(params).catch(() => [])
            return data
        },
        // 获取云默认配置信息
        async resourceSchema (ctx, params) {
            const data = await resourceSchema(params).catch((res) => ({}))
            return data
        },
        // 云镜像列表
        async cloudOsImage (ctx, params) {
            const data = await cloudOsImage(params).catch(() => [])
            return data
        },
        // Node机型列表
        async cloudInstanceTypes (ctx, params) {
            const data = await cloudInstanceTypes(params).catch(() => [])
            return data
        },
        // 安全组列表
        async cloudSecurityGroups (ctx, params) {
            const data = await cloudSecurityGroups(params).catch(() => [])
            return data
        },
        // VPC子网
        async cloudSubnets (ctx, params) {
            const data = await cloudSubnets(params).catch(() => [])
            return data
        },
        // 获取节点池详情
        async nodeGroupDetail (ctx, params) {
            const data = await nodeGroupDetail(params).catch(() => ({}))
            return data
        },
        // 创建节点池
        async createNodeGroup (ctx, params) {
            const result = await createNodeGroup(params).then(() => true).catch(() => false)
            return result
        },
        // 更新节点池
        async updateNodeGroup (ctx, params) {
            const result = await updateNodeGroup(params).then(() => true).catch(() => false)
            return result
        },
        // 集群扩缩容配置
        async clusterAutoScaling (ctx, params) {
            const data = await clusterAutoScaling(params).catch(() => false)
            return data
        },
        // 开启或关闭扩缩容配置
        async toggleClusterAutoScalingStatus (ctx, params) {
            const result = await toggleClusterAutoScalingStatus(params)
                .then(() => true).catch(() => false)
            return result
        },
        // 更新扩缩容配置
        async updateClusterAutoScaling (ctx, params) {
            const result = await updateClusterAutoScaling(params)
                .then(() => true).catch(() => false)
            return result
        },
        // 删除节点组
        async deleteNodeGroup (ctx, params) {
            const result = await deleteNodeGroup(params)
                .then(() => true).catch(() => false)
            return result
        },
        // 禁用弹性伸缩
        async disableNodeGroupAutoScale (ctx, params) {
            const result = await disableNodeGroupAutoScale(params)
                .then(() => true).catch(() => false)
            return result
        },
        // 启用弹性伸缩
        async enableNodeGroupAutoScale (ctx, params) {
            const result = await enableNodeGroupAutoScale(params)
                .then(() => true).catch(() => false)
            return result
        },
        // 扩缩容记录
        async clusterAutoScalingLogs (ctx, params) {
            const data = await clusterAutoScalingLogs(params).catch(() => [])
            return data
        },
        // 节点池节点列表
        async nodeGroupNodeList (ctx, params) {
            const data = await nodeGroupNodeList(params).catch(() => [])
            return data
        },
        // POD迁移
        async clusterNodeDrain (ctx, params) {
            const result = await clusterNodeDrain(params)
                .then(() => true).catch(() => false)
            return result
        },
        // 删除节点组节点
        async deleteNodeGroupNode (ctx, params) {
            const result = await deleteNodeGroupNode(params)
                .then(() => true).catch(() => false)
            return result
        },
        async nodeCordon (ctx, params) {
            const result = await nodeCordon(params)
                .then(() => true).catch(() => false)
            return result
        },
        async nodeUnCordon (ctx, params) {
            const result = await nodeUnCordon(params)
                .then(() => true).catch(() => false)
            return result
        }
    }
}
