<template>
    <div class="biz-content">
        <div class="biz-top-bar">
            <div class="biz-event-query-title">
                {{$t('事件查询')}}
            </div>
            <bk-guide></bk-guide>
        </div>
        <div class="biz-content-wrapper" style="padding: 0;" v-bkloading="{ isLoading: isInitLoading, opacity: 0.1 }">
            <template v-if="!isInitLoading">
                <div class="biz-panel-header biz-event-query-query" style="padding-right: 0;">
                    <div class="left">
                        <bk-selector :placeholder="$t('集群')"
                            :selected.sync="clusterIndex"
                            :disabled="!!curClusterId"
                            :list="dropdownClusterList"
                            :setting-key="'cluster_id'"
                            :display-key="'name'"
                            :allow-clear="true"
                            @clear="clusterClear">
                        </bk-selector>
                    </div>
                    <div class="left">
                        <bk-selector :placeholder="$t('事件对象')"
                            :selected.sync="kindIndex"
                            :list="kindList"
                            :setting-key="'id'"
                            :display-key="'name'"
                            :allow-clear="true"
                            @clear="kindClear">
                        </bk-selector>
                    </div>
                    <div class="left">
                        <bk-selector :placeholder="$t('事件级别')"
                            :selected.sync="levelIndex"
                            :list="levelList"
                            :setting-key="'id'"
                            :display-key="'name'"
                            :allow-clear="true"
                            @clear="levelClear">
                        </bk-selector>
                    </div>
                    <div class="left component">
                        <bk-selector :placeholder="$t('组件')"
                            :selected.sync="componentIndex"
                            :list="componentList"
                            :setting-key="'id'"
                            :display-key="'name'"
                            :allow-clear="true"
                            @clear="componentClear">
                        </bk-selector>
                    </div>
                    <div class="left range-picker">
                        <bk-date-picker
                            :placeholder="$t('选择日期')"
                            :shortcuts="shortcuts"
                            :type="'datetimerange'"
                            :placement="'bottom-end'"
                            @change="change">
                        </bk-date-picker>
                    </div>
                    <div class="left">
                        <bk-button type="primary" :title="$t('查询')" icon="search" @click="handleClick">
                            {{$t('查询')}}
                        </bk-button>
                    </div>
                </div>
                <div class="biz-table-wrapper">
                    <bk-table
                        v-bkloading="{ isLoading: isPageLoading && !isInitLoading }"
                        :data="dataList"
                        :size="'medium'"
                        :page-params="pageConf"
                        @page-change="pageChangeHandler"
                        @page-limit-change="changePageSize">
                        <bk-table-column :label="$t('时间')" :show-overflow-tooltip="true" min-width="150" prop="eventTime" />
                        <bk-table-column :label="$t('组件')" min-width="150" prop="component" />
                        <bk-table-column :label="$t('对象及级别')" min-width="150" prop="extra">
                            <template slot-scope="{ row }">
                                <p class="extra-info" :title="row.extra.level || '--'"><span>{{$t('级别：')}}</span>{{ row.extra.level || '--' }}</p>
                                <p class="extra-info" :title="row.extra.kind || '--'"><span>{{$t('对象：')}}</span>{{ row.extra.kind || '--' }}</p>
                            </template>
                        </bk-table-column>
                        <bk-table-column :label="$t('所属集群')" min-width="200" prop="cluster_id">
                            <template slot-scope="{ row }">
                                <bcs-popover :content="row.cluster_id" placement="top">
                                    {{ row.clusterName || '--' }}
                                </bcs-popover>
                            </template>
                        </bk-table-column>
                        <bk-table-column :label="$t('事件内容')" min-width="150" prop="describe">
                            <template slot-scope="{ row }">
                                <bcs-popover placement="top" :delay="500">
                                    <div class="description">
                                        {{ row.describe || '--' }}
                                    </div>
                                    <template slot="content">
                                        <p style="text-align: left; white-space: normal;word-break: break-all;">{{row.describe}}</p>
                                    </template>
                                </bcs-popover>
                            </template>
                        </bk-table-column>
                    </bk-table>
                </div>
            </template>
        </div>
    </div>
</template>

<script>
    import moment from 'moment'

    export default {
        name: 'event-query',
        data () {
            // 事件对象下拉框 list
            const kindList = [
                { id: 'all', name: this.$t('全部1') },
                { id: 'rc', name: 'Rc' },
                { id: 'Endpoints', name: 'Endpoints' },
                { id: 'Pod', name: 'Pod' },
                { id: 'deployment', name: 'Deployment' },
                { id: 'Node', name: 'Node' },
                { id: 'HorizontalPodAutoscaler', name: 'HPA' },
                { id: 'Service', name: 'Service' },
                { id: 'task', name: 'Task' },
                { id: 'slaver', name: 'Slaver' }
            ]
            // 事件对象 map
            const kindMap = {}
            kindList.forEach(item => {
                kindMap[item.id] = item.name
            })

            // 事件级别下拉框 list
            const levelList = [
                { id: 'all', name: this.$t('全部1') },
                { id: 'Normal', name: 'Normal' },
                { id: 'Warning', name: 'Warning' }
            ]
            // 事件级别 map
            const levelMap = {}
            levelList.forEach(item => {
                levelMap[item.id] = item.name
            })

            // 组件下拉框 list
            const componentList = [
                { id: 'all', name: this.$t('全部1') },
                { id: 'scheduler/controller', name: 'Scheduler/controller' },
                { id: 'controller', name: 'Controller' },
                { id: 'kubelet', name: 'Kubelet' },
                { id: 'scheduler', name: 'Scheduler' }
            ]
            // 事件级别 map
            const componentMap = {}
            componentList.forEach(item => {
                componentMap[item.id] = item.name
            })

            return {
                ranges: {
                    [this.$t('昨天')]: [moment().subtract(1, 'days'), moment()],
                    [this.$t('最近一周')]: [moment().subtract(7, 'days'), moment()],
                    [this.$t('最近一个月')]: [moment().subtract(1, 'month'), moment()],
                    [this.$t('最近三个月')]: [moment().subtract(3, 'month'), moment()]
                },
                // 集群下拉框选中索引
                clusterIndex: -1,

                // 事件对象 map
                kindMap,
                // 事件对象下拉框 list
                kindList,
                kindIndex: -1,

                // 事件级别 map
                levelMap,
                // 事件级别下拉框 list
                levelList,
                levelIndex: -1,

                // 组件下拉框 map
                componentMap,
                // 组件下拉框 list
                componentList,
                componentIndex: -1,

                // 查询时间范围
                dataRange: ['', ''],
                // 列表数据
                dataList: [],
                isInitLoading: true,
                isPageLoading: false,

                pageConf: {
                    // 总数
                    total: 0,
                    // 总页数
                    totalPage: 1,
                    // 每页多少条
                    pageSize: 10,
                    // 当前页
                    curPage: 1,
                    // 是否显示翻页条
                    show: false
                },
                bkMessageInstance: null,
                shortcuts: [
                    {
                        text: this.$t('今天'),
                        value () {
                            const end = new Date()
                            const start = new Date(end.getFullYear(), end.getMonth(), end.getDate())
                            return [start, end]
                        }
                    },
                    {
                        text: this.$t('近7天'),
                        value () {
                            const end = new Date()
                            const start = new Date()
                            start.setTime(start.getTime() - 3600 * 1000 * 24 * 7)
                            return [start, end]
                        }
                    },
                    {
                        text: this.$t('近15天'),
                        value () {
                            const end = new Date()
                            const start = new Date()
                            start.setTime(start.getTime() - 3600 * 1000 * 24 * 15)
                            return [start, end]
                        }
                    },
                    {
                        text: this.$t('近30天'),
                        value () {
                            const end = new Date()
                            const start = new Date()
                            start.setTime(start.getTime() - 3600 * 1000 * 24 * 30)
                            return [start, end]
                        }
                    }
                ]
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
            dropdownClusterList () {
                return this.$store.state.cluster.clusterList
            }
        },
        watch: {
            curClusterId: {
                immediate: true,
                handler () {
                    this.clusterIndex = this.curClusterId
                    // this.handleClick()
                }
            }
        },
        mounted () {
            if (!this.curClusterId) {
                this.fetchData({
                    projId: this.projectId,
                    limit: this.pageConf.pageSize,
                    offset: 0
                })
            } else {
                this.fetchData({
                    projId: this.projectId,
                    cluster_id: this.curClusterId,
                    limit: this.pageConf.pageSize,
                    offset: 0
                })
            }
        },
        destroyed () {
            this.bkMessageInstance && this.bkMessageInstance.close()
        },
        methods: {
            /**
             * 分页大小更改
             *
             * @param {number} pageSize pageSize
             */
            changePageSize (pageSize) {
                this.pageConf.pageSize = pageSize
                this.pageConf.curPage = 1
                this.fetchData({
                    projId: this.projectId,
                    limit: this.pageConf.pageSize,
                    offset: 0
                })
            },

            /**
             * 获取表格数据
             *
             * @param {Object} params ajax 查询参数
             */
            async fetchData (params = {}) {
                // 集群
                // const clusterName = this.clusterIndex !== -1
                //     ? this.dropdownClusterList[this.clusterIndex].name
                //     : null
                const curCluster = this.dropdownClusterList.filter(
                    cluster => cluster.cluster_id === this.clusterIndex
                )[0]

                const clusterId = curCluster ? curCluster.cluster_id : null
                // this.clusterIndex === -1 ? null : this.clusterIndex

                // 事件对象
                // const kind = this.kindIndex !== -1
                //     ? this.kindList[this.kindIndex].id
                //     : null
                const kind = this.kindIndex === -1 ? null : this.kindIndex

                // 事件级别
                // const level = this.levelIndex !== -1
                //     ? this.levelList[this.levelIndex].id
                //     : null
                const level = this.levelIndex === -1 ? null : this.levelIndex

                // 组件
                // const component = this.componentIndex !== -1
                //     ? this.componentList[this.componentIndex].id
                //     : null
                const component = this.componentIndex === -1 ? null : this.componentIndex

                // 开始结束时间
                const [beginTime, endTime] = this.dataRange

                this.isPageLoading = true
                try {
                    const res = await this.$store.dispatch('mc/getActivityEvents', Object.assign({}, params, {
                        clusterId,
                        kind,
                        level,
                        component,
                        beginTime,
                        endTime
                    }))

                    this.dataList = []

                    const count = res.count
                    if (count <= 0) {
                        this.pageConf.show = false
                        this.pageConf.totalPage = 0
                        this.pageConf.total = 0
                        return
                    }

                    this.pageConf.total = count
                    this.pageConf.totalPage = Math.ceil(count / this.pageConf.pageSize)
                    if (this.pageConf.totalPage < this.pageConf.curPage) {
                        this.pageConf.curPage = 1
                    }
                    this.pageConf.show = true

                    const list = res.results || []
                    list.forEach(item => {
                        this.dataList.push({
                            // 事件时间
                            // 应该是 item.event_time 接口有问题，暂时先取 create_time
                            eventTime: moment(item.event_time).format('YYYY-MM-DD HH:mm:ss'),
                            // 组件
                            component: item.component,
                            cluster_id: item.cluster_id,
                            // 集群
                            clusterName: item.cluster_name,
                            extra: {
                                // 级别
                                level: item.level,
                                // 对象
                                kind: item.kind
                            },
                            // 描述
                            describe: item.describe
                        })
                    })
                } catch (e) {
                } finally {
                    this.isPageLoading = false
                    // 晚消失是为了防止整个页面loading和表格数据loading效果叠加产生闪动
                    setTimeout(() => {
                        this.isInitLoading = false
                    }, 200)
                }
            },

            /**
             * 翻页
             *
             * @param {number} page 页码
             */
            pageChangeHandler (page = 1) {
                this.pageConf.curPage = page
                this.fetchData({
                    projId: this.projectId,
                    limit: this.pageConf.pageSize,
                    offset: this.pageConf.pageSize * (page - 1)
                })
            },

            /**
             * 清除集群
             */
            clusterClear () {
                this.clusterIndex = -1
            },

            /**
             * 清除事件对象
             */
            kindClear () {
                this.kindIndex = -1
            },

            /**
             * 清除事件级别
             */
            levelClear () {
                this.levelIndex = -1
            },

            /**
             * 清除组件
             */
            componentClear () {
                this.componentIndex = -1
            },

            /**
             * 日期范围搜索条件
             *
             * @param {string} newValue 变化前的值
             */
            change (newValue) {
                this.dataRange = newValue
            },

            /**
             * 搜索按钮点击
             *
             * @param {Object} e 时间对象
             */
            handleClick (e) {
                this.pageConf.curPage = 1
                this.fetchData({
                    projId: this.projectId,
                    limit: this.pageConf.pageSize,
                    offset: 0
                })
            }
        }
    }
</script>

<style scoped>
    @import './event-query.css';
</style>
