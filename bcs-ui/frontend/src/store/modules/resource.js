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
/* eslint-disable @typescript-eslint/no-unused-vars */
import http from '@/api';
import { json2Query } from '@/common/util';

export default {
  namespaced: true,
  state: {
    configmapList: [],
    secretList: [],
    ingressList: [],
  },
  mutations: {
    /**
         * 更新 store.resource 中的 configmpaList
         *
         * @param {Object} state store
         * @param {Array} data 列表
         */
    updateConfigmapList(state, data) {
      state.configmapList.splice(0, state.configmapList.length, ...data);
    },

    /**
         * 更新 store.resource 中的 secretList
         *
         * @param {Object} state store
         * @param {Array} data 列表
         */
    updateSecretList(state, data) {
      state.secretList.splice(0, state.secretList.length, ...data);
    },

    /**
         * 更新 store.resource 中的 ingressList
         *
         * @param {Object} state store
         * @param {Array} data 列表
         */
    updateIngressList(state, data) {
      state.ingressList.splice(0, state.ingressList.length, ...data);
    },
  },
  actions: {
    /**
         * 根据项目 id 获取项目下的configmap
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    getConfigmapList(context, { projectId, params }, config = {}) {
      return http.get(`${DEVOPS_BCS_API_URL}/api/resource/${projectId}/configmaps/?${json2Query(params)}`, {}, config).then((response) => {
        const res = response;
        res.data.forEach((item) => {
          if (item.data.metadata.labels) {
            const labels = Object.entries(item.data.metadata.labels);
            item.labels = labels;
          } else {
            item.labels = [];
          }
        });
        context.commit('updateConfigmapList', res.data);
        return res;
      });
    },

    /**
         * 根据项目 id 获取项目下的secret
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    getSecretList(context, { projectId, params }, config = {}) {
      return http.get(`${DEVOPS_BCS_API_URL}/api/resource/${projectId}/secrets/?${json2Query(params)}`).then((response) => {
        const res = response;
        res.data.forEach((item) => {
          if (item.data.metadata.labels) {
            const labels = Object.entries(item.data.metadata.labels);
            item.labels = labels;
          } else {
            item.labels = [];
          }
        });
        context.commit('updateSecretList', res.data);
        return res;
      });
    },

    // 获取Ingress 列表
    getIngressList(context, { projectId, params }, config = {}) {
      // return http.get('/app/resource?invoke=getIngressList', {}, config).then(response => {
      return http.get(`${DEVOPS_BCS_API_URL}/api/resource/${projectId}/ingresses/?${json2Query(params)}`, {}, config).then((response) => {
        const res = response;
        res.data.forEach((item) => {
          item.labels = [];
          item.rules = [];
          item.tls = [];
          if (item.data.metadata.labels) {
            const labels = Object.entries(item.data.metadata.labels);
            item.labels = labels;
          }

          if (item.data.spec.rules) {
            const { rules } = item.data.spec;
            rules.forEach((rule) => {
              // eslint-disable-next-line @typescript-eslint/prefer-optional-chain
              if (rule.http && rule.http.paths) {
                const https = rule.http.paths;
                https.forEach((http) => {
                  item.rules.push({
                    host: rule.host,
                    path: http.path,
                    serviceName: http.backend.serviceName,
                    servicePort: http.backend.servicePort,
                  });
                });
              }
            });
          }

          if (item.data.spec.tls) {
            const { tls } = item.data.spec;
            tls.forEach((computer) => {
              const tmp = {
                host: '',
                secretName: '',
              };
              if (computer.hosts) {
                tmp.host = computer.hosts.join(';');
              }
              if (computer.secretName) {
                tmp.secretName = computer.secretName;
              }
              item.tls.push(tmp);
            });
          }
        });
        context.commit('updateIngressList', res.data);
        return res;
      });
    },

    // 点击更新时查询单个Secret
    updateSelectSecret(context, { projectId, namespace, name, clusterId }, config = {}) {
      return http.get(`${DEVOPS_BCS_API_URL}/api/resource/${projectId}/secrets/?cluster_id=${clusterId}&namespace=${namespace}&name=${name}&decode=1`);
    },

    // 更新单个Secret
    updateSingleSecret(context, { projectId, clusterId, namespace, name, data }, config = {}) {
      return http.post(`${DEVOPS_BCS_API_URL}/api/resource/${projectId}/secrets/clusters/${clusterId}/namespaces/${namespace}/endpoints/${name}/`, data, config);
    },

    // 删除单个Secret
    deleteSecret(context, { projectId, clusterId, namespace, name }, config = {}) {
      return http.delete(`${DEVOPS_BCS_API_URL}/api/resource/${projectId}/secrets/clusters/${clusterId}/namespaces/${namespace}/endpoints/${name}/`, {}, config);
    },

    // 删除Secrets
    deleteSecrets(context, { projectId, data }, config = {}) {
      const params = {
        data,
      };
      return http.post(`${DEVOPS_BCS_API_URL}/api/resource/${projectId}/secrets/batch/`, params, config);
    },

    // 删除单个Ingress
    deleteIngress(context, { projectId, clusterId, namespace, name }, config = {}) {
      return http.delete(`${DEVOPS_BCS_API_URL}/api/resource/${projectId}/ingresses/clusters/${clusterId}/namespaces/${namespace}/endpoints/${name}/`, {}, config);
    },

    // 删除Ingresss
    deleteIngresses(context, { projectId, data }, config = {}) {
      const params = {
        data,
      };
      return http.post(`${DEVOPS_BCS_API_URL}/api/resource/${projectId}/ingresses/batch/`, params, config);
    },

    /**
         * 更新单个configmap
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    updateSingleConfigmap(context, { projectId, clusterId, namespace, name, data }, config = {}) {
      return http.post(`${DEVOPS_BCS_API_URL}/api/resource/${projectId}/configmaps/clusters/${clusterId}/namespaces/${namespace}/endpoints/${name}/`, data);
    },

    /**
         * 删除单个Configmap
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    deleteConfigmap(context, { projectId, clusterId, namespace, name }, config = {}) {
      return http.delete(`${DEVOPS_BCS_API_URL}/api/resource/${projectId}/configmaps/clusters/${clusterId}/namespaces/${namespace}/endpoints/${name}/`);
    },

    /**
         * 删除Configmaps
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    deleteConfigmaps(context, { projectId, data }, config = {}) {
      const params = {
        data,
      };
      return http.post(`${DEVOPS_BCS_API_URL}/api/resource/${projectId}/configmaps/batch/`, params);
    },

    /**
         * 点击更新时查询单个configmap
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    updateSelectConfigmap(context, { projectId, namespace, name, clusterId }, config = {}) {
      return http.get(`${DEVOPS_BCS_API_URL}/api/resource/${projectId}/configmaps/?cluster_id=${clusterId}&namespace=${namespace}&name=${name}&decode=1`);
    },

    /**
         * 保存Ingress
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含：projectId, clusterId, namespace, ingressId, data
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    saveIngressDetail(context, { projectId, clusterId, namespace, ingressId, data }, config = {}) {
      return http.put(`${DEVOPS_BCS_API_URL}/api/projects/${projectId}/clusters/${clusterId}/namespaces/${namespace}/ingresses/${ingressId}/`, data);
    },
  },
};
