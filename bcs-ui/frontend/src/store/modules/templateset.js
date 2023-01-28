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
  },
  mutations: {
  },
  actions: {

    /**
         * 获取模板集列表
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数，包含：projectId ,queryString: ?offset=0&limit=20&search=2730）
         * @param {string} templateId template id
         *
         * @return {Promise} promise 对象
         */
    getTemplateList(context, { projectId, queryString }, config = {}) {
      let url = `${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/templates/`;
      if (queryString) {
        url = `${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/templates/?${queryString}`;
      }
      return http.get(url, {}, config);
    },

    /**
         * 删除模板集
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {string} templateId template id
         *
         * @return {Promise} promise 对象
         */
    removeTemplate(context, { templateId, projectId }, config = {}) {
      return http.delete(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/template/${templateId}/`, {}, config);
    },

    /**
         * 获取模板集版本
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {string} templateId template id
         *
         * @return {Promise} promise 对象
         */
    getTemplatesetVerList(context, { projectId, templateId, hasFilter }, config = {}) {
      let urlPrefix = 'show/versions';
      // 删除实例使用独立的url
      if (hasFilter) {
        urlPrefix = 'exist/show_version_name';
      }
      return http.get(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/${urlPrefix}/${templateId}/`, {}, config);
    },

    /**
         * 根据模板集 版本下已经实例化过的资源
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {string} templateId template id
         *
         * @return {Promise} promise 对象
         */
    getTemplateInsResourceById(context, { projectId, templateId, showVerName }, config = {}) {
      return http.get(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/exist/resource/${templateId}/?show_version_name=${showVerName}`, {}, config);
    },

    /**
         * 根据模板集 id 获取模板
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {string} tplVerId template version id
         *
         * @return {Promise} promise 对象
         */
    getTemplateListById(context, { projectId, tplVerId }, config = {}) {
      return http.get(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/resource/${tplVerId}/`, {}, config);
    },

    /**
         * 根据模板集 id, category, tmpl_app_name 获取模板
         * 用于应用列表跳转到实例化页面
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {string} tplVerId template version id
         * @param {string} tmplAppName template appName
         * @param {string} category category category
         *
         * @return {Promise} promise 对象
         */
    getTemplateListByIdCategoryTmplName(context, { projectId, tplVerId, tmplAppName, category }, config = {}) {
      return http.get(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/resource/${tplVerId}/` + `?category=${category}&tmpl_app_name=${tmplAppName}`, {}, config);
    },

    /**
         * 获取命名空间
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {boolean} isGroupBy 是否需要 group_by
         *
         * @return {Promise} promise 对象
         */
    getNamespaceList(context, { projectId, isGroupBy }, config = {}) {
      return http.get(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/namespace/${isGroupBy ? '?group_by=cluster_name' : ''}`, {}, config);
    },

    /**
         * 预览配置
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         *
         * @return {Promise} promise 对象
         */
    previewNamespace(context, params, config = {}) {
      const { projectId } = params;
      delete params.projectId;
      return http.post(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/preview/`, params, config).then(response => response.data);
    },

    /**
         * 预览配置
         * 用于应用列表跳转到实例化页面
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         *
         * @return {Promise} promise 对象
         */
    previewNamespace1(context, params, config = {}) {
      const { projectId } = params;
      delete params.projectId;
      return http.put(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/app/instantiation/`, params, config).then(response => response.data);
    },

    /**
         * 模板实例化，创建实例化
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         *
         * @return {Promise} promise 对象
         */
    createInstance(context, params, config = {}) {
      const { projectId } = params;
      delete params.projectId;
      return http.post(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/instances/`, params).then(response => response.data);
    },

    /**
         * 模板实例化，创建实例化
         * 用于应用列表跳转到实例化页面
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         *
         * @return {Promise} promise 对象
         */
    createInstance1(context, params, config = {}) {
      const { projectId } = params;
      delete params.projectId;
      return http.post(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/app/instantiation/`, params).then(response => response.data);
    },

    /**
         * 预查询 命名空间 下的 loadbalance 信息览配置
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         *
         * @return {Promise} promise 对象
         */
    getLbInfo(context, { projectId, namespaceId }, config = {}) {
      return http.get(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/loadbalance/${namespaceId}/`).then(response => response.data);
    },

    /**
         * 添加命名空间
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         *
         * @return {Promise} promise 对象
         */
    addNamespace(context, params, config = {}) {
      const { projectId } = params;
      delete params.projectId;
      return http.post(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/namespace/`, params).then(response => response.data);
    },

    /**
         * 修改命名空间
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         *
         * @return {Promise} promise 对象
         */
    editNamespace(context, params, config = {}) {
      const { projectId, namespaceId } = params;
      delete params.projectId;
      delete params.namespaceId;
      return http.put(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/namespace/${namespaceId}/`, params).then(response => response.data);
    },

    /**
         * 删除命名空间
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         *
         * @return {Promise} promise 对象
         */
    delNamespace(context, params, config = {}) {
      const { projectId, namespaceId } = params;
      delete params.projectId;
      delete params.namespaceId;
      return http.delete(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/namespace/${namespaceId}/`, params).then(response => response.data);
    },

    /**
         * 删除实例时的获取命名空间集合接口
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
    getNamespaceList4DelInstance(context, { projectId, tplMusterId, tplsetVerId, tplId, category }, config = {}) {
      return http.get(`${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/musters/${tplMusterId}/instances/`
                    + `namespaces/?perm_can_use=1&group_by=cluster_id&category=${category}&`
                    + `show_version_name=${tplsetVerId}&res_name=${tplId}`);
    },

    /**
         * 反实例化删除命名空间
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         *
         * @return {Promise} promise 对象
         */
    delNamespaceInDelInstance(context, params, config = {}) {
      const { projectId, tplMusterId, tplsetVerId, tplId } = params;
      delete params.projectId;
      delete params.tplMusterId;
      delete params.tplsetVerId;
      delete params.tplId;
      return http.delete(`${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/musters/${tplMusterId}/instances/resources/?show_version_name=${tplsetVerId}&res_name=${tplId}`, { data: params }, config);
    },

    /**
         * 复制模板集
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         *
         * @return {Promise} promise 对象
         */
    copyTemplate(context, params, config = {}) {
      const { projectId, templateId } = params;
      delete params.projectId;
      delete params.templateId;
      return http.put(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/template/${templateId}/`, params, config);
    },

    /**
         * 删除模板集时获取 exist_version
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
    getExistVersion(context, { projectId, templateId }, config = {}) {
      return http.get(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/template/exist_versions/${templateId}/`);
    },

    /**
         * 查询 lb 和 variable
         *
         * @param {Object} context store 上下文对象
         *
         * @return {Promise} promise 对象
         */
    getLbVariable(context, { projectId, tplVerId, namespaces, instanceEntity }, config = {}) {
      return http.post(`${DEVOPS_BCS_API_URL}/api/configuration/${projectId}/variable/resource/${tplVerId}/`, {
        namespaces,
        instance_entity: instanceEntity,
      });
    },
  },
};
