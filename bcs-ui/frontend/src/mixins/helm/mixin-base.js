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

import bkKeyer from '@open/components/keyer'
import ace from '@open/components/ace-editor'
import bkFormCreater from '@open/components/form-creater'

export default {
    components: {
        bkKeyer,
        ace,
        bkFormCreater
    },
    data () {
        return {
            curProjectId: ''
        }
    },
    computed: {
        curProject () {
            const project = this.$store.state.curProject
            return project
        },
        projectId () {
            this.curProjectId = this.$route.params.projectId
            return this.curProjectId
        }
    },
    watch: {
        // 如果不是k8s类型的项目，无法访问些页面，重定向回集群首页
        curProjectId () {
            if (this.curProject && this.curProject.kind !== PROJECT_K8S) {
                this.$router.push({
                    name: 'clusterMain',
                    params: {
                        projectId: this.projectId,
                        projectCode: this.projectCode
                    }
                })
            }
        }
    },
    methods: {
        /**
         * 简单判断是否为图片
         * @param  {string}  img 图片url
         * @return {boolean} true/false
         */
        isImage (img) {
            if (!img) {
                return false
            }
            if (img.startsWith('http://') || img.startsWith('https://') || img.startsWith('data:image/')) {
                return true
            }
            return false
        },

        /**
         * 根据path（eg: a.b.c）设置对象属性
         * @param {object} obj 对象
         * @param {string} path  路径
         * @param {string number...} value 值
         */
        setProperty (obj, path, value) {
            const paths = path.split('.')
            let temp = obj
            const pathLength = paths.length
            if (pathLength) {
                for (let i = 0; i < pathLength; i++) {
                    const item = paths[i]
                    if (temp.hasOwnProperty(item)) {
                        if (i === (pathLength - 1)) {
                            temp[item] = value
                        } else {
                            temp = temp[item]
                        }
                    }
                }
            }
        },

        /**
         * 根据path（eg: a.b.c）判断对象属性是否存在
         * @param {object} obj 对象
         * @param {string} path  路径
         * @return {boolean} true/false
         */
        hasProperty (obj, path) {
            const paths = path.split('.')
            let temp = obj
            const pathLength = paths.length
            if (pathLength) {
                for (let i = 0; i < pathLength; i++) {
                    const item = paths[i]
                    if (temp.hasOwnProperty(item)) {
                        temp = temp[item]
                    } else {
                        return false
                    }
                }
                return true
            }
            return false
        },

        /**
         * 根据path（eg: a.b.c）获取对象属性
         * @param {object} obj 对象
         * @param {string} path  路径
         * @return {string number...} value
         */
        getProperty (obj, path) {
            const paths = path.split('.')
            let temp = obj
            if (paths.length) {
                for (const item of paths) {
                    if (temp.hasOwnProperty(item)) {
                        temp = temp[item]
                    } else {
                        return undefined
                    }
                }
                return temp
            }
            return undefined
        },

        /**
         * 返回
         */
        goBack () {
            window.history.go(-1)
        }
    }
}
