<template>
    <div class="biz-content">
        <div class="biz-top-bar">
            <div class="biz-topbar-title">
                PersistentVolume
            </div>
            <bk-guide></bk-guide>
        </div>
        <div class="biz-content-wrapper p0" v-bkloading="{ isLoading: isInitLoading, opacity: 0.1 }">
            <app-exception
                v-if="exceptionCode && !isInitLoading"
                :type="exceptionCode.code"
                :text="exceptionCode.msg">
            </app-exception>

            <template v-if="!exceptionCode && !isInitLoading">
                <div class="biz-panel-header">
                    <div class="right">
                        <searcher
                            :placeholder="$t('输入名称，按Enter搜索')"
                            :scope-list="clusterList"
                            :search-scope.sync="searchClusterId"
                            :search-key.sync="searchKeyword"
                            :cluster-fixed="!!curClusterId"
                            @update:searchScope="fetchData"
                            @update:searchKey="searchStorageByWord"
                            @refresh="refresh">
                        </searcher>
                    </div>
                </div>

                <div class="biz-pv">
                    <div class="biz-table-wrapper">
                        <bk-table
                            v-bkloading="{ isLoading: isPageLoading && !isInitLoading }"
                            class="biz-pv-table"
                            :data="curPageData"
                            :page-params="pageConf"
                            @page-change="pageChange"
                            @page-limit-change="changePageSize">
                            <bk-table-column :label="$t('名称')" prop="name">
                                <template slot-scope="{ row }">
                                    <bcs-popover placement="top" :delay="500">
                                        <p class="item-name">{{row.name || '--'}}</p>
                                        <template slot="content">
                                            <p style="text-align: left; white-space: normal;word-break: break-all;">{{row.name || '--'}}</p>
                                        </template>
                                    </bcs-popover>
                                </template>
                            </bk-table-column>
                            <bk-table-column :label="$t('状态')" prop="status">
                                <template slot-scope="{ row }">
                                    {{ row.status || '--' }}
                                </template>
                            </bk-table-column>
                            <bk-table-column :label="$t('访问权限')" prop="access_modes">
                                <template slot-scope="{ row }">
                                    <bcs-popover placement="top" :delay="500">
                                        <p class="item-name">{{ row.access_modes || '--' }}</p>
                                        <template slot="content">
                                            <p style="text-align: left; white-space: normal;word-break: break-all;">{{ row.access_modes || '--' }}</p>
                                        </template>
                                    </bcs-popover>
                                </template>
                            </bk-table-column>
                            <bk-table-column :label="$t('回收策略')" prop="reclaim_policy">
                                <template slot-scope="{ row }">
                                    {{ row.reclaim_policy || '--' }}
                                </template>
                            </bk-table-column>
                            <bk-table-column label="PVC" prop="pvc_name">
                                <template slot-scope="{ row }">
                                    {{ row.pvc_name || '--' }}
                                </template>
                            </bk-table-column>
                            <bk-table-column label="StorageClass" prop="sc_name">
                                <template slot-scope="{ row }">
                                    {{ row.sc_name || '--' }}
                                </template>
                            </bk-table-column>
                        </bk-table>
                    </div>
                </div>
            </template>
        </div>
    </div>
</template>

<script>
    import Searcher from './searcher'

    export default {
        components: {
            Searcher
        },
        data () {
            return {
                isInitLoading: true,
                isPageLoading: false,
                exceptionCode: null,
                searchKeyword: '',
                searchClusterId: '',
                curPageData: [],
                pageConf: {
                    total: 1,
                    totalPage: 1,
                    pageSize: 10,
                    curPage: 1,
                    show: true
                },
                dataListTmp: [],
                dataList: []
            }
        },
        computed: {
            isEn () {
                return this.$store.state.isEn
            },
            projectId () {
                return this.$route.params.projectId
            },
            curClusterId () {
                return this.$store.state.curClusterId
            },
            clusterList () {
                return this.$store.state.cluster.clusterList.map(item => {
                    return {
                        id: item.cluster_id,
                        cluster_id: item.cluster_id,
                        cluster_name: item.name,
                        name: item.name
                    }
                })
            }
        },
        watch: {
            curClusterId () {
                this.searchClusterId = this.curClusterId
                this.fetchData()
            }
        },
        async created () {
            await this.getClusters()
        },
        methods: {
            /**
             * 获取所有的集群
             */
            async getClusters () {
                try {
                    if (this.clusterList.length) {
                        const clusterIds = this.clusterList.map(item => item.id)
                        // 使用当前缓存
                        if (sessionStorage['bcs-cluster'] && clusterIds.includes(sessionStorage['bcs-cluster'])) {
                            this.searchClusterId = sessionStorage['bcs-cluster']
                        } else {
                            this.searchClusterId = this.clusterList[0].cluster_id
                        }

                        await this.fetchData()
                    } else {
                        // 没有集群时，这里就终止了，不会执行 fetchData，所以这里关闭 loading，不能在 finally 里面关闭
                        // 因为如果集群存在时，还需要 loading fetchData
                        setTimeout(() => {
                            this.isInitLoading = false
                        }, 200)
                    }
                } catch (e) {
                    console.error(e)
                    setTimeout(() => {
                        this.isInitLoading = false
                    }, 200)
                }
            },

            /**
             * 加载configmap列表数据
             */
            async fetchData () {
                if (!this.searchClusterId) {
                    return false
                }
                try {
                    this.isPageLoading = true
                    const res = await this.$store.dispatch('storage/getList', {
                        projectId: this.projectId,
                        clusterId: this.searchClusterId,
                        idx: 'pv'
                    })
                    /* const res = {
                        'data': [
                            {
                                'name': 'pvc-70a18011-5172-11ea-ac5a-525400e9e7cd',
                                'status': 'Bound',
                                'access_modes': 'ReadWriteOnce',
                                'pvc_name': 'bellkepvctest',
                                'sc_name': 'bellketest',
                                'reclaim_policy': 'Delete',
                                'create_time': '2020-02-17 18:44:17'
                            },
                            {
                                'name': 'test3',
                                'status': 'Released',
                                'access_modes': 'ReadWriteOnce',
                                'pvc_name': 'test321',
                                'sc_name': 'cbs',
                                'reclaim_policy': 'Retain',
                                'create_time': '2020-02-16 19:08:21'
                            },
                            {
                                'name': 'test4',
                                'status': 'Released',
                                'access_modes': 'ReadWriteOnce',
                                'pvc_name': 'bellkepvctest',
                                'sc_name': 'bellketest',
                                'reclaim_policy': 'Retain',
                                'create_time': '2020-02-17 14:50:09'
                            }
                        ],
                        'code': 0,
                        'message': 'OK',
                        'request_id': '4c6e6be3c0f7eed941f8758e49962b03'
                    } */
                    const list = res.data || []
                    this.dataListTmp.splice(0, this.dataListTmp.length, ...list)
                    this.dataList.splice(0, this.dataList.length, ...list)
                    this.pageConf.curPage = 1
                    this.searchStorageByWord()
                } catch (e) {
                    console.error(e)
                } finally {
                    // 晚消失是为了防止整个页面loading和表格数据loading效果叠加产生闪动
                    setTimeout(() => {
                        this.isInitLoading = false
                        this.isPageLoading = false
                    }, 200)
                }
            },

            /**
             * 根据关键字搜索
             */
            searchStorageByWord () {
                const search = String(this.searchKeyword || '').trim().toLowerCase()
                let results = []
                if (search === '') {
                    this.dataList.splice(0, this.dataList.length, ...this.dataListTmp)
                } else {
                    results = this.dataListTmp.filter(m => {
                        return m.name.toLowerCase().indexOf(search) > -1
                    })
                    this.dataList.splice(0, this.dataList.length, ...results)
                }
                this.initPageConf()
                this.curPageData = this.getDataByPage()
            },

            /**
             * 初始化翻页条
             */
            initPageConf () {
                const total = this.dataList.length
                if (total <= this.pageConf.pageSize) {
                    this.pageConf.show = false
                } else {
                    this.pageConf.show = true
                }
                this.pageConf.total = total
                this.pageConf.totalPage = Math.ceil(total / this.pageConf.pageSize) || 1
            },

            /**
             * 翻页回调
             *
             * @param {number} page 当前页
             */
            pageChange (page = 1) {
                this.pageConf.curPage = page
                const data = this.getDataByPage(page)
                this.curPageData.splice(0, this.curPageData.length, ...data)
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
                this.pageChange()
            },

            /**
             * 获取当前这一页的数据
             *
             * @param {number} page 当前页
             *
             * @return {Array} 当前页数据
             */
            getDataByPage (page) {
                // 如果没有page，重置
                if (!page) {
                    this.pageConf.curPage = page = 1
                }
                let startIndex = (page - 1) * this.pageConf.pageSize
                let endIndex = page * this.pageConf.pageSize
                if (startIndex < 0) {
                    startIndex = 0
                }
                if (endIndex > this.dataList.length) {
                    endIndex = this.dataList.length
                }
                return this.dataList.slice(startIndex, endIndex)
            },

            /**
             * 手动刷新表格数据
             */
            async refresh () {
                this.pageConf.curPage = 1
                this.searchKeyword = ''
                if (this.curClusterId) {
                    this.searchClusterId = this.curClusterId
                } else {
                    this.searchClusterId = this.clusterList[0].cluster_id
                }
                await this.fetchData()
            },

            /**
             * 重新加载当面页数据
             * @return {[type]} [description]
             */
            reloadCurPage () {
                this.initPageConf()
                this.curPageData = this.getDataByPage(this.pageConf.curPage)
            }
        }
    }
</script>
<style scoped>
    @import './pv.css';
</style>
