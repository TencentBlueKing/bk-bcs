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
import useVariable from '@/views/deploy-manage/variable/use-variable';

export default {
  namespaced: true,
  state: {
    varList: [],
  },
  mutations: {
    updateVarList(state, data) {
      state.varList.splice(0, state.varList.length, ...data);
    },
  },
  actions: {
    getVarList(context, projectId, config = {}) {
      return http.get(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/variables/?limit=10000&offset=0`).then((response) => {
        if (response.code === 0) {
          context.commit('updateVarList', response.data.results);
        }
        return response.data;
      });
    },
    getVarListByPage(context, { projectId, offset = 0, limit = 100000, scope = '', keyword = '' }, config = {}) {
      return http.get(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/variables/?limit=${limit}&offset=${offset}&search_key=${keyword}&scope=${scope}`).then(response => response.data);
    },
    getNamespaceBatchVarList(context, { projectId, variableId }, config = {}) {
      return http.get(
        `${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/variable/batch/namespace/${variableId}/`,
        {},
        config,
      );
    },
    getClusterBatchVarList(context, { projectId, variableId }, config = {}) {
      return http.get(
        `${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/variable/batch/cluster/${variableId}/`,
        {},
        config,
      );
    },

    async getBaseVarList(context, projectId, config = {}) {
      const { getVariableDefinitions } = useVariable();
      const data = await getVariableDefinitions({ all: true });
      context.commit('updateVarList', data.results);
      return data;
    },

    getQuoteDetail(context, { projectId, varId }, config = {}) {
      return http.get(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/variable/quotes/${varId}/`, {}, config);
    },
    addVar(context, { projectId, data }, config = {}) {
      return http.post(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/variables/`, data, config);
    },
    updateVar(context, { projectId, varId, data }, config = {}) {
      return http.post(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/variable/${varId}/`, data, config);
    },
    updateNamespaceBatchVar(context, { projectId, varId, data }, config = {}) {
      return http.post(
        `${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/variable/batch/namespace/${varId}/`,
        data,
        config,
      );
    },
    updateClusterBatchVar(context, { projectId, varId, data }, config = {}) {
      return http.post(
        `${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/variable/batch/cluster/${varId}/`,
        data,
        config,
      );
    },
    deleteVar(context, { projectId, data }, config = {}) {
      return http.delete(
        `${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/variables/batch/?id_list=${data.id_list}`,
        {},
        config,
      );
    },
    importVars(context, { projectId, data }, config = {}) {
      return http.post(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/variables/batch/`, data, config);
    },
  },
};
