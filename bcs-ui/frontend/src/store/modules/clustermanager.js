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
  nodeTemplateList,
  createNodeTemplate,
  deleteNodeTemplate,
  updateNodeTemplate,
  nodeTemplateDetail,
  bkSopsList,
  bkSopsParamsList,
  cloudModulesParamsList,
  bkSopsTemplatevalues,
  bkSopsDebug,
  taskRetry,
} from '@/api/base';

export default {
  namespaced: true,
  actions: {
    async fetchCloudList(ctx, params) {
      const data = await cloudList(params).catch(() => []);
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
      const data = await cloudAccounts(params).catch(() => []);
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
  },
};
