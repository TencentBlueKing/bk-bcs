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

export default {
  namespaced: true,
  state: {
    HPAList: [],
  },
  mutations: {
    /**
         * 更新HPA list
         * @param {Object} state store state
         * @param {Object} data data
         */
    updateHPAList(state, data) {
      state.HPAList = data;
    },
  },
  actions: {
    /**
         * 获取HPA list
         *
         * @param {Object} context store 上下文对象
         * @param {String} projectId, projectId
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    getHPAList(context, { projectId, clusterId }) {
      // 清空上次数据
      context.commit('updateHPAList', []);
      const url = `${DEVOPS_BCS_API_URL}/api/hpa/projects/${projectId}/?cluster_id=${clusterId}`;
      return http.get(url, {}, { cancelWhenRouteChange: true }).then((res) => {
        const list = res.data || [];
        list.forEach((item) => {
          const conditions = item.conditions || [];
          const conditionsLen = conditions.length;
          for (let i = 0; i < conditionsLen; i++) {
            if (conditions[i].status.toLowerCase() === 'false') {
              item.needShowConditions = true;
              break;
            }
          }
        });
        context.commit('updateHPAList', list);
        return res;
      });
    },

    /**
         * 批量删除HPA
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数对象，包括projectId，hpa列表
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    batchDeleteHPA(context, { projectId, params }, config = {}) {
      const url = `${DEVOPS_BCS_API_URL}/api/hpa/projects/${projectId}/`;
      return http.delete(url, { data: params }, config);
    },

    /**
         * 批量HPA
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数对象，包括projectId，clusterId, namespace, name
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    deleteHPA(context, { projectId, clusterId, namespace, name }, config = {}) {
      const url = `${DEVOPS_BCS_API_URL}/api/hpa/projects/${projectId}/clusters/${clusterId}/namespaces/${namespace}/${name}/`;
      return http.delete(url, {}, config);
    },
  },
};
