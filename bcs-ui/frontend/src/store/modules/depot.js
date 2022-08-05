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

export default {
  namespaced: true,
  state: {
    // 公共镜像
    imageLibrary: {
      // 存储列表数据
      dataList: [],
    },
    // 项目镜像
    projectImage: {
      // 存储列表数据
      dataList: [],
    },
    // 当前镜像
    curImage: null,
    // 我的收藏
    myCollect: {
      // 存储列表数据
      dataList: [],
    },
  },
  mutations: {
    /**
         * 更新 store 中的 imageLibrary.dataList
         *
         * @param {Object} state store state
         * @param {list} list 项目列表
         */
    forceUpdateImageLibraryList(state, list) {
      state.imageLibrary.dataList.splice(0, state.imageLibrary.dataList.length, ...list);
    },

    /**
         * 更新 store 中的 projectImage.dataList
         *
         * @param {Object} state store state
         * @param {list} list 项目列表
         */
    forceUpdateProjectImageList(state, list) {
      state.projectImage.dataList.splice(0, state.projectImage.dataList.length, ...list);
    },

    /**
         * 更新 store 中的 myCollect.dataList
         *
         * @param {Object} state store state
         * @param {list} list 项目列表
         */
    forceUpdateMyCollectList(state, list) {
      state.myCollect.dataList.splice(0, state.myCollect.dataList.length, ...list);
    },

    /**
         * 更新 store 中的 当前镜像信息
         *
         * @param {Object} state store state
         * @param {Object} data 当前镜像信息
         */
    forceUpdateCurImage(state, data) {
      state.curImage = data;
    },
  },
  actions: {
    /**
         * 获取公共镜像
         * /api/depot/public/?search_key=jdk_onion
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    getImageLibrary(context, params, config = {}) {
      // return http.get(`/app/depot?invoke=getImageLibrary&${json2Query(params)}`, params, config).then(
      return http.get(`${DEVOPS_BCS_API_URL}/api/depot/images/public/?${json2Query(params)}`, params, config).then((response) => {
        const res = response.data;
        context.commit('forceUpdateImageLibraryList', res.results || []);
        return res;
      });
    },

    /**
         * 获取项目镜像
         * /api/depot/project/000?search_key=jdk_onion
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    getProjectImage(context, params, config = {}) {
      const { projId } = params;
      delete params.projId;
      return http.get(
        `${DEVOPS_BCS_API_URL}/api/depot/images/project/${projId}/?${json2Query(params)}`,
        params,
        config,
      ).then((response) => {
        const res = response.data;
        context.commit('forceUpdateProjectImageList', res.results || []);
        return res;
      });
    },

    /**
         * 获取我的镜像详情
         * /api/depot/collect/?search_key=jdk_onion
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    getImageLibraryDetail(context, params, config = {}) {
      const { projectId } = params;
      delete params.projectId;
      return http.get(
        `${DEVOPS_BCS_API_URL}/api/depot/images/project/${projectId}/info/image/?${json2Query(params)}`,
        params,
        config,
      );
    },
  },
};
