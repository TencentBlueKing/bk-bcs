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

import http from '@/api';
import { json2Query } from '@/common/util';
import {
  clusterTools,
  clusterToolsInstalledDetail,
  clusterToolsInstall,
  clusterToolsUpgrade,
  clusterToolsUninstall,
  createLogCollect,
  logCollectList,
  updateLogCollect,
  deleteLogCollect,
  retrieveLogCollect,
} from '@/api/base';

export default {
  namespaced: true,
  state: {
    crdInstanceList: [],
    nameSpaceList: [],
  },
  mutations: {
    /**
         * 更新crdInstanceList
         * @param {Object} state store state
         * @param {Object} data data
         */
    updateCrdInstanceList(state, data) {
      state.crdInstanceList = data;
    },

    /**
         * 更新namespace 列表
         *
         * @param {Object} state store state
         * @param {Object} data data
         */
    updateNameSpaceList(state, data) {
      state.nameSpaceList.splice(0, state.nameSpaceList.length, ...data);
    },
  },
  actions: {
    // 获取组件库
    async clusterTools(ctx, params) {
      const data = await clusterTools(params).catch(() => ([]));
      return data;
    },

    // 组件启用
    async clusterToolsInstall(ctx, params) {
      const result = await clusterToolsInstall(params).then(() => true)
        .catch(() => false);
      return result;
    },

    // 组件更新
    async clusterToolsUpgrade(ctx, params) {
      const data = await clusterToolsUpgrade(params).then(() => true)
        .catch(() => false);
      return data;
    },

    // 组件库详情
    async clusterToolsInstalledDetail(ctx, params) {
      const data = await clusterToolsInstalledDetail(params).catch(() => ({}));
      return data;
    },

    // 组件卸载
    async clusterToolsUninstall(ctx, params) {
      const result = await clusterToolsUninstall(params).then(() => true)
        .catch(() => false);
      return result;
    },

    /**
         * 查询crd列表 (老，仅DB授权使用)
         *
         * @param {Object} context store 上下文对象
         * @param {Object} projectId, clusterId, crdKind
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    getCrdInstanceList(context, { projectId, clusterId, crdKind, params = {} }, config = {}) {
      context.commit('updateCrdInstanceList', []);
      const url = `${DEVOPS_BCS_API_URL}/api/bcs_crd/projects/${projectId}/clusters/${clusterId}/crds/${crdKind}/custom_objects/?${json2Query(params)}`;
      return http.get(url, {}, config).then((res) => {
        context.commit('updateCrdInstanceList', res.data);
        return res;
      });
    },

    /**
         * 查询crd列表 (新)
         *
         * @param {Object} context store 上下文对象
         * @param {Object} projectId, clusterId, crdKind
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    getBcsCrdsList(context, { projectId, params = {} }, config = {}) {
      context.commit('updateCrdInstanceList', []);
      const url = `${DEVOPS_BCS_API_URL}/api/bcs_crd/projects/${projectId}/crds/?${json2Query(params)}`;
      return http.get(url, {}, config).then((res) => {
        context.commit('updateCrdInstanceList', res.data);
        return res;
      });
    },

    /**
         * 查询单个crd (仅用于DB授权)
         *
         * @param {Object} context store 上下文对象
         * @param {Object} projectId, clusterId, crdId
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    getCrdInstanceDetail(context, { projectId, clusterId, crdKind, crdId }, config = {}) {
      const url = `${DEVOPS_BCS_API_URL}/api/bcs_crd/projects/${projectId}/clusters/${clusterId}/crds/${crdKind}/custom_objects/-/detail/?id=${crdId}`;
      return http.get(url, {}, config);
    },

    /**
         * 查询单个crd (日志采集)
         *
         * @param {Object} context store 上下文对象
         * @param {Object} projectId, crdKind, crdId
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    getLogCrdInstanceDetail(context, { projectId, crdId, crdKind, clusterId }, config = {}) {
      const url = `${DEVOPS_BCS_API_URL}/api/bcs_crd/projects/${projectId}/clusters/${clusterId}/crds/${crdKind}/custom_objects/-/detail/?id=${crdId}`;
      return http.get(url, {}, config);
    },

    /**
         * 获取版本列表
         */
    getChartVersionsList(context, { projectId, chartName }) {
      return http.get(`${DEVOPS_BCS_API_URL}/api/helm_chart/projects/${projectId}/charts/${chartName}/versions/?is_public_repo=true`);
    },

    /**
         * 获取当前集群下的namespace 列表
         *
         * @param {Object} context store 上下文对象
         * @param {Object} projectId, clusterId
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    getNameSpaceListByCluster(context, { projectId, clusterId }, config = {}) {
      return http.get(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/namespace/?cluster_id=${clusterId}`, {}, config).then((res) => {
        context.commit('updateNameSpaceList', res.data);
        return res;
      });
    },

    /**
         * 添加crdInstance
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含：projectId, data
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    addCrdInstance(context, { projectId, clusterId, crdKind, data }, config = {}) {
      return http.post(`${DEVOPS_BCS_API_URL}/api/bcs_crd/projects/${projectId}/clusters/${clusterId}/crds/${crdKind}/custom_objects/`, data, config);
    },

    /**
         * 更新crdInstance
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含：projectId, data
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    updateCrdInstance(context, { projectId, clusterId, crdKind, data }, config = {}) {
      return http.put(`${DEVOPS_BCS_API_URL}/api/bcs_crd/projects/${projectId}/clusters/${clusterId}/crds/${crdKind}/custom_objects/`, data, config);
    },

    /**
         * 删除单个crd
         *
         * @param {Object} context store 上下文对象
         * @param {Object} projectId, clusterId, crdId
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    deleteCrdInstance(context, { projectId, clusterId, crdKind, crdId }, config = {}) {
      const url = `${DEVOPS_BCS_API_URL}/api/bcs_crd/projects/${projectId}/clusters/${clusterId}/crds/${crdKind}/custom_objects/`;
      return http.delete(url, { data: { id: crdId } }, config);
    },
    async createLogCollect(ctx, params) {
      const result = await createLogCollect(params).then(() => true)
        .catch(() => false);
      return result;
    },
    async logCollectList(ctx, params) {
      const data = await logCollectList(params).catch(() => []);
      return data;
    },
    async updateLogCollect(ctx, params) {
      const result = await updateLogCollect(params).then(() => true)
        .catch(() => false);
      return result;
    },
    async deleteLogCollect(ctx, params) {
      const result = await deleteLogCollect(params).then(() => true)
        .catch(() => false);
      return result;
    },
    async retrieveLogCollect(ctx, params) {
      const data = await retrieveLogCollect(params).catch(() => ({}));
      return data;
    },
  },
};
