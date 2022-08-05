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

import _ from 'lodash';

import http from '@/api';
import { json2Query } from '@/common/util';

export default {
  namespaced: true,
  state: {
    // 操作审计
    operateAudit: {
      // 表格返回数据
      data: {},
    },
    // 事件查询
    eventQuery: {
      // 表格返回数据
      data: {},
    },
  },
  mutations: {
  },
  actions: {
    /**
         * 操作审计页面获取所有操作对象类型
         * /api/activity_logs/resource_types
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    getResourceTypes(context, params, config = {}) {
      return http.get(
        `${DEVOPS_BCS_API_URL}/api/activity_logs/resource_types?${json2Query(params)}`,
        params,
        config,
      );
    },

    /**
         * 获取操作审计数据
         * /api/activity_logs/project/000?begin_time=2017-09-29T08:45:00&end_time=2017-09-30T09:46:00
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    getActivityLogs(context, params, config = {}) {
      const { projId } = params;

      const list = Object.keys(params);
      const len = list.length;

      for (let i = len - 1; i >= 0; i--) {
        const key = list[i];
        const value = params[key];
        if (value === null || value === '' || value === 'all' || value === '全部' || key === 'projId') {
          delete params[key];
          continue;
        }
        delete params[key];
        params[_.snakeCase(key)] = value;
      }

      delete params.projId;
      // return http.get(`/api/activity_logs/project/${projId}`, {
      return http.get(
        `${DEVOPS_BCS_API_URL}/api/activity_logs/project/${projId}?${json2Query(params)}`,
        params,
        config,
      ).then((response) => {
        const res = response.data;
        context.state.operateAudit.data = res.data;
        return res;
      });
    },

    /**
         * 获取事件查询数据
         * /api/activity_events/project/000?begin_time=2017-09-29T08:45:00&end_time=2017-09-30T09:46:00
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    getActivityEvents(context, params, config = {}) {
      const { projId } = params;
      const list = Object.keys(params);
      const len = list.length;

      for (let i = len - 1; i >= 0; i--) {
        const key = list[i];
        const value = params[key];
        if (value === null || value === '' || value === 'all' || value === '全部' || key === 'projId') {
          delete params[key];
          continue;
        }
        delete params[key];
        params[_.snakeCase(key)] = value;
      }

      return http.get(
        `${DEVOPS_BCS_API_URL}/api/activity_events/project/${projId}?${json2Query(params)}`,
        params,
        config,
      ).then((response) => {
        const res = response.data;
        context.state.eventQuery.data = res.data;
        return res;
      });
    },
  },
};
