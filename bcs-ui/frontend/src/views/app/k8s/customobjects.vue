<template>
    <div class="biz-content">
        <div class="biz-top-bar">
            <div class="biz-app-title">
                CustomObjects
            </div>
            <bk-guide></bk-guide>
        </div>
        <div class="biz-content-wrapper" style="padding: 0;" v-bkloading="{ isLoading: isInitLoading, opacity: 0.1 }">
            <template v-if="!isInitLoading">
                <div class="biz-panel-header biz-event-query-query" style="padding-right: 0;">
                    <div class="left">
                        <bk-selector
                            :placeholder="$t('请选择集群')"
                            :searchable="true"
                            :setting-key="'cluster_id'"
                            :display-key="'name'"
                            :selected.sync="selectedClusterId"
                            :disabled="!!curClusterId"
                            :list="clusterList"
                            :search-placeholder="$t('输入集群名称搜索')"
                            @item-selected="handleChangeCluster">
                        </bk-selector>
                    </div>
                    <div class="left" style="width: 290px;">
                        <bk-selector
                            :placeholder="$t('请选择CRD')"
                            :searchable="true"
                            :setting-key="'name'"
                            :display-key="'name'"
                            :selected.sync="selectedCRD"
                            :list="crdList"
                            :is-loading="crdLoading"
                            :search-placeholder="$t('输入CRD搜索')"
                            @item-selected="handleChangeCRD">
                        </bk-selector>
                    </div>
                    <div class="left" v-if="isNamespaceScope">
                        <bk-selector
                            :placeholder="$t('请选择命名空间')"
                            :searchable="true"
                            :allow-clear="true"
                            :setting-key="'name'"
                            :display-key="'name'"
                            :selected.sync="selectedNamespaceName"
                            :list="namespaceList"
                            :is-loading="namespaceLoading"
                            :search-placeholder="$t('输入命名空间搜索')"
                            @item-selected="handleChangeNamespace"
                            @clear="handleClearNamespace">
                        </bk-selector>
                    </div>
                    <div class="left">
                        <bk-input v-model="searchKey" clearable :placeholder="$t('输入名称搜索')" @clear="clearSearch" />
                    </div>
                    <div class="left" style="width: auto;">
                        <bk-button type="primary" :title="$t('查询')" icon="search" @click="handleClick">
                            {{$t('查询')}}
                        </bk-button>
                    </div>
                </div>
                <div v-bkloading="{ isLoading: isPageLoading && !isInitLoading }">
                    <div class="biz-table-wrapper gamestatefullset-table-wrapper">
                        <bk-table
                            :data="curPageData"
                            :page-params="pageConf"
                            @page-change="handlePageChange"
                            @page-limit-change="handlePageSizeChange">
                            <bk-table-column v-for="(column, index) in columnList" :label="(defaultColumnMap[column] && defaultColumnMap[column].label) || column"
                                :min-width="defaultColumnMap[column] ? defaultColumnMap[column].minWidth : 'auto'"
                                :key="index">
                                <template slot-scope="{ row }">
                                    <div>
                                        <template v-if="column === 'name'">
                                            <a href="javascript:void(0);" class="bk-text-button name-col bcs-ellipsis" style="font-weight: 700;"
                                                @click="showSideslider(row[column], row['namespace'])">{{row[column] || '--'}}</a>
                                        </template>
                                        <template v-else>
                                            {{row[column] || '--'}}
                                        </template>
                                    </div>
                                </template>
                            </bk-table-column>
                            <bk-table-column :label="$t('操作')" width="150">
                                <template slot-scope="{ row }">
                                    <a href="javascript:void(0);" class="bk-text-button" @click.stop="del(row, index)">{{$t('删除')}}</a>
                                </template>
                            </bk-table-column>
                        </bk-table>
                    </div>
                </div>
            </template>
        </div>
        <gamestatefulset-sideslider
            :is-show="isShowSideslider"
            :cluster-id="selectedClusterId"
            :namespace-name="curShowNamespace"
            :name="curShowName"
            :crd="selectedCRD"
            @hide-sideslider="hideSideslider">
        </gamestatefulset-sideslider>
    </div>
</template>

<script>
    import { catchErrorHandler } from '@/common/util'

    import GamestatefulsetSideslider from './gamestatefulset-sideslider'

    export default {
        components: {
            GamestatefulsetSideslider
        },
        data () {
            return {
                CATEGORY: 'gamestatefulset',

                isInitLoading: true,
                isPageLoading: false,

                pageConf: {
                    total: 0,
                    totalPage: 1,
                    pageSize: 10,
                    curPage: 1,
                    show: true
                },
                defaultColumnMap: {
                    'name': {
                        label: this.$t('名称'),
                        minWidth: 150
                    },
                    'cluster_id': {
                        label: this.$t('集群'),
                        minWidth: 140
                    },
                    'namespace': {
                        label: this.$t('命名空间'),
                        minWidth: 100
                    }
                },
                bkMessageInstance: null,
                selectedClusterId: '',
                selectedNamespaceName: '',
                namespaceList: [],
                namespaceLoading: false,
                selectedCRD: '',
                crdList: [],
                crdLoading: false,
                isNamespaceScope: false,
                columnList: ['name', 'cluster_id', 'namespace', 'Age'],
                renderList: [],
                renderListTmp: [],
                curPageData: [],
                isShowSideslider: false,
                curShowName: '',
                curShowNamespace: '',
                searchKey: ''
            }
        },
        computed: {
            projectId () {
                return this.$route.params.projectId
            },
            isEn () {
                return this.$store.state.isEn
            },
            curClusterId () {
                return this.$store.state.curClusterId
            },
            clusterList () {
                return this.$store.state.cluster.clusterList
            }
        },
        watch: {
            curClusterId: {
                handler (v) {
                    this.selectedClusterId = v
                    this.curPageData = []
                },
                immediate: true
            }
        },
        async mounted () {
            await this.getClusters()
        },
        destroyed () {
            this.bkMessageInstance && this.bkMessageInstance.close()
        },
        methods: {
            /**
             * 获取所有的集群
             */
            async getClusters () {
                try {
                    if (this.clusterList.length) {
                        if (!this.curClusterId) {
                            this.selectedClusterId = this.clusterList[0].cluster_id
                        }
                        await this.getCRDList()
                        this.getNameSpaceList()
                    }
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    this.isPageLoading = false
                    // 晚消失是为了防止整个页面loading和表格数据loading效果叠加产生闪动
                    setTimeout(() => {
                        this.isInitLoading = false
                    }, 200)
                }
            },

            /**
             * 获取 crd 列表
             */
            async getCRDList () {
                try {
                    this.crdLoading = true
                    const res = await this.$store.dispatch('app/getCRDList', {
                        projectId: this.projectId,
                        clusterId: this.selectedClusterId
                    })
                    const list = res.data || []
                    this.crdList.splice(0, this.crdList.length, ...list)
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    this.crdLoading = false
                }
            },

            /**
             * 获取命名空间列表
             */
            async getNameSpaceList () {
                try {
                    this.namespaceLoading = true
                    const res = await this.$store.dispatch('crdcontroller/getNameSpaceListByCluster', {
                        projectId: this.projectId,
                        clusterId: this.selectedClusterId
                    })
                    const list = res.data || []
                    this.namespaceList.splice(0, this.namespaceList.length, ...list)
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    this.namespaceLoading = false
                }
            },

            /**
             * 集群下拉框 item-selected 事件
             *
             * @param {string} clusterId 集群 id
             * @param {Object} data 集群对象
             */
            async handleChangeCluster (clusterId, data) {
                this.selectedClusterId = clusterId

                this.selectedNamespaceName = ''
                this.namespaceList.splice(0, this.namespaceList.length, ...[])

                this.selectedCRD = ''
                this.crdList.splice(0, this.crdList.length, ...[])

                await this.getNameSpaceList()
                await this.getCRDList()
            },

            /**
             * 命名空间下拉框 item-selected 事件
             *
             * @param {string} selectedNamespaceName 命名空间 name
             * @param {Object} data 命名空间对象
             */
            handleChangeNamespace (selectedNamespaceName, data) {
                this.selectedNamespaceName = selectedNamespaceName
            },

            /**
             * 命名空间下拉框 clear 事件
             */
            handleClearNamespace () {
                this.selectedNamespaceName = ''
            },

            /**
             * CRD 下拉框 item-selected 事件
             *
             * @param {string} selectedCRD crd name
             * @param {Object} data crd 对象
             */
            handleChangeCRD (selectedCRD, data) {
                this.selectedCRD = selectedCRD
                this.isNamespaceScope = (data.scope || '').toLowerCase() === 'namespaced'
            },

            /**
             * 获取表格数据
             */
            async fetchData () {
                this.isPageLoading = true
                try {
                    if (!this.selectedClusterId) return

                    const params = {}
                    if (this.selectedNamespaceName && this.isNamespaceScope) {
                        params.namespace = this.selectedNamespaceName
                    }
                    const res = await this.$store.dispatch('app/getGameStatefulsetList', {
                        projectId: this.projectId,
                        clusterId: this.selectedClusterId,
                        gamestatefulsets: this.selectedCRD,
                        data: params
                    })

                    const data = res.data || { td_list: [], th_list: [] }

                    if (data.th_list.length) {
                        this.columnList.splice(0, this.columnList.length, ...data.th_list)
                    } else {
                        this.columnList.splice(0, this.columnList.length, ...['name', 'cluster_id', 'namespace', 'Age'])
                    }
                    this.renderListTmp.splice(0, this.renderListTmp.length, ...data.td_list)

                    if (this.searchKey.trim()) {
                        this.renderList.splice(
                            0,
                            this.renderList.length,
                            ...this.renderListTmp.filter(item => item.name.indexOf(this.searchKey) > -1)
                        )
                    } else {
                        this.renderList.splice(0, this.renderList.length, ...this.renderListTmp)
                    }

                    this.pageConf.curPage = 1
                    this.initPageConf()
                    this.curPageData = this.getDataByPage(this.pageConf.curPage)
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    this.isPageLoading = false
                    // 晚消失是为了防止整个页面loading和表格数据loading效果叠加产生闪动
                    setTimeout(() => {
                        this.isInitLoading = false
                    }, 200)
                }
            },

            /**
             * 初始化分页配置
             */
            initPageConf () {
                const total = this.renderList.length
                this.pageConf.total = total
                this.pageConf.totalPage = Math.ceil(total / this.pageConf.pageSize)
            },

            /**
             * 分页大小更改
             *
             * @param {number} pageSize pageSize
             */
            handlePageSizeChange (pageSize) {
                this.pageConf.pageSize = pageSize
                this.pageConf.curPage = 1
                this.initPageConf()
                this.handlePageChange()
            },

            /**
             * 获取分页数据
             * @param  {number} page 第几页
             * @return {object} data 数据
             */
            getDataByPage (page) {
                let startIndex = (page - 1) * this.pageConf.pageSize
                let endIndex = page * this.pageConf.pageSize
                if (startIndex < 0) {
                    startIndex = 0
                }
                if (endIndex > this.renderList.length) {
                    endIndex = this.renderList.length
                }
                return this.renderList.slice(startIndex, endIndex)
            },

            /**
             * 页数改变回调
             * @param  {number} page 第几页
             */
            handlePageChange (page = 1) {
                this.pageConf.curPage = page
                const data = this.getDataByPage(page)
                this.curPageData = data
            },

            /**
             * 搜索框清除事件
             */
            clearSearch () {
                this.searchKey = ''
                this.handleClick()
            },

            /**
             * 搜索按钮点击
             *
             * @param {Object} e 时间对象
             */
            async handleClick (e) {
                if (!this.selectedCRD) {
                    this.bkMessageInstance && this.bkMessageInstance.close()
                    this.bkMessageInstance = this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请选择CRD')
                    })
                    return
                }
                await this.fetchData({
                    projId: this.projectId,
                    limit: this.pageConf.pageSize,
                    offset: 0
                })
            },

            /**
             * 删除当前行
             *
             * @param {Object} item 当前行对象
             * @param {number} index 当前行对象索引
             *
             * @return {string} returnDesc
             */
            async del (item, index) {
                const me = this

                const boxStyle = {
                    'margin-top': '-20px',
                    'margin-bottom': '-20px'
                }
                // const titleStyle = {
                //     style: {
                //         'text-align': 'left',
                //         'font-size': '20px',
                //         'margin-bottom': '15px',
                //         'color': '#313238'
                //     }
                // }
                const itemStyle = {
                    style: {
                        'text-align': 'left',
                        'font-size': '14px',
                        'margin-bottom': '3px',
                        'color': '#71747c'
                    }
                }

                const contexts = [
                    // me.$createElement('h5', titleStyle, me.$t('确定要删除？')),
                    me.$createElement('p', itemStyle, `${me.$t('名称')}：${item.name}`),
                    me.$createElement('p', itemStyle, `${me.$t('所属集群')}：${item.cluster_id}`),
                    me.$createElement('p', itemStyle, `${me.$t('命名空间')}：${item.namespace}`)
                ]
                me.$bkInfo({
                    title: me.$t('确认删除'),
                    confirmLoading: true,
                    clsName: 'biz-remove-dialog',
                    content: me.$createElement('div', { class: 'biz-confirm-desc', style: boxStyle }, contexts),
                    async confirmFn () {
                        try {
                            await me.$store.dispatch('app/deleteGameStatefulsetInfo', {
                                projectId: me.projectId,
                                clusterId: me.selectedClusterId,
                                gamestatefulsets: me.selectedCRD,
                                name: item.name,
                                data: {
                                    namespace: item.namespace
                                }
                            })
                            me.bkMessageInstance && me.bkMessageInstance.close()
                            me.bkMessageInstance = me.$bkMessage({
                                theme: 'success',
                                message: me.$t('删除成功'),
                                delay: 1000
                            })
                            await me.fetchData()
                        } catch (e) {
                            console.error(e)
                        }
                    }
                })
            },

            /**
             * 显示 sideslider
             */
            async showSideslider (name, namespace) {
                this.curShowName = name
                this.curShowNamespace = namespace
                this.isShowSideslider = true
            },

            /**
             * 隐藏 sideslider
             */
            hideSideslider () {
                this.curShowName = ''
                this.curShowNamespace = ''
                this.isShowSideslider = false
            }
        }
    }
</script>

<style scoped>
    @import '../customobjects.css';
</style>
