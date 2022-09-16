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

import http from '@/api'
import { json2Query } from '@/common/util'

export default {
    namespaced: true,
    state: {
        // 公共镜像
        imageLibrary: {
            // 存储列表数据
            dataList: []
        },
        // 项目镜像
        projectImage: {
            // 存储列表数据
            dataList: []
        },
        // 当前镜像
        curImage: null,
        // 我的收藏
        myCollect: {
            // 存储列表数据
            dataList: []
        }
    },
    mutations: {
        /**
         * 更新 store 中的 imageLibrary.dataList
         *
         * @param {Object} state store state
         * @param {list} list 项目列表
         */
        forceUpdateImageLibraryList (state, list) {
            state.imageLibrary.dataList.splice(0, state.imageLibrary.dataList.length, ...list)
        },

        /**
         * 更新 store 中的 projectImage.dataList
         *
         * @param {Object} state store state
         * @param {list} list 项目列表
         */
        forceUpdateProjectImageList (state, list) {
            state.projectImage.dataList.splice(0, state.projectImage.dataList.length, ...list)
        },

        /**
         * 更新 store 中的 myCollect.dataList
         *
         * @param {Object} state store state
         * @param {list} list 项目列表
         */
        forceUpdateMyCollectList (state, list) {
            state.myCollect.dataList.splice(0, state.myCollect.dataList.length, ...list)
        },

        /**
         * 更新 store 中的 当前镜像信息
         *
         * @param {Object} state store state
         * @param {Object} data 当前镜像信息
         */
        forceUpdateCurImage (state, data) {
            state.curImage = data
        }
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
        getImageLibrary (context, params, config = {}) {
            // return http.get(`/app/depot?invoke=getImageLibrary&${json2Query(params)}`, params, config).then(
            return http.get(`${DEVOPS_BCS_API_URL}/api/depot/images/public/?${json2Query(params)}`, params, config).then(
                response => {
                    const res = response.data
                    context.commit('forceUpdateImageLibraryList', res.results || [])
                    return res
                }
            )
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
        getProjectImage (context, params, config = {}) {
            const projId = params.projId
            delete params.projId
            return http.get(
                `${DEVOPS_BCS_API_URL}/api/depot/images/project/${projId}/?${json2Query(params)}`,
                params,
                config
            ).then(response => {
                const res = response.data
                context.commit('forceUpdateProjectImageList', res.results || [])
                return res
            })
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
        getImageLibraryDetail (context, params, config = {}) {
            const projectId = params.projectId
            delete params.projectId
            return http.get(
                `${DEVOPS_BCS_API_URL}/api/depot/images/project/${projectId}/info/image/?${json2Query(params)}`,
                params,
                config
            )
        }
    }
}
