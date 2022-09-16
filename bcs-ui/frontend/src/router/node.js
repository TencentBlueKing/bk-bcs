/**
 * Tencent is pleased to support the open source community by making 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition) available.
 * Copyright (C) 2017-2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

const Node = () => import(/* webpackChunkName: 'node' */'@/views/node/node.vue')
const AutoScalerConfig = () => import(/* webpackChunkName: 'node' */'@/views/node/autoscaler-config.vue')
const NodePool = () => import(/* webpackChunkName: 'node' */'@/views/node/node-pool.vue')
const NodePoolDetail = () => import(/* webpackChunkName: 'node' */'@/views/node/node-pool-detail.vue')
const EditNodePool = () => import(/* webpackChunkName: 'node' */'@/views/node/edit-node-pool.vue')
 
const childRoutes = [
    // domain/bcs/projectCode/node 节点页面
    {
        path: ':projectCode/node',
        name: 'nodeMain',
        component: Node,
        meta: {
            title: window.i18n.t('节点'),
            hideBack: true
        }
    },
    {
        path: ':projectCode/cluster/:clusterId/autoscaler',
        name: 'autoScalerConfig',
        props: true,
        component: AutoScalerConfig
    },
    {
        path: ':projectCode/cluster/:clusterId/nodepool',
        name: 'nodePool',
        props: true,
        component: NodePool
    },
    {
        path: ':projectCode/cluster/:clusterId/nodepool/:nodeGroupID',
        name: 'editNodePool',
        props: true,
        component: EditNodePool
    },
    {
        path: ':projectCode/cluster/:clusterId/nodepool/detail/:nodeGroupID',
        name: 'nodePoolDetail',
        props: true,
        component: NodePoolDetail
    }
]
 
export default childRoutes
