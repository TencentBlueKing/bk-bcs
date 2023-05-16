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
  dashbordList,
  retrieveDetail,
  retrieveCustomObjectDetail,
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
  customResourceList,
  retrieveCustomResourceDetail,
  customResourceCreate,
  customResourceUpdate,
  customResourceDelete,
  reschedulePod,
  logLinks,
  dashbordListWithoutNamespace,
  crdList,
  formSchema,
  renderManifestPreview,
  enlargeCapacityChange,
  crdEnlargeCapacityChange,
  batchReschedulePod,
  batchRescheduleCrdPod,
} from '@/api/base';

import { pvcMountInfo, getNetworksEndpointsFlag, getReplicasets } from '@/api/modules/cluster-resource';

export default {
  namespaced: true,
  state: {
  },
  mutations: {
  },
  actions: {
    // 获取表格数据通用方法
    async getTableData(context, params) {
      const res = await dashbordList(params, { needRes: true }).catch(() => ({
        data: {
          manifest: {},
          manifestExt: {},
        },
      }));
      return res;
    },

    // 获取表格数据通用方法（无命名空间）
    async getTableDataWithoutNamespace(context, params) {
      const res = await dashbordListWithoutNamespace(params, { needRes: true }).catch(() => ({
        data: {
          manifest: {},
          manifestExt: {},
        },
      }));
      return res;
    },

    // 订阅接口
    // async subscribeList (context, params, config = { needRes: true }) {
    //     if (!context.rootState?.curClusterId) return { events: [], latest_rv: null }
    //     const res = await subscribeList(params, config).catch((err) => {
    //         if (err.code === 4005005) { // resourceVersion 重载当前窗口（也可以在每个界面重新调用获取列表详情的接口，目前这样快速处理）
    //             location.reload()
    //         }
    //         return {
    //             data: { events: [], latest_rv: null }
    //         }
    //     })
    //     return res.data
    // },

    /**
         * 获取工作负载详情
         * @param {*} context
         * @param {*} params
         * @param {*} config
         * @returns
         */
    async getResourceDetail(context, params) {
      const res = await retrieveDetail(params, { needRes: true }).catch(() => ({
        data: {
          manifest: {},
          manifestExt: {},
        },
      }));
      return res;
    },

    async getCustomObjectResourceDetail(context, params) {
      const res = await retrieveCustomObjectDetail(params, { needRes: true }).catch(() => ({
        data: {
          manifest: {},
          manifestExt: {},
        },
      }));
      return res;
    },

    /**
         * 获取pod指标项
         * @param {*} context
         * @param {*} params
         * @param {*} config
         * @returns
         */
    async podMetric(context, params, config = {}) {
      const data = await podMetric(params, config).catch(() => ({
        result: [],
        resultType: '',
      }));
      return data;
    },

    /**
         * 容器指标
         * @param {*} context
         * @param {*} params
         * @param {*} config
         * @returns
         */
    async containerMetric(context, params, config = {}) {
      const data = await containerMetric(params, config).catch(() => ({
        result: [],
        resultType: '',
      }));
      return data;
    },

    /**
         * 获取工作负载下属的pod
         * @param {*} context
         * @param {*} params
         * @param {*} config
         * @returns
         */
    async listWorkloadPods(context, params, config = {}) {
      const data = await listWorkloadPods(params, config).catch(() => ({
        manifest: {},
        manifestExt: {},
      }));
      return data;
    },

    /**
         * 获取pod下的存储信息
         */
    async listStoragePods(context, params, config = {}) {
      const data = await listStoragePods(params, config).catch(() => ({
        manifest: {},
        manifestExt: {},
      }));
      return data;
    },

    /**
         * 获取指定 pod 下 container 列表
         * @param {*} context
         * @param {*} params
         * @param {*} config
         * @returns
         */
    async listContainers(context, params, config = {}) {
      const data = await listContainers(params, config).catch(() => ([]));
      return data;
    },

    /**
         * 获取pod 下Container 详情
         * @param {*} context
         * @param {*} params
         * @param {*} config
         */
    async retrieveContainerDetail(context, params, config = {}) {
      const data = await retrieveContainerDetail(params, config).catch(() => ({}));
      return data;
    },

    /**
         * 容器的环境变量
         * @param {*} context
         * @param {*} params
         * @param {*} config
         * @returns
         */
    async fetchContainerEnvInfo(context, params, config = {}) {
      const data = await fetchContainerEnvInfo(params, config).catch(() => ([]));
      return data;
    },
    // 资源删除
    async resourceDelete(context, params, config = {}) {
      const data = await resourceDelete(params, config).then(() => true)
        .catch(() => false);
      return data;
    },
    // 资源创建
    async resourceCreate(context, params, config = {}) {
      // 需要单独处理错误信息
      const data = await resourceCreate(params, config);
      return data;
    },
    // 资源更新
    async resourceUpdate(context, params, config = {}) {
      // 需要单独处理错误信息
      const data = await resourceUpdate(params, config);
      return data;
    },
    // yaml实例
    async exampleManifests(context, params, config = {}) {
      const data = await exampleManifests(params, config).catch(() => ({
        kind: '',
        reference: '',
        items: [],
      }));
      return data;
    },
    // 获取CRD列表
    async crdList(context, params) {
      const res = await crdList(params, { needRes: true }).catch(() => ({
        data: {
          manifest: {},
          manifestExt: {},
        },
      }));
      return res;
    },
    // 自定义资源列表
    async customResourceList(context, params) {
      const res = await customResourceList(params, { needRes: true }).catch(() => ({
        data: {
          manifest: {},
          manifestExt: {},
        },
      }));
      return res;
    },
    // 自定义资源详情
    async retrieveCustomResourceDetail(context, params) {
      const res = retrieveCustomResourceDetail(params, { needRes: true }).catch(() => ({
        data: {
          manifest: {},
          manifestExt: {},
        },
      }));
      return res;
    },
    // 自定义资源创建（需要单独处理错误信息）
    async customResourceCreate(context, params) {
      const data = await customResourceCreate(params);
      return data;
    },
    // 自定义资源更新（需要单独处理错误信息）
    async customResourceUpdate(context, params) {
      const data = await customResourceUpdate(params);
      return data;
    },
    // 自定义资源删除
    async customResourceDelete(context, params) {
      const data = await customResourceDelete(params).then(() => true)
        .catch(() => false);
      return data;
    },
    // 重新调度
    async reschedulePod(context, params) {
      const data = await reschedulePod(params).then(() => true)
        .catch(() => false);
      return data;
    },
    // 容器日志链接
    async logLinks(context, params) {
      const data = await logLinks(params).catch(() => ({}));
      return data;
    },
    // 获取表单化配置
    async getFormSchema(context, params) {
      const data = await formSchema(params).catch(() => ({}));
      return data;
    },
    // 表单数据转manifest
    async renderManifestPreview(context, params) {
      const data = await renderManifestPreview(params).catch(() => ({}));
      return data;
    },
    // 扩缩容
    async enlargeCapacityChange(context, params) {
      const data = await enlargeCapacityChange(params).catch(() => false);
      return data;
    },
    // pod批量重新调度
    async batchReschedulePod(context, params) {
      const data = await batchReschedulePod(params).then(() => true);
      return data;
    },
    // crd扩缩容
    async crdEnlargeCapacityChange(context, params) {
      const data = await crdEnlargeCapacityChange(params).catch(() => false);
      return data;
    },
    // crd重新调度
    async batchRescheduleCrdPod(context, params) {
      const data = await batchRescheduleCrdPod(params).then(() => true);
      return data;
    },
    async getPvcMountInfo(context, params) {
      const data = await pvcMountInfo(params).catch(() => ({ podNames: [] }));
      return data;
    },
    async getNetworksEndpointsFlag(context, params) {
      const data = await getNetworksEndpointsFlag(params).catch(() => false);
      return data.epReady;
    },
    async getReplicasets(ctx, params) {
      const data = await getReplicasets(params).catch(() => ({}));
      return data;
    },
  },
};
