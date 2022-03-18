<template>
    <div class="biz-content">
        <div class="biz-top-bar">
            <div class="biz-app-title">
                {{$t('HPA管理')}}
            </div>
            <bk-guide></bk-guide>
        </div>
        <div class="biz-content-wrapper" style="padding: 0;" v-bkloading="{ isLoading: isInitLoading, opacity: 0.1 }">
            <app-exception
                v-if="exceptionCode && !isInitLoading"
                :type="exceptionCode.code"
                :text="exceptionCode.msg">
            </app-exception>

            <template v-if="!exceptionCode && !isInitLoading">
                <div class="biz-panel-header">
                    <div class="left">
                        <bk-button @click.stop.prevent="removeHPAs">
                            <span>{{$t('批量删除')}}</span>
                        </bk-button>
                    </div>
                    <div class="right">
                        <bk-data-searcher
                            :scope-list="searchScopeList"
                            :search-key.sync="searchKeyword"
                            :search-scope.sync="searchScope"
                            :cluster-fixed="!!curClusterId"
                            @search="searchHPA"
                            @refresh="refresh">
                        </bk-data-searcher>
                    </div>
                </div>
                <div class="biz-hpa biz-table-wrapper">
                    <bk-table
                        v-bkloading="{ isLoading: isPageLoading && !isInitLoading }"
                        :data="curPageData"
                        :page-params="pageConf"
                        @page-change="pageChangeHandler"
                        @page-limit-change="changePageSize"
                        @select="handlePageSelect"
                        @select-all="handlePageSelectAll">
                        <bk-table-column type="selection" width="60" :selectable="rowSelectable" />
                        <bk-table-column :label="$t('名称')" prop="name" min-width="100">
                            <template slot-scope="{ row }">
                                {{row.name}}
                            </template>
                        </bk-table-column>
                        <bk-table-column :label="$t('集群')" prop="cluster_name" min-width="100">
                            <template slot-scope="{ row }">
                                {{row.cluster_name}}
                            </template>
                        </bk-table-column>
                        <bk-table-column :label="$t('命名空间')" prop="namespace" :show-overflow-tooltip="true" min-width="150" />
                        <bk-table-column :label="$t('Metric(当前/目标)')" prop="current_metrics_display" :show-overflow-tooltip="true" min-width="150">
                        </bk-table-column>
                        <bk-table-column width="30">
                            <template slot-scope="{ row }">
                                <i class="bcs-icon bcs-icon-info-circle" style="color: #ffb400;" v-if="row.needShowConditions" @click="showConditions(row, index)"></i>
                            </template>
                        </bk-table-column>
                        <bk-table-column :label="$t('实例数(当前/范围)')" prop="replicas" min-width="150">
                            <template slot-scope="{ row }">
                                {{ row.current_replicas }} / {{ row.min_replicas }}-{{ row.max_replicas }}
                            </template>
                        </bk-table-column>
                        <bk-table-column :label="$t('关联资源')" :show-overflow-tooltip="true" prop="deployment" min-width="150">
                            <template slot-scope="{ row }">
                                <a class="bk-text-button biz-text-wrapper" target="_blank"
                                    @click="handleGotoAppDetail(row)">{{row.deployment_name}}</a>
                            </template>
                        </bk-table-column>
                        <bk-table-column :label="$t('来源')" prop="source_type">
                            <template slot-scope="{ row }">
                                {{ row.source_type || '--' }}
                            </template>
                        </bk-table-column>
                        <bk-table-column :label="$t('创建时间')" prop="create_time" min-width="100">
                            <template slot-scope="{ row }">
                                {{ row.create_time || '--' }}
                            </template>
                        </bk-table-column>
                        <bk-table-column :label="$t('创建人')" prop="creator" min-width="100">
                            <template slot-scope="{ row }">
                                {{row.creator || '--'}}
                            </template>
                        </bk-table-column>
                        <bk-table-column :label="$t('操作')" prop="permissions">
                            <template slot-scope="{ row }">
                                <div>
                                    <a href="javascript:void(0);" :class="['bk-text-button']" @click="removeHPA(row)">{{$t('删除')}}</a>
                                </div>
                            </template>
                        </bk-table-column>
                    </bk-table>
                </div>
            </template>

            <bk-dialog
                :is-show="batchDialogConfig.isShow"
                :width="430"
                :has-header="false"
                :quick-close="false"
                :title="$t('确认删除')"
                @confirm="deleteHPAs(batchDialogConfig.data)"
                @cancel="batchDialogConfig.isShow = false">
                <template slot="content">
                    <div class="biz-batch-wrapper">
                        <p class="batch-title">{{$t('确定要删除以下')}} HPA？</p>
                        <ul class="batch-list">
                            <li v-for="(item, index) of batchDialogConfig.list" :key="index">{{item}}</li>
                        </ul>
                    </div>
                </template>
            </bk-dialog>

            <conditions-dialog
                :is-show="isShowConditions"
                :item="rowItem"
                @hide-update="hideConditionsDialog">
            </conditions-dialog>
        </div>
    </div>
</template>

<script>
    import { catchErrorHandler } from '@/common/util'

    import ConditionsDialog from './conditions-dialog'

    export default {
        components: {
            ConditionsDialog
        },
        data () {
            return {
                exceptionCode: null,
                isInitLoading: true,
                isPageLoading: false,
                curPageData: [],
                searchKeyword: '',
                searchScope: '',
                pageConf: {
                    total: 1,
                    totalPage: 1,
                    pageSize: 10,
                    curPage: 1,
                    show: true
                },
                alreadySelectedNums: 0,
                isBatchRemoving: false,
                curSelectedData: [],
                batchDialogConfig: {
                    isShow: false,
                    list: [],
                    data: []
                },
                isShowConditions: false,
                rowItem: null,
                hpaSelectedList: []
            }
        },
        computed: {
            isEn () {
                return this.$store.state.isEn
            },
            projectId () {
                return this.$route.params.projectId
            },
            HPAList () {
                const list = this.$store.state.hpa.HPAList
                return JSON.parse(JSON.stringify(list))
            },
            searchScopeList () {
                const clusterList = this.$store.state.cluster.clusterList
                const results = clusterList.map(item => {
                    return {
                        id: item.cluster_id,
                        name: item.name
                    }
                })

                return results
            },
            curClusterId () {
                return this.$store.state.curClusterId
            }
        },
        watch: {
            searchScope () {
                this.init()
            },
            curClusterId () {
                this.searchScope = this.curClusterId
                this.searchHPA()
            }
        },
        created () {
            if (this.searchScopeList.length) {
                const clusterIds = this.searchScopeList.map(item => item.id)
                // 使用当前缓存
                if (this.curClusterId) {
                    this.searchScope = this.curClusterId
                } else if (sessionStorage['bcs-cluster'] && clusterIds.includes(sessionStorage['bcs-cluster'])) {
                    this.searchScope = sessionStorage['bcs-cluster']
                } else {
                    this.searchScope = this.searchScopeList[0].id
                }
            }
        },
        // mounted () {
        //     this.init()
        // },
        methods: {
            /**
             * 初始化入口
             */
            init () {
                this.initPageConf()
                this.getHPAList()
            },

            /**
             * 获取HPA 列表
             */
            async getHPAList () {
                try {
                    await this.$store.dispatch('hpa/getHPAList', {
                        projectId: this.projectId,
                        clusterId: this.searchScope
                    })

                    this.initPageConf()
                    this.curPageData = this.getDataByPage(this.pageConf.curPage)

                    // 如果有搜索关键字，继续显示过滤后的结果
                    if (this.searchScope || this.searchKeyword) {
                        this.searchHPA()
                    }
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    // 晚消失是为了防止整个页面loading和表格数据loading效果叠加产生闪动
                    this.isInitLoading = false
                    this.isPageLoading = false
                }
            },

            /**
             * 每行的多选框点击事件
             */
            rowClick () {
                this.$nextTick(() => {
                    this.alreadySelectedNums = this.HPAList.filter(item => item.isChecked).length
                })
            },

            /**
             * 选择当前页数据
             */
            selectHPAs () {
                const list = this.curPageData
                const selectList = list.filter((item) => {
                    return item.isChecked === true
                })
                this.curSelectedData.splice(0, this.curSelectedData.length, ...selectList)
            },

            /**
             * 刷新列表
             */
            refresh () {
                this.pageConf.curPage = 1
                this.isPageLoading = true
                this.getHPAList()
            },

            /**
             * 分页大小更改
             *
             * @param {number} pageSize pageSize
             */
            changePageSize (pageSize) {
                this.pageConf.pageSize = pageSize
                this.pageConf.curPage = 1
                this.initPageConf()
                this.pageChangeHandler()
            },

            /**
             * 单选
             * @param {array} selection 已经选中的行数
             * @param {object} row 当前选中的行
             */
            handlePageSelect (selection, row) {
                this.hpaSelectedList = selection
            },

            /**
             * 全选
             */
            handlePageSelectAll (selection, row) {
                this.hpaSelectedList = selection
            },

            /**
             * 搜索HPA
             */
            searchHPA () {
                const keyword = this.searchKeyword.trim()
                const keyList = ['name', 'cluster_name', 'creator']
                let list = JSON.parse(JSON.stringify(this.$store.state.hpa.HPAList))
                const results = []

                if (this.searchScope) {
                    list = list.filter(item => {
                        return item.cluster_id === this.searchScope
                    })
                }

                list.forEach(item => {
                    item.isChecked = false
                    for (const key of keyList) {
                        if (item[key]?.indexOf(keyword) > -1) {
                            results.push(item)
                            return true
                        }
                    }
                })

                this.HPAList.splice(0, this.HPAList.length, ...results)
                this.pageConf.curPage = 1
                this.initPageConf()
                this.curPageData = this.getDataByPage(this.pageConf.curPage)
            },

            /**
             * 初始化分页配置
             */
            initPageConf () {
                const total = this.HPAList.length
                this.pageConf.total = total
                this.pageConf.totalPage = Math.ceil(total / this.pageConf.pageSize)
                if (this.pageConf.curPage > this.pageConf.totalPage) {
                    this.pageConf.curPage = this.pageConf.totalPage
                }
            },

            /**
             * 获取分页数据
             * @param  {number} page 第几页
             * @return {object} data 数据
             */
            getDataByPage (page) {
                if (page < 1) {
                    this.pageConf.curPage = page = 1
                }
                let startIndex = (page - 1) * this.pageConf.pageSize
                let endIndex = page * this.pageConf.pageSize
                this.isPageLoading = true
                if (startIndex < 0) {
                    startIndex = 0
                }
                if (endIndex > this.HPAList.length) {
                    endIndex = this.HPAList.length
                }
                setTimeout(() => {
                    this.isPageLoading = false
                }, 200)
                this.hpaSelectedList = []
                return this.HPAList.slice(startIndex, endIndex)
            },

            /**
             * 页数改变回调
             * @param  {number} page 第几页
             */
            pageChangeHandler (page = 1) {
                this.pageConf.curPage = page

                const data = this.getDataByPage(page)
                this.curPageData = data
            },

            /**
             * 重新加载当面页数据
             */
            reloadCurPage () {
                this.initPageConf()
                if (this.pageConf.curPage > this.pageConf.totalPage) {
                    this.pageConf.curPage = this.pageConf.totalPage
                }
                this.curPageData = this.getDataByPage(this.pageConf.curPage)
            },

            /**
             * 清空当前页选择
             */
            clearSelectHPAs () {
                this.HPAList.forEach((item) => {
                    item.isChecked = false
                })
            },

            /**
             * 确认批量删除HPA
             */
            async removeHPAs () {
                const data = []
                const names = []

                if (this.hpaSelectedList.length) {
                    this.hpaSelectedList.forEach(item => {
                        data.push({
                            cluster_id: item.cluster_id,
                            namespace: item.namespace,
                            name: item.name
                        })
                        names.push(item.name)
                    })
                }
                if (!data.length) {
                    this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请选择要删除的HPA！')
                    })
                    return false
                }

                this.batchDialogConfig.list = names
                this.batchDialogConfig.data = data
                this.batchDialogConfig.isShow = true
            },

            /**
             * 确认删除HPA
             * @param  {object} HPA HPA
             */
            async removeHPA (HPA) {
                const self = this

                this.$bkInfo({
                    title: this.$t('确认删除'),
                    clsName: 'biz-remove-dialog',
                    content: this.$createElement('p', {
                        class: 'biz-confirm-desc'
                    }, `${this.$t('确定要删除HPA')}【${HPA.name}】？`),
                    async confirmFn () {
                        self.deleteHPA(HPA)
                    }
                })
            },

            /**
             * 批量删除HPA
             * @param  {object} data HPAs
             */
            async deleteHPAs (data) {
                this.batchDialogConfig.isShow = false
                this.isPageLoading = true
                const projectId = this.projectId

                try {
                    await this.$store.dispatch('hpa/batchDeleteHPA', {
                        projectId,
                        params: { data }
                    })

                    this.$bkMessage({
                        theme: 'success',
                        message: this.$t('删除成功')
                    })
                    this.initPageConf()
                    this.getHPAList()
                } catch (e) {
                    // 4004，已经被删除过，但接口不能立即清除，防止重复删除
                    if (e.code === 4004) {
                        this.initPageConf()
                        this.getHPAList()
                    }
                    this.$bkMessage({
                        theme: 'error',
                        delay: 8000,
                        hasCloseIcon: true,
                        message: e.message
                    })
                    this.isPageLoading = false
                }
            },

            /**
             * 删除HPA
             * @param {object} HPA HPA
             */
            async deleteHPA (HPA) {
                const projectId = this.projectId
                const namespace = HPA.namespace
                const clusterId = HPA.cluster_id
                const name = HPA.name
                this.isPageLoading = true
                try {
                    await this.$store.dispatch('hpa/deleteHPA', {
                        projectId,
                        clusterId,
                        namespace,
                        name
                    })

                    this.$bkMessage({
                        theme: 'success',
                        message: this.$t('删除成功')
                    })
                    this.initPageConf()
                    this.getHPAList()
                } catch (e) {
                    catchErrorHandler(e, this)
                    this.isPageLoading = false
                }
            },

            /**
             * 显示 conditions 弹框
             *
             * @param {Object} item 当前行对象
             * @param {number} index 当前行对象索引
             *
             * @return {string} returnDesc
             */
            showConditions (item, index) {
                this.isShowConditions = true
                this.rowItem = item
            },

            /**
             * 关闭 conditions 弹框
             */
            hideConditionsDialog () {
                this.isShowConditions = false
                setTimeout(() => {
                    this.rowItem = null
                }, 300)
            },

            rowSelectable (row, index) {
                return row.can_delete
            },
            handleGotoAppDetail (row) {
                const kindMap = {
                    deployment: 'deploymentsInstanceDetail2',
                    daemonset: 'daemonsetInstanceDetail2',
                    job: 'jobInstanceDetail2'
                }
                const location = this.$router.resolve({
                    name: kindMap[row.resource_kind] || '404',
                    params: {
                        instanceName: row.deployment_name,
                        instanceNamespace: row.namespace,
                        instanceCategory: row.resource_kind
                    },
                    query: {
                        cluster_id: row.cluster_id
                    }
                })
                window.open(location.href)
            }
        }
    }
</script>

<style scoped>
    @import './index.css';
</style>
