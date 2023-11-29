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
import {
  fetchNodePodsData,
} from '@/api/base';
import { fetchClusterList } from '@/api/modules/cluster-manager';
import { projectBusiness } from '@/api/modules/project';
import { json2Query } from '@/common/util';

export default {
  namespaced: true,
  state: {
    // 当前项目下的集群列表，全局 store 中只会存在一个 clusterList，因为每个涉及到集群的模块都是和 projectId 有关的
    // 所以每次改变项目 projectId 的时候，这里的 clusterList 会重新获取
    clusterList: [],
    isClusterDataReady: false,
    clusterWebAnnotations: { perms: {} },
    maintainers: [],
  },
  mutations: {
    /**
         * 更新 store.cluster 中的 clusterList
         *
         * @param {Object} state store state
         * @param {Array} list cluster 列表
         */
    forceUpdateClusterList(state, list) {
      const clusterList = list.map(item => ({
        cluster_id: item.clusterID,
        name: item.clusterName,
        project_id: item.projectID,
        ...item,
      }));
      // const sortData = sort(clusterList, 'clusterName');
      state.clusterList.splice(0, state.clusterList.length, ...clusterList);
      state.isClusterDataReady = true;
    },
    updateClusterWebAnnotations(state, data) {
      state.clusterWebAnnotations = data;
    },
    updateMaintainers(state, data) {
      state.maintainers = data;
    },
  },
  actions: {
    /**
         * 根据项目 id 获取项目下所有集群
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    async getClusterList(context) {
      const res = await fetchClusterList({}, {
        needRes: true,
        cancelWhenRouteChange: false,
      }).catch(() => ({ data: [] }));
      const clusterExtraInfo = res.clusterExtraInfo || {};
      // 兼容以前集群数据
      res.data = res.data.map(item => ({
        cluster_id: item.clusterID,
        name: item.clusterName,
        project_id: item.projectID,
        ...item,
        ...clusterExtraInfo[item.clusterID],
      }));
      context.commit('forceUpdateClusterList', res?.data || []);
      context.commit('updateClusterWebAnnotations', res.web_annotations || { perms: {} });
      return res;
    },

    /**
         * 根据项目 id 获取项目下有权限的集群
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    getPermissionClusterList(context, projectId, config = {}) {
      return http.get(
        `${DEVOPS_BCS_API_URL}/api/projects/${projectId}/clusters?limit=1000&perm_can_use=1`,
        {},
        config,
      );
    },

    /**
         * 创建集群时获取所属地域信息
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    getAreaList(context, params, config = {}) {
      return http.get(`${DEVOPS_BCS_API_URL}/api/projects/${params.projectId}/areas?${json2Query(params.data)}`, {}, config);
    },

    /**
         * TKE 创建集群时获取所属 VPC 信息
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 请求参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    getVPC(context, params, config = {}) {
      const { projectId } = params;
      return http.get(`${DEVOPS_BCS_API_URL}/api/projects/${projectId}/vpcs/?${json2Query(params.data)}`, {}, config);
    },

    /**
         * 创建集群时获取备选服务器列表
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    getCCHostList(context, params, config = {}) {
      const { projectId } = params;
      delete params.projectId;
      // return http.post(`/app/cluster?invoke=getCCHostList&${projectId}`, params, config)
      return http.post(`${DEVOPS_BCS_API_URL}/api/projects/${projectId}/cc_host_info/`, params, config);
    },

    /**
         * 创建集群
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    createCluster(context, params, config = {}) {
      const { projectId } = params;
      delete params.projectId;
      // return http.post(`/api/projects/cluster?invoke=createCluster`, params).then(response => {
      return http.post(`${DEVOPS_BCS_API_URL}/api/projects/${projectId}/clusters`, params, config);
    },

    /**
         * 获取集群详细信息
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {string} clusterId 集群 id
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    getCluster(context, { projectId, clusterId }, config = {}) {
      return http.get(`${DEVOPS_BCS_API_URL}/api/projects/${projectId}/cluster/${clusterId}`, {}, config);
    },

    /**
         * 更新集群信息
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {string} clusterId 集群 id
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    updateCluster(context, { projectId, clusterId, data }, config = {}) {
      return http.put(
        `${DEVOPS_BCS_API_URL}/api/projects/${projectId}/cluster/${clusterId}`,
        data,
        config,
      );
    },

    /**
         * 重新初始化集群
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {string} clusterId 集群 id
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    reInitializationCluster(context, { projectId, clusterId }, config = {}) {
      // return http.get(`/api/projects/cluster?invoke=reInitializationCluster`).then(response => {
      return http.post(`${DEVOPS_BCS_API_URL}/api/projects/${projectId}/cluster/${clusterId}`, {}, config);
    },

    /**
         * 获取集群总览信息
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {string} clusterId 集群 id
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    getClusterOverview(context, { projectId, clusterId }, config = {}) {
      // return http.get(`/api/projects/cluster?invoke=getClusterOverview`
      return http.get(
        `${DEVOPS_BCS_API_URL}/api/projects/${projectId}/metrics/cluster/?res_id=${clusterId}&metric=cpu_summary`,
        {},
        config,
      );
    },

    /**
         * 集群总览页面获取下面三个圈的数据
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {string} clusterId 集群 id
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    getClusterMetrics(context, { projectId, clusterId }, config = {}) {
      // return http.get(`/api/projects/cluster?invoke=getClusterMetrics`
      return http.get(
        `${DEVOPS_BCS_API_URL}/api/projects/${projectId}/metrics/cluster/summary/?res_id=${clusterId}`,
        {},
        config,
      );
    },

    /**
         * 集群 节点列表获取数据
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    getNodeList(context, params, config = {}) {
      const { projectId } = params;
      const { clusterId } = params;
      params.ip = params.ip || '';
      delete params.projectId;
      delete params.clusterId;


      return http.get(
        `${DEVOPS_BCS_API_URL}/api/projects/${projectId}/cluster_nodes/${clusterId}?with_containers=1&${json2Query(params)}`,
        params,
        config,
      );
    },

    /**
         * 集群 节点列表 搜索 查询节点标签的 key
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    getNodeKeyList(context, params, config = {}) {
      const { projectId } = params;
      const { clusterId } = params;
      return http.get(
        `${DEVOPS_BCS_API_URL}/api/projects/${projectId}/nodes/label_keys/?cluster_id=${clusterId}`,
        params,
        config,
      );
    },

    /**
         * 集群 节点列表 搜索 根据节点标签的 key 查询节点标签的 value
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    getNodeValueListByKey(context, params, config = {}) {
      const { projectId } = params;
      const { clusterId } = params;
      const { keyName } = params;
      return http.get(
        `${DEVOPS_BCS_API_URL}/api/projects/${projectId}/nodes/label_values/?cluster_id=${clusterId}&key_name=${keyName}`,
        params,
        config,
      );
    },

    /**
         * 集群 节点列表 批量操作
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    batchNode(context, params, config = {}) {
      const { projectId } = params;
      const { operateType } = params;
      const { clusterId } = params;
      const { ipList } = params;
      const { status } = params;

      // 删除
      if (operateType === '3') {
        return http.delete(
          `${DEVOPS_BCS_API_URL}/api/projects/${projectId}/clusters/${clusterId}/nodes/batch/`,
          { data: { inner_ip_list: ipList } },
          config,
        );
      }

      return http.put(
        `${DEVOPS_BCS_API_URL}/api/projects/${projectId}/clusters/${clusterId}/nodes/batch/`,
        { inner_ip_list: ipList, status },
        config,
      );
    },

    /**
         * 集群 节点列表 批量 重新添加 操作
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    batchNodeReInstall(context, params, config = {}) {
      const { projectId, clusterId } = params;
      delete params.projectId;
      delete params.clusterId;

      return http.post(
        `${DEVOPS_BCS_API_URL}/api/projects/${projectId}/clusters/${clusterId}/nodes/reinstall/`,
        params,
        config,
      );
    },

    /**
         * 移除节点
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {string} clusterId 集群 id
         * @param {string} nodeId 节点 id 即 ip
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    removeNode(context, { projectId, clusterId, nodeId }, config = {}) {
      // return http.delete(`/app/cluster?invoke=removeNode`)
      return http.delete(
        `${DEVOPS_BCS_API_URL}/api/projects/${projectId}/cluster/${clusterId}/node/${nodeId}`,
        {},
        config,
      );
    },

    /**
         * 强制移除节点
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {string} clusterId 集群 id
         * @param {string} nodeId 节点 id 即 ip
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    forceRemoveNode(context, { projectId, clusterId, nodeId }, config = {}) {
      // return http.delete(`/app/cluster?invoke=forceRemoveNode`)
      return http.delete(
        `${DEVOPS_BCS_API_URL}/api/projects/${projectId}/clusters/${clusterId}/nodes/${nodeId}/force_delete/`,
        {},
        config,
      );
    },

    /**
         * 重新初始化节点
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {string} clusterId 集群 id
         * @param {string} nodeId 节点 id 即 ip
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    reInitializationNode(context, { projectId, clusterId, nodeId }, config = {}) {
      // return http.post(`/api/projects/cluster?invoke=reInitializationNode`).then(response => {
      //     return response.data
      // })
      return http.post(
        `${DEVOPS_BCS_API_URL}/api/projects/${projectId}/cluster/${clusterId}/node/${nodeId}`,
        {},
        config,
      );
    },

    /**
         * 查询节点日志
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {string} clusterId 集群 id
         * @param {string} nodeId 节点 id 即 ip
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    getNodeLogs(context, { projectId, clusterId, nodeId }, config = {}) {
      // return http.get(`/app/cluster?invoke=getNodeLogs`)
      return http.get(
        `${DEVOPS_BCS_API_URL}/api/projects/${projectId}/cluster/${clusterId}/node/${nodeId}/logs`,
        {},
        config,
      );
    },

    /**
         * 集群 节点列表获取 cpu 磁盘 内存占用的数据
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {string} nodeId 节点 id 即 ip
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    getNodeSummary(context, { projectId, nodeId }, config = {}) {
      // return http.get(`/app/cluster?invoke=getNodeSummary&${projectId}&${nodeId}`, {}, config)
      return http.get(
        `${DEVOPS_BCS_API_URL}/api/projects/${projectId}/metrics/node/summary/?res_id=${nodeId}`,
        {},
        config,
      );
    },

    /**
         * 集群 节点详情 上方数据
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {string} clusterId 集群 id
         * @param {string} nodeId node id 即 ip
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    getNodeInfo(context, { projectId, clusterId, nodeId }, config = {}) {
      return http.get(
        `${DEVOPS_BCS_API_URL}/api/projects/${projectId}/cluster/${clusterId}/node/info/?res_id=${nodeId}`,
        {},
        config,
      );
    },

    /**
         * 节点详情页面 下方容器选项卡 表格数据接口
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {string} clusterId 集群 id
         * @param {string} nodeId node id 即 ip
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    getNodeContainerList(context, { projectId, clusterId, nodeId }, config = {}) {
      return http.get(
        `${DEVOPS_BCS_API_URL}/api/projects/${projectId}/cluster/${clusterId}/node/containers/?res_id=${nodeId}`,
        {},
        config,
      );
    },

    /**
         * 节点 overview 页面下方表格点击容器获取容器详情
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    getContainerInfo(context, { projectId, clusterId, containerId }, config = {}) {
      return http.get(
        `${DEVOPS_BCS_API_URL}/api/app/projects/${projectId}/clusters/${clusterId}/container/`
                    + `?container_id=${containerId}`,
        {},
        config,
      );
    },

    /**
         * 查询节点标签
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    getNodeLabel(context, { projectId, nodeIds }, config = {}) {
      return http.get(
        `${DEVOPS_BCS_API_URL}/api/projects/${projectId}/node_label_info/?node_ids=${nodeIds}`,
        {},
        config,
      );
    },

    /**
         * 节点页面获取所有节点
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    getAllNodeList(context, { projectId }, config = {}) {
      // return http.get('/app/cluster?invoke=getAllNodeList', {}, config)
      return http.get(`${DEVOPS_BCS_API_URL}/api/projects/${projectId}/node_label_list/`, {}, config);
    },

    /**
         * 移除节点 /api/projects/(?P[\w-]+)/cluster/(?P[\w-]+)/node/(?P\d+)/failed_delete/
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {string} clusterId 集群 id
         * @param {string} nodeId 节点 id 即 ip
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    failedDelNode(context, { projectId, clusterId, nodeId }, config = {}) {
      // return http.delete(`/app/cluster?invoke=failedDelNode`)
      return http.delete(
        `${DEVOPS_BCS_API_URL}/api/projects/${projectId}/cluster/${clusterId}/node/${nodeId}/failed_delete/`,
        {},
        config,
      );
    },

    /**
         * 设置 helm enable
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {string} clusterId 集群 id
         * @param {string} nodeId 节点 id 即 ip
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    enableClusterHelm(context, { projectId, clusterId }, config = {}) {
      const url = `${DEVOPS_BCS_API_URL}/api/bcs/k8s/configuration/${projectId}/helm_init/?cluster_id=${clusterId}`;
      return http.post(url, { cluster_id: clusterId }, config);
    },

    /**
         * 移除故障节点 /api/projects/(?P[\w-]+)/clusters/(?P[\w-]+)/nodes/(?P\d+)/
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {string} clusterId 集群 id
         * @param {string} nodeId 节点 id 即 ip
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    faultRemoveNode(context, { projectId, clusterId, nodeId }, config = {}) {
      return http.delete(
        `${DEVOPS_BCS_API_URL}/api/projects/${projectId}/clusters/${clusterId}/nodes/${nodeId}/`,
        {},
        config,
      );
    },

    /**
         * 获取 tke 配置信息
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    getTKEConf(context, { projectId }, config = {}) {
      return http.get(`${DEVOPS_BCS_API_URL}/api/projects/${projectId}/tke_conf/?coes=tke`, {}, config);
    },

    /**
         * 获取 k8s version 配置信息
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    getK8SConf(context, { projectId }, config = {}) {
      return http.get(`${DEVOPS_BCS_API_URL}/api/projects/${projectId}/cluster_type_versions/?coes=k8s`, {}, config);
    },

    // ------------------------------------------------------------------------------------------------ //

    /**
         * 集群使用率概览
         * /api/metrics/projects/{project_id}/clusters/{cluster_id}/overview/
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    clusterOverview(context, { projectId, clusterId }, config = {}) {
      return http.post(
        `${DEVOPS_BCS_API_URL}/api/metrics/projects/${projectId}/clusters/${clusterId}/overview/`,
        {},
        config,
      );
    },

    /**
         * 集群CPU使用率
         * /api/metrics/projects/{project_id}/clusters/{cluster_id}/cpu_usage/
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    clusterCpuUsage(context, { projectId, clusterId }, config = {}) {
      return http.get(
        `${DEVOPS_BCS_API_URL}/api/metrics/projects/${projectId}/clusters/${clusterId}/cpu_usage/`,
        {},
        config,
      );
    },

    /**
         * 集群内存使用率
         * /api/metrics/projects/{project_id}/clusters/{cluster_id}/memory_usage/
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    clusterMemUsage(context, { projectId, clusterId }, config = {}) {
      return http.get(
        `${DEVOPS_BCS_API_URL}/api/metrics/projects/${projectId}/clusters/${clusterId}/memory_usage/`,
        {},
        config,
      );
    },

    /**
         * 集群磁盘使用率
         * /api/metrics/projects/{project_id}/clusters/{cluster_id}/disk_usage/
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    clusterDiskUsage(context, { projectId, clusterId }, config = {}) {
      return http.get(
        `${DEVOPS_BCS_API_URL}/api/metrics/projects/${projectId}/clusters/${clusterId}/disk_usage/`,
        {},
        config,
      );
    },

    /**
         * 集群 节点列表 节点概览
         * /api/metrics/projects/{project_id}/clusters/{cluster_id}/node/overview/
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {string} nodeId 节点 id 即 ip
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    getNodeOverview(context, { projectId, clusterId, nodeIp, data = {} }, config = {}) {
      return http.post(
        `${DEVOPS_BCS_API_URL}/api/metrics/projects/${projectId}/clusters/${clusterId}/nodes/${nodeIp}/overview/`,
        data,
        config,
      );
    },

    /**
         * node CPU使用率
         * /api/metrics/projects/{project_id}/clusters/{cluster_id}/nodes/{node_ip}/cpu_usage/
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    nodeCpuUsage(context, params, config = {}) {
      const { projectId, clusterId } = params;

      delete params.projectId;
      delete params.clusterId;

      const list = Object.keys(params);
      const len = list.length;

      for (let i = 0; i < len; i++) {
        const key = list[i];
        const value = params[key];
        if (value === null || value === '' || key === 'projectId') {
          delete params[key];
          continue;
        }
        delete params[key];
        params[_.snakeCase(key)] = value;
      }

      return http.get(
        `${DEVOPS_BCS_API_URL}/api/metrics/projects/${projectId}/clusters/${clusterId}/nodes/${params.res_id}/cpu_usage/?${json2Query(params)}`,
        {},
        config,
      );
    },

    /**
         * node mem使用率
         * /api/metrics/projects/{project_id}/clusters/{cluster_id}/nodes/{node_ip}/memory_usage/
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    nodeMemUsage(context, params, config = {}) {
      const { projectId, clusterId } = params;

      delete params.projectId;
      delete params.clusterId;

      const list = Object.keys(params);
      const len = list.length;

      for (let i = 0; i < len; i++) {
        const key = list[i];
        const value = params[key];
        if (value === null || value === '' || key === 'projectId') {
          delete params[key];
          continue;
        }
        delete params[key];
        params[_.snakeCase(key)] = value;
      }

      return http.get(
        `${DEVOPS_BCS_API_URL}/api/metrics/projects/${projectId}/clusters/${clusterId}/nodes/${params.res_id}/memory_usage/?${json2Query(params)}`,
        {},
        config,
      );
    },

    /**
         * node diskio 使用率
         * /api/metrics/projects/{project_id}/clusters/{cluster_id}/nodes/{node_ip}/diskio_usage/
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    nodeDiskioUsage(context, params, config = {}) {
      const { projectId, clusterId } = params;

      delete params.projectId;
      delete params.clusterId;

      const list = Object.keys(params);
      const len = list.length;

      for (let i = 0; i < len; i++) {
        const key = list[i];
        const value = params[key];
        if (value === null || value === '' || key === 'projectId') {
          delete params[key];
          continue;
        }
        delete params[key];
        params[_.snakeCase(key)] = value;
      }

      return http.get(
        `${DEVOPS_BCS_API_URL}/api/metrics/projects/${projectId}/clusters/${clusterId}/nodes/${params.res_id}/diskio_usage/?${json2Query(params)}`,
        {},
        config,
      );
    },

    /**
         * node net 入流量
         * /api/metrics/projects/{project_id}/clusters/{cluster_id}/nodes/{node_ip}/network_receive/
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    nodeNetReceive(context, params, config = {}) {
      const { projectId, clusterId } = params;

      delete params.projectId;
      delete params.clusterId;

      const list = Object.keys(params);
      const len = list.length;

      for (let i = 0; i < len; i++) {
        const key = list[i];
        const value = params[key];
        if (value === null || value === '' || key === 'projectId') {
          delete params[key];
          continue;
        }
        delete params[key];
        params[_.snakeCase(key)] = value;
      }

      return http.get(
        `${DEVOPS_BCS_API_URL}/api/metrics/projects/${projectId}/clusters/${clusterId}/nodes/${params.res_id}/network_receive/?${json2Query(params)}`,
        {},
        config,
      );
    },

    /**
         * node net 出流量
         * /api/metrics/projects/{project_id}/clusters/{cluster_id}/nodes/{node_ip}/network_transmit/
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    nodeNetTransmit(context, params, config = {}) {
      const { projectId, clusterId } = params;

      delete params.projectId;
      delete params.clusterId;

      const list = Object.keys(params);
      const len = list.length;

      for (let i = 0; i < len; i++) {
        const key = list[i];
        const value = params[key];
        if (value === null || value === '' || key === 'projectId') {
          delete params[key];
          continue;
        }
        delete params[key];
        params[_.snakeCase(key)] = value;
      }

      return http.get(
        `${DEVOPS_BCS_API_URL}/api/metrics/projects/${projectId}/clusters/${clusterId}/nodes/${params.res_id}/network_transmit/?${json2Query(params)}`,
        {},
        config,
      );
    },

    /**
         * 集群 节点详情 上方数据，prometheus 获取
         * /api/metrics/projects/{project_id}/clusters/{cluster_id}/nodes/{node_ip}/info/
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {string} clusterId 集群 id
         * @param {string} nodeId node id 即 ip
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    nodeInfo(context, { projectId, clusterId, nodeId }, config = {}) {
      return http.get(
        `${DEVOPS_BCS_API_URL}/api/metrics/projects/${projectId}/clusters/${clusterId}/nodes/${nodeId}/info/`,
        {},
        config,
      );
    },

    /**
         * 获取集群可升级的版本
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {string} clusterId 集群 id
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    getClusterVersion(context, { projectId, clusterId }, config = {}) {
      return http.get(
        `${DEVOPS_BCS_API_URL}/api/projects/${projectId}/clusters/${clusterId}/upgradeable_versions/`,
        {},
        config,
      );
    },

    /**
         * 升级集群信息
         *
         * @param {Object} context store 上下文对象
         * @param {string} projectId 项目 id
         * @param {string} clusterId 集群 id
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    upgradeCluster(context, { projectId, clusterId, data }, config = {}) {
      return http.put(
        `${DEVOPS_BCS_API_URL}/api/projects/${projectId}/clusters/${clusterId}/version/`,
        data,
        config,
      );
    },

    /**
         * 获取 SCR 主机
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    getSCRHosts(context, params, config = {}) {
      const { projectId } = params;
      delete params.projectId;
      return http.post(
        `${DEVOPS_BCS_API_URL}/api/hosts/projects/${projectId}/cvm_types/`,
        params,
        config,
      );
    },

    /**
         * 申请主机
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    applySCRHost(context, params, config = {}) {
      const { projectId } = params;
      delete params.projectId;
      return http.post(
        `${DEVOPS_BCS_API_URL}/api/hosts/projects/${projectId}/apply_hosts/`,
        params,
        config,
      );
    },

    /**
         * 查看主机申请状态
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    checkApplyHostStatus(context, { projectId }, config = {}) {
      return http.get(
        `${DEVOPS_BCS_API_URL}/api/hosts/projects/${projectId}/apply_hosts/logs/`,
        {},
        config,
      );
    },

    /**
         * 查看主机权限
         *
         * @param {Object} context store 上下文对象
         * @param {Object} params 参数
         * @param {Object} config 请求的配置
         *
         * @return {Promise} promise 对象
         */
    async getBizMaintainers(context, params = {}, config = { cancelWhenRouteChange: false }) {
      const data = await projectBusiness(params, config).catch(() => ({
        maintainer: [],
      }));
      context.commit('updateMaintainers', data.maintainer || []);
      return data;
    },

    /**
         * 获取园区列表
         * @param {String} projectId 项目ID
         * @param {String} region 所属地域
         */
    async getZoneList(context, { projectId, region }, config = {}) {
      return http.get(
        `${DEVOPS_BCS_API_URL}/api/hosts/projects/${projectId}/zones/?region=${region}`,
        config,
      );
    },

    /**
         * 获取数据盘类型列表
         */
    async getDiskTypeList(context, { projectId }, config = {}) {
      return http.get(
        `${DEVOPS_BCS_API_URL}/api/hosts/projects/${projectId}/disk_types/`,
        config,
      );
    },

    async fetchPodsData(context, params) {
      const data = fetchNodePodsData(params, { needRes: true });
      return data;
    },
  },
};
