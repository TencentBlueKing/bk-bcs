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

import cookie from 'cookie';
import Vue from 'vue';
import Vuex from 'vuex';

import cloudMetadata from './cloud-metadata';

import http from '@/api';
import { getProject } from '@/api/modules/project';
import { VUEX_STROAGE_KEY } from '@/common/constant';
import { json2Query } from '@/common/util';
import VuexStorage from '@/common/vuex-storage';
import { ICluster, IProject } from '@/composables/use-app';
import app from '@/store/modules/app';
import cluster from '@/store/modules/cluster';
import clustermanager from '@/store/modules/clustermanager';
import configuration from '@/store/modules/configuration';
import crdcontroller from '@/store/modules/crdcontroller';
import dashboard from '@/store/modules/dashboard';
import depot from '@/store/modules/depot';
import helm from '@/store/modules/helm';
import hpa from '@/store/modules/hpa';
import k8sTemplate from '@/store/modules/k8s-template';
import log from '@/store/modules/log';
import mc from '@/store/modules/mc';
import metric from '@/store/modules/metric';
import network from '@/store/modules/network';
import resource from '@/store/modules/resource';
import storage from '@/store/modules/storage';
import templateset from '@/store/modules/templateset';
import token from '@/store/modules/token';
import variable from '@/store/modules/variable';
import { IMenu, MenuID } from '@/views/app/use-menu';

Vue.use(Vuex);
// cookie 中 zh-cn / en
let lang = cookie.parse(document.cookie).blueking_language || 'zh-cn';
if (['zh-CN', 'zh-cn', 'cn', 'zhCN', 'zhcn'].indexOf(lang) > -1) {
  lang = 'zh-CN';
} else {
  lang = 'en-US';
}

type FeatureFlagKey = MenuID | 'k8s' | 'VCLUSTER' | 'BKAI' | 'IMPORTSOPSCLUSTER' | 'FEDERATION';

const store = new Vuex.Store<{
  featureFlags: Record<FeatureFlagKey, boolean>
  curProject: IProject | Record<string, any>
  curCluster: ICluster
  curNav: IMenu | Record<string, any>
  curNamespace: string
  user: {
    username: string
  }
  openSideMenu: boolean
  isEn: boolean
  crdInstanceList: any[]
  dashboardViewID: string
  dashboardViewList: Array<IViewData>
  tmpViewData: IViewData | undefined // 非编辑态时临时视图数据
  editViewData: IViewData | undefined // 编辑态时视图数据
  crdData: ICrdData | undefined
  viewNsList: string[] // 集群视图命名空间缓存
  globalPageSize: number
  hideSharedCluster: boolean
  isViewEditable: boolean // 视图编辑态
  isViewConfigShow: boolean // 视图配置面板是否展开
  // 模块
  cluster: typeof cluster.state
  cloudMetadata: typeof cloudMetadata.state
}>({
  // todo 废弃模块
  modules: {
    depot,
    metric,
    mc,
    cluster,
    resource,
    app,
    variable,
    configuration,
    templateset,
    network,
    k8sTemplate,
    helm,
    hpa,
    crdcontroller,
    storage,
    dashboard,
    log,
    clustermanager,
    token,
    cloudMetadata,
  },
  plugins: [
    VuexStorage({
      key: VUEX_STROAGE_KEY,
      paths: [
        'curProject.projectID',
        'curProject.projectCode',
        'curProject.businessID',
        'curCluster.clusterID',
        'openSideMenu',
        'curNamespace',
        'globalPageSize',
        'dashboardViewID',
        'hideSharedCluster',
        'viewNsList',
      ],
      mutationEffect: [
        {
          type: 'cluster/forceUpdateClusterList',
          effect: (state, $store) => {
            const exist = state.cluster.clusterList.some(item => item.clusterID === state.curCluster?.clusterID);
            if (state.curCluster?.clusterID && !exist) {
              $store.commit('updateCurCluster');
            }
          },
        },
        {
          type: 'updateDashboardViewList',
          effect: (state, $store) => {
            const exist = state.dashboardViewList.find(item => item.id === state.dashboardViewID);
            if (!exist) {
              $store.commit('updateDashboardViewID', '');
            }
          },
        },
      ],
    }),
  ],
  // 公共 store
  state: {
    featureFlags: {},
    curProject: {},
    curCluster: {},
    curNav: {},
    curNamespace: '',
    user: {},
    openSideMenu: true, // 菜单是否折叠
    isEn: lang === 'en-US', // todo 废弃
    crdInstanceList: [], // todo 放入对应的module中
    dashboardViewID: '', // 默认视图ID
    dashboardViewList: [], // 视图列表
    tmpViewData: {},
    editViewData: {},
    crdData: {},
    viewNsList: [],
    globalPageSize: 10,
    hideSharedCluster: true,
    isViewEditable: false,
    isViewConfigShow: false,
  } as any,
  // 公共 getters
  getters: {
    user: state => state.user,
    curProjectCode: state => state.curProject?.projectCode,
    curProjectId: state => state.curProject?.projectID,
    curClusterId: state => state.curCluster?.clusterID,
    isSharedCluster: state => !!state.curCluster?.is_shared,
  },
  // 公共 mutations
  mutations: {
    /**
     * 更新当前用户 user
     *
     * @param {Object} state store state
     * @param {Object} user user 对象
     */
    updateUser(state, user) {
      state.user = Object.assign({}, user);
    },

    /**
     * 更新菜单展开状态
     * @param {*} state
     * @param {*} open
     */
    updateOpenSideMenu(state, open) {
      state.openSideMenu = !!open;
    },

    /**
     * 更新当前命名空间
     * @param {*} state
     * @param {*} name
     */
    updateCurNamespace(state, name) {
      state.curNamespace = name;
    },
    /**
     * 更改当前项目信息
     *
     * @param {Object} state store state
     * @param {String} projectId
     */
    updateCurProject(state, project) {
      state.curProject = project || {};
      window._project_code_ = project.projectCode;
      window._project_id_ =  project.projectID;
    },
    /**
     * 更新 store.cluster 中的 curCluster
     *
     * @param {Object} state store state
     * @param {Object} cluster cluster 对象
     */
    updateCurCluster(state, cluster) {
      state.curCluster = cluster || {};
    },
    /**
     * 更新crdInstanceList
     * @param {Object} state store state
     * @param {Object} data data
     */
    updateCrdInstanceList(state, data) {
      state.crdInstanceList = data;
    },
    // 更新当前一级导航信息
    updateCurNav(state, data) {
      state.curNav = data;
    },
    updateFeatureFlags(state, data) {
      state.featureFlags = data;
    },
    updateDashboardViewID(state, ID) {
      state.dashboardViewID = ID;
    },
    updateDashboardViewList(state, data) {
      state.dashboardViewList = data || [];
    },
    updatePageSize(state, size = 10) {
      state.globalPageSize = size;
    },
    updateHideClusterStatus(state, data) {
      state.hideSharedCluster = data;
    },
    updateTmpViewData(state, data) {
      state.tmpViewData = data;
    },
    updateEditViewData(state, data) {
      state.editViewData = data;
    },
    updateCrdData(state, data) {
      state.crdData = data;
    },
    updateViewEditable(state, data) {
      state.isViewEditable = data;
    },
    updateViewConfigStatus(state, data) {
      state.isViewConfigShow = !!data;
    },
    updateViewNsList(state, data = []) {
      state.viewNsList = data;
    },
  },
  actions: {
    /**
     * 获取项目信息
     *
     * @param {Object} context store 上下文对象
     * @param {Object} params 请求参数
     * @param {Object} config 请求的配置
     *
     * @return {Promise} promise 对象
     */
    getProject(context, params) {
      const { projectId } = params;
      return getProject({ $projectId: projectId }, { needRes: true, cancelWhenRouteChange: false })
        .then((res) => {
          const data = {
            ...res.data,
            cc_app_id: res.data.businessID,
            cc_app_name: res.data.businessName,
            project_id: res.data.projectID,
            project_name: res.data.name,
            project_code: res.data.projectCode,
          };
          context.commit('updateCurProject', data);
          return { data };
        })
        .catch(() => ({}));
    },

    /**
     * 获取项目日志采集信息
     *
     * @param {Object} context store 上下文对象
     * @param {string} projectId 项目 id
     * @param {Object} config 请求的配置
     *
     * @return {Promise} promise 对象
     */
    getLogPlans(context, projectId, config = {}) {
      return http.get(
        `${DEVOPS_BCS_API_URL}/api/datalog/projects/${projectId}/log_plans/`,
        {},
        config,
      );
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
    getBcsCrdsList(context, { projectId, clusterId, crdKind, params = {} }, config = {}) {
      context.commit('updateCrdInstanceList', []);
      const url = `${DEVOPS_BCS_API_URL}/api/bcs_crd/projects/${projectId}/clusters/${clusterId}/crds/${crdKind}/custom_objects/?${json2Query(params, '')}`;
      return http.get(url, {}, config).then((res) => {
        context.commit('updateCrdInstanceList', res.data);
        return res;
      });
    },
  },
});

export default store;
