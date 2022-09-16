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

export default {
    computed: {
        onlineProjectList () {
            return this.$store.state.sideMenu.onlineProjectList || []
        },
        curProjectId () {
            return this.$store.state.curProjectId
        }
    },
    methods: {
        /**
         * 初始化当前项目的数据
         *
         * @param {Function} callback 成功后的回调处理
         */
        initCurProject (callback) {
            if (this.onlineProjectList.length) {
                this.projectId = this.$route.params.projectId
                    || this.curProjectId
                    || this.onlineProjectList[0].project_id

                const curProject = this.onlineProjectList.find(project => {
                    return project.project_id === this.projectId
                })

                return Object.assign({}, curProject)
            }

            return {}
        }
    }
}
