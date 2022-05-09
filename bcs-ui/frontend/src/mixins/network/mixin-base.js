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
    data () {
        return {
            isBatchRemoving: false,
            curSelectedData: [],
            batchDialogConfig: {
                isShow: false,
                list: [],
                data: []
            }
        }
    },
    computed: {
        isCheckCurPageAll () {
            if (this.curPageData.length) {
                const list = this.curPageData
                const selectList = list.filter((item) => {
                    return item.isChecked === true
                })
                const canSelectList = list.filter((item) => {
                    return item.can_delete
                })
                if (selectList.length && (selectList.length === canSelectList.length)) {
                    return true
                } else {
                    return false
                }
            } else {
                return false
            }
        }
    },
    methods: {

        /**
         * 每行的多选框点击事件
         */
        rowClick () {
            this.$nextTick(() => {
                this.alreadySelectedNums = this.serviceList.filter(item => item.isChecked).length
            })
        },

        /**
         * 选择当前页数据
         */
        selectServices () {
            const list = this.curPageData
            const selectList = list.filter((item) => {
                return item.isChecked === true
            })
            this.curSelectedData.splice(0, this.curSelectedData.length, ...selectList)
        },

        /**
         * 清空当前页选择
         */
        clearSelectServices () {
            this.serviceList.forEach((item) => {
                item.isChecked = false
            })
        },

        /**
         * 确认批量删除service
         */
        async removeServices () {
            const data = []
            const names = []

            this.serviceSelectedList.forEach(item => {
                data.push({
                    cluster_id: item.clusterId,
                    namespace: item.namespace,
                    name: item.resourceName
                })
                names.push(`${item.cluster_id} / ${item.namespace} / ${item.resourceName}`)
            })
            if (!data.length) {
                this.$bkMessage({
                    theme: 'error',
                    message: this.$t('请选择要删除的Service')
                })
                return false
            }

            this.batchDialogConfig.list = names
            this.batchDialogConfig.data = data
            this.batchDialogConfig.isShow = true
        },

        /**
         * 批量删除service
         * @param  {object} data services
         */
        async deleteServices (data) {
            this.batchDialogConfig.isShow = false
            this.isPageLoading = true
            const projectId = this.projectId

            try {
                await this.$store.dispatch('network/deleteServices', {
                    projectId,
                    data
                })

                this.$bkMessage({
                    theme: 'success',
                    message: this.$t('删除成功')
                })
                this.initPageConf()
                this.getServiceList()
            } catch (e) {
                // 4004，已经被删除过，但接口不能立即清除，防止重复删除
                if (e.code === 4004) {
                    this.initPageConf()
                    this.getServiceList()
                }
                this.$bkMessage({
                    theme: 'error',
                    delay: 8000,
                    hasCloseIcon: true,
                    message: e.message
                })
                this.isPageLoading = false
            }
        }
    }
}
