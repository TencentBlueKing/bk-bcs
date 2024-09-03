/*
* Tencent is pleased to support the open source community by making
* 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition) available.
*
* Copyright (C) 2021 THL A29 Limited, a Tencent company.  All rights reserved.
*
* 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition) is licensed under the MIT License.
*
* License for 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition):
*
* ---------------------------------------------------
* Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated
* documentation files (the "Software"), to deal in the Software without restriction, including without limitation
* the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and
* to permit persons to whom the Software is furnished to do so, subject to the following conditions:
*
* The above copyright notice and this permission notice shall be included in all copies or substantial portions of
* the Software.
*
* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO
* THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF
* CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
* IN THE SOFTWARE.
*/

import {
  addClusterNode,
  bkSopsDebug,
  bkSopsList,
  bkSopsParamsList,
  bkSopsTemplatevalues,
  cloudAccounts,
  cloudClusterList,
  cloudInstanceTypes,
  cloudList,
  cloudModulesParamsList,
  cloudOsImage,
  cloudNoderoles,
  cloudRegion,
  cloudRegionByAccount,
  cloudResourceGroupByAccount,
  cloudSecurityGroups,
  cloudSubnets,
  cloudVpc,
  clusterAutoScaling,
  clusterAutoScalingLogs,
  clusterDetail,
  clusterNode,
  clusterNodeDrain,
  createCloudAccounts,
  createCluster,
  createNodeGroup,
  createNodeTemplate,
  deleteCloudAccounts,
  deleteCluster,
  deleteClusterNode,
  deleteNodeGroup,
  deleteNodeGroupNode,
  deleteNodeTemplate,
  disableNodeGroupAutoScale,
  enableNodeGroupAutoScale,
  importCluster,
  kubeConfig,
  modifyCluster,
  nodeAvailable,
  nodeCordon,
  nodeGroup,
  nodeGroupDetail,
  nodeGroupNodeList,
  nodeTemplateDetail,
  nodeTemplateList,
  nodeUnCordon,
  resourceSchema,
  retryCluster,
  taskDetail,
  taskList,
  taskRetry,
  toggleClusterAutoScalingStatus,
  updateClusterAutoScaling,
  updateNodeGroup,
  updateNodeTemplate,
  vpccidrList,
} from '@/api/base';
import { cloudDetail, cloudKeyPairs, clusterAutoScalingLogsV2 } from '@/api/modules/cluster-manager';

export default {
  namespaced: true,
  actions: {
    async fetchCloudList(ctx, params) {
      const data = await cloudList(params).catch(() => []);
      return data;
    },
    async cloudDetail(ctx, params) {
      const data = await cloudDetail(params).catch(() => ({}));
      return data;
    },
    async createCluster(ctx, params) {
      const data = await createCluster(params).catch(() => false);
      return data;
    },
    async fetchCloudVpc(ctx, params) {
      const data = await cloudVpc(params).catch(() => []);
      return data;
    },
    async fetchCloudRegion(ctx, params) {
      const data = await cloudRegion(params).catch(() => []);
      return data;
    },
    async fetchVpccidrList(ctx, params) {
      const data = await vpccidrList(params).catch(() => []);
      return data;
    },
    async deleteCluster(ctx, params) {
      const data = await deleteCluster({
        ...params,
        operator: ctx.rootState.user?.username,
      }).then(() => true)
        .catch(() => false);
      return data;
    },
    async retryCluster(ctx, params) {
      const data = await retryCluster({
        ...params,
        operator: ctx.rootState.user?.username,
      }).catch(() => null);
      return data;
    },
    async taskList(ctx, params) {
      const res = await taskList(params, { needRes: true }).catch(() => ({
        data: [],
        latestTask: null,
      }));
      return res;
    },
    async taskDetail(ctx, params) {
      const data = await taskDetail(params).catch(() => ({}));
      return data;
    },
    async clusterNode(ctx, params) {
      const data = await clusterNode(params).catch(() => []);
      return data;
    },
    async addClusterNode(ctx, params) {
      const data = await addClusterNode({
        ...params,
        operator: ctx.rootState.user?.username,
      }).catch(() => false);
      return data;
    },
    async deleteClusterNode(ctx, params) {
      const data = await deleteClusterNode({
        ...params,
        operator: ctx.rootState.user?.username,
      }).then(() => true)
        .catch(() => false);
      return data;
    },
    async clusterDetail(ctx, params) {
      const data = await clusterDetail(params, { needRes: true }).catch(() => ({}));
      return data;
    },
    async modifyCluster(ctx, params) {
      const data = await modifyCluster(params).then(() => true)
        .catch(() => false);
      return data;
    },
    async importCluster(ctx, params) {
      const data = await importCluster(params).then(() => true)
        .catch(() => false);
      return data;
    },
    // 可用性测试
    async kubeConfig(ctx, params) {
      const data = await kubeConfig(params).then(() => true)
        .catch(() => false);
      return data;
    },
    // 节点是否可用
    async nodeAvailable(ctx, params) {
      const data = await nodeAvailable(params).catch(() => ({}));
      return data;
    },
    // 云账户信息
    async cloudAccounts(ctx, params) {
      const data = await cloudAccounts(params, { needRes: true }).catch(() => []);
      return data;
    },
    async createCloudAccounts(ctx, params) {
      const result = await createCloudAccounts(params).then(() => true)
        .catch(() => false);
      return result;
    },
    async deleteCloudAccounts(ctx, params) {
      const result = await deleteCloudAccounts(params).then(() => true)
        .catch(() => false);
      return result;
    },
    async cloudResourceGroupByAccount(ctx, params) {
      const data = await cloudResourceGroupByAccount(params).catch(() => []);
      return data;
    },
    async cloudRegionByAccount(ctx, params) {
      const data = await cloudRegionByAccount(params).catch(() => []);
      return data;
    },
    async cloudClusterList(ctx, params) {
      const data = await cloudClusterList(params).catch(() => []);
      return data;
    },
    async nodeTemplateList(ctx, params) {
      const data = await nodeTemplateList(params).catch(() => []);
      return data;
    },
    async createNodeTemplate(ctx, params) {
      const result = await createNodeTemplate(params).then(() => true)
        .catch(() => false);
      return result;
    },
    async deleteNodeTemplate(ctx, params) {
      const result = await deleteNodeTemplate(params).then(() => true)
        .catch(() => false);
      return result;
    },
    async updateNodeTemplate(ctx, params) {
      const result = await updateNodeTemplate(params).then(() => true)
        .catch(() => false);
      return result;
    },
    async nodeTemplateDetail(ctx, params) {
      const data = await nodeTemplateDetail(params).catch(() => ({}));
      return data;
    },
    async bkSopsList(ctx, params) {
      const data = await bkSopsList(params).catch(() => []);
      return data;
    },
    async bkSopsParamsList(ctx, params) {
      const data = await bkSopsParamsList(params).catch(() => ({ templateUrl: '', values: [] }));
      return data;
    },
    async cloudModulesParamsList(ctx, params) {
      const data = await cloudModulesParamsList(params).catch(() => []);
      return data;
    },
    async bkSopsTemplatevalues(ctx, params) {
      const data = await bkSopsTemplatevalues(params).catch(() => []);
      return data;
    },
    async bkSopsDebug(ctx, params) {
      const data = await bkSopsDebug(params).catch(() => ({}));
      return data;
    },
    async taskRetry(ctx, params) {
      const result = await taskRetry(params).catch(() => false);
      return result;
    },
    // 节点池列表
    async nodeGroup(ctx, params) {
      const data = await nodeGroup(params).catch(() => []);
      return data;
    },
    // 获取云默认配置信息
    async resourceSchema(ctx, params) {
      const data = await resourceSchema(params).catch(() => ({}));
      return data;
    },
    // 云镜像列表
    async cloudOsImage(ctx, params) {
      const data = await cloudOsImage(params).catch(() => []);
      return data;
    },
    // IAM角色
    async cloudNoderoles(ctx, params) {
      const data = await cloudNoderoles(params).catch(() => []);
      return data;
    },
    // Node机型列表
    async cloudInstanceTypes(ctx, params) {
      const data = await cloudInstanceTypes(params).catch(() => []);
      return data;
    },
    // 安全组列表
    async cloudSecurityGroups(ctx, params) {
      const data = await cloudSecurityGroups(params).catch(() => []);
      return data;
    },
    // VPC子网
    async cloudSubnets(ctx, params) {
      const data = await cloudSubnets(params).catch(() => []);
      return data;
    },
    // 获取节点池详情
    async nodeGroupDetail(ctx, params) {
      const data = await nodeGroupDetail(params).catch(() => ({}));
      return data;
    },
    // 创建节点池
    async createNodeGroup(ctx, params) {
      const result = await createNodeGroup(params).then(() => true)
        .catch(() => false);
      return result;
    },
    // 更新节点池
    async updateNodeGroup(ctx, params) {
      const result = await updateNodeGroup(params).then(() => true)
        .catch(() => false);
      return result;
    },
    // 集群扩缩容配置
    async clusterAutoScaling(ctx, params) {
      const data = await clusterAutoScaling(params).catch(() => ({}));
      return data;
    },
    // 开启或关闭扩缩容配置
    async toggleClusterAutoScalingStatus(ctx, params) {
      const result = await toggleClusterAutoScalingStatus(params)
        .then(() => true)
        .catch(() => false);
      return result;
    },
    // 更新扩缩容配置
    async updateClusterAutoScaling(ctx, params) {
      const result = await updateClusterAutoScaling(params)
        .then(() => true)
        .catch(() => false);
      return result;
    },
    // 删除节点组
    async deleteNodeGroup(ctx, params) {
      const result = await deleteNodeGroup(params)
        .then(() => true)
        .catch(() => false);
      return result;
    },
    // 禁用弹性伸缩
    async disableNodeGroupAutoScale(ctx, params) {
      const result = await disableNodeGroupAutoScale(params)
        .then(() => true)
        .catch(() => false);
      return result;
    },
    // 启用弹性伸缩
    async enableNodeGroupAutoScale(ctx, params) {
      const result = await enableNodeGroupAutoScale(params)
        .then(() => true)
        .catch(() => false);
      return result;
    },
    // 扩缩容记录
    async clusterAutoScalingLogs(ctx, params) {
      const data = await clusterAutoScalingLogs(params).catch(() => []);
      return data;
    },
    // 扩缩容记录V2(集群维度)
    async clusterAutoScalingLogsV2(ctx, params = {}) {
      const data = await clusterAutoScalingLogsV2({
        ...params,
        v2: true,
      }).catch(() => []);
      return data;
    },
    // 节点池节点列表
    async nodeGroupNodeList(ctx, params) {
      const data = await nodeGroupNodeList(params).catch(() => []);
      return data;
    },
    // POD迁移
    async clusterNodeDrain(ctx, params) {
      const result = await clusterNodeDrain(params)
        .then(() => true)
        .catch(() => false);
      return result;
    },
    // 删除节点组节点
    async deleteNodeGroupNode(ctx, params) {
      const result = await deleteNodeGroupNode(params)
        .then(() => true)
        .catch(() => false);
      return result;
    },
    async nodeCordon(ctx, params) {
      const result = await nodeCordon(params)
        .then(() => true)
        .catch(() => false);
      return result;
    },
    async nodeUnCordon(ctx, params) {
      const result = await nodeUnCordon(params)
        .then(() => true)
        .catch(() => false);
      return result;
    },
    async cloudKeyPairs(ctx, params) {
      const data = await cloudKeyPairs(params).catch(() => []);
      return data;
    },
  },
};
