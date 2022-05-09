<template>
    <div class="biz-content">
        <div class="biz-top-bar">
            <div class="biz-cluster-node-overview-title">
                <i class="bcs-icon bcs-icon-arrows-left back" @click="goNode"></i>
                <span @click="refreshCurRouter">{{nodeId}}</span>
            </div>
            <bk-guide></bk-guide>
        </div>
        <div class="biz-content-wrapper biz-cluster-node-overview">
            <app-exception
                v-if="exceptionCode"
                :type="exceptionCode.code"
                :text="exceptionCode.msg">
            </app-exception>
            <div v-else class="biz-cluster-node-overview-wrapper">
                <div class="biz-cluster-node-overview-header">
                    <div class="header-item">
                        <div class="key-label">IP：</div>
                        <bcs-popover :content="nodeId" placement="bottom">
                            <div class="value-label">{{nodeId}}</div>
                        </bcs-popover>
                    </div>
                    <div class="header-item">
                        <div class="key-label">CPU：</div>
                        <bcs-popover :content="nodeInfo.cpu_count" placement="bottom">
                            <div class="value-label">{{nodeInfo.cpu_count}}</div>
                        </bcs-popover>
                    </div>
                    <div class="header-item">
                        <div class="key-label">{{$t('内存：')}}</div>
                        <bcs-popover :content="nodeInfo.memory" placement="bottom">
                            <div class="value-label">{{nodeInfo.memory}}</div>
                        </bcs-popover>
                    </div>
                    <div class="header-item">
                        <div class="key-label">{{$t('存储：')}}</div>
                        <bcs-popover :content="nodeInfo.disk" placement="bottom">
                            <div class="value-label">{{nodeInfo.disk}}</div>
                        </bcs-popover>
                    </div>
                    <div class="header-item">
                        <div class="key-label">{{$t('IP来源：')}}</div>
                        <bcs-popover :content="nodeInfo.provider" placement="bottom">
                            <div class="value-label">{{nodeInfo.provider}}</div>
                        </bcs-popover>
                    </div>
                    <div class="header-item">
                        <div class="key-label">{{$t('内核：')}}</div>
                        <bcs-popover :content="nodeInfo.release" placement="bottom">
                            <div class="value-label">{{nodeInfo.release}}</div>
                        </bcs-popover>
                    </div>
                    <div class="header-item">
                        <div class="key-label">Docker：</div>
                        <bcs-popover :content="nodeInfo.dockerVersion" placement="bottom">
                            <div class="value-label">{{nodeInfo.dockerVersion}}</div>
                        </bcs-popover>
                    </div>
                    <div class="header-item">
                        <div class="key-label">{{$t('操作系统：')}}</div>
                        <bcs-popover :content="nodeInfo.sysname" placement="bottom">
                            <div class="value-label">{{nodeInfo.sysname}}</div>
                        </bcs-popover>
                    </div>
                </div>
                <div class="biz-cluster-node-overview-chart-wrapper">
                    <div class="biz-cluster-node-overview-chart">
                        <div class="part top-left">
                            <div class="info">
                                <div class="left">{{$t('CPU使用率')}}</div>
                                <div class="right">
                                    <bk-dropdown-menu :align="'right'" ref="cpuDropdown">
                                        <div style="cursor: pointer;" slot="dropdown-trigger">
                                            <span>{{cpuToggleRangeStr}}</span>
                                            <button class="biz-dropdown-button">
                                                <i class="bcs-icon bcs-icon-angle-down" style="margin-top: 1px;"></i>
                                            </button>
                                        </div>
                                        <ul class="bk-dropdown-list" slot="dropdown-content">
                                            <li>
                                                <a href="javascript:;" @click.stop="toggleRange('cpuDropdown', 'cpuToggleRangeStr', 'cpu_summary', '1')">{{$t('1小时')}}</a>
                                            </li>
                                            <li>
                                                <a href="javascript:;" @click.stop="toggleRange('cpuDropdown', 'cpuToggleRangeStr', 'cpu_summary', '2')">{{$t('24小时')}}</a>
                                            </li>
                                            <li>
                                                <a href="javascript:;" @click.stop="toggleRange('cpuDropdown', 'cpuToggleRangeStr', 'cpu_summary', '3')">{{$t('近7天')}}</a>
                                            </li>
                                        </ul>
                                    </bk-dropdown-menu>
                                </div>
                            </div>
                            <chart :options="cpuChartOptsK8S" ref="cpuLine1" auto-resize></chart>
                        </div>
                        <div class="part top-right">
                            <div class="info">
                                <div class="left">{{$t('内存使用率')}}</div>
                                <div class="right">
                                    <bk-dropdown-menu :align="'right'" ref="memoryDropdown">
                                        <div style="cursor: pointer;" slot="dropdown-trigger">
                                            <span>{{memToggleRangeStr}}</span>
                                            <button class="biz-dropdown-button">
                                                <i class="bcs-icon bcs-icon-angle-down" style="margin-top: 1px;"></i>
                                            </button>
                                        </div>
                                        <ul class="bk-dropdown-list" slot="dropdown-content">
                                            <li>
                                                <a href="javascript:;" @click.stop="toggleRange('memoryDropdown', 'memToggleRangeStr', 'mem', '1')">{{$t('1小时')}}</a>
                                            </li>
                                            <li>
                                                <a href="javascript:;" @click.stop="toggleRange('memoryDropdown', 'memToggleRangeStr', 'mem', '2')">{{$t('24小时')}}</a>
                                            </li>
                                            <li>
                                                <a href="javascript:;" @click.stop="toggleRange('memoryDropdown', 'memToggleRangeStr', 'mem', '3')">{{$t('近7天')}}</a>
                                            </li>
                                        </ul>
                                    </bk-dropdown-menu>
                                </div>
                            </div>
                            <chart :options="memChartOptsK8S" ref="memoryLine1" auto-resize></chart>
                        </div>
                    </div>
                    <div class="biz-cluster-node-overview-chart">
                        <div class="part bottom-left">
                            <div class="info">
                                <div class="left">{{$t('网络')}}</div>
                                <div class="right">
                                    <bk-dropdown-menu :align="'right'" ref="networkDropdown">
                                        <div style="cursor: pointer;" slot="dropdown-trigger">
                                            <span>{{networkToggleRangeStr}}</span>
                                            <button class="biz-dropdown-button" style="vertical-align: middle;">
                                                <i class="bcs-icon bcs-icon-angle-down" style="margin-top: 1px;"></i>
                                            </button>
                                        </div>
                                        <ul class="bk-dropdown-list" slot="dropdown-content">
                                            <li>
                                                <a href="javascript:;" @click.stop="toggleRange('networkDropdown', 'networkToggleRangeStr', 'net', '1')">{{$t('1小时')}}</a>
                                            </li>
                                            <li>
                                                <a href="javascript:;" @click.stop="toggleRange('networkDropdown', 'networkToggleRangeStr', 'net', '2')">{{$t('24小时')}}</a>
                                            </li>
                                            <li>
                                                <a href="javascript:;" @click.stop="toggleRange('networkDropdown', 'networkToggleRangeStr', 'net', '3')">{{$t('近7天')}}</a>
                                            </li>
                                        </ul>
                                    </bk-dropdown-menu>
                                </div>
                            </div>
                            <chart :options="networkChartOptsK8S" ref="networkLine1" auto-resize></chart>
                        </div>
                        <div class="part">
                            <div class="info">
                                <div class="left">{{$t('IO使用率')}}</div>
                                <div class="right">
                                    <bk-dropdown-menu :align="'right'" ref="storageDropdown">
                                        <div style="cursor: pointer;" slot="dropdown-trigger">
                                            <span>{{storageToggleRangeStr}}</span>
                                            <button class="biz-dropdown-button">
                                                <i class="bcs-icon bcs-icon-angle-down" style="margin-top: 1px;"></i>
                                            </button>
                                        </div>
                                        <ul class="bk-dropdown-list" slot="dropdown-content">
                                            <li>
                                                <a href="javascript:;" @click.stop="toggleRange('storageDropdown', 'storageToggleRangeStr', 'io', '1')">{{$t('1小时')}}</a>
                                            </li>
                                            <li>
                                                <a href="javascript:;" @click.stop="toggleRange('storageDropdown', 'storageToggleRangeStr', 'io', '2')">{{$t('24小时')}}</a>
                                            </li>
                                            <li>
                                                <a href="javascript:;" @click.stop="toggleRange('storageDropdown', 'storageToggleRangeStr', 'io', '3')">{{$t('近7天')}}</a>
                                            </li>
                                        </ul>
                                    </bk-dropdown-menu>
                                </div>
                            </div>
                            <chart :options="diskioChartOptsK8S" ref="storageLine1" auto-resize></chart>
                        </div>
                    </div>
                </div>
                <div class="biz-cluster-node-overview-table-wrapper">
                    <!-- <bk-tab class="biz-tab-container" :type="'fill'" :active-name="'container'" @tab-changed="tabChanged">
                        <bk-tab-panel name="container" :title="$t('容器')">
                            <div class="container-table-wrapper" v-bkloading="{ isLoading: containerTableLoading }">
                                <table class="bk-table has-table-hover biz-table biz-cluster-node-overview-table">
                                    <thead>
                                        <tr>
                                            <th style="padding-left: 20px;">{{$t('名称')}}</th>
                                            <th>{{$t('状态')}}</th>
                                            <th>{{$t('镜像')}}</th>
                                            <th style="min-width: 120px;">{{$t('操作')}}</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        <template v-if="containerTableCurPageData.length">
                                            <tr v-for="(containerTableItem, index) in containerTableCurPageData" :key="index">
                                                <td style="padding-left: 20px;" v-if="(curProject.kind === 1 || curProject.kind === 3) && containerTableItem.status !== 'running'">
                                                    <div class="name">
                                                        <a href="javascript:void(0)" class="bk-text-button is-disabled">{{containerTableItem.name}}</a>
                                                    </div>
                                                </td>
                                                <td style="padding-left: 20px;" v-else>
                                                    <bcs-popover placement="top" :delay="500" :transfer="true">
                                                        <div class="name">
                                                            <a href="javascript:void(0)" @click="goContainerDetail(containerTableItem)" class="bk-text-button">{{containerTableItem.name}}</a>
                                                        </div>
                                                        <template slot="content">
                                                            <p style="text-align: left; white-space: normal;word-break: break-all;font-weight: 400;">{{containerTableItem.name}}</p>
                                                        </template>
                                                    </bcs-popover>
                                                </td>
                                                <td v-if="containerTableItem.status === 'terminated'">
                                                    <i class="bcs-icon bcs-icon-circle-shape danger"></i>terminated
                                                </td>
                                                <td v-else-if="containerTableItem.status === 'running'">
                                                    <i class="bcs-icon bcs-icon-circle-shape running"></i>running
                                                </td>
                                                <td v-else>
                                                    <i class="bcs-icon bcs-icon-circle-shape warning"></i>{{containerTableItem.status}}
                                                </td>
                                                <td>
                                                    <bcs-popover placement="top" :delay="500" :transfer="true">
                                                        <div class="mirror">
                                                            {{containerTableItem.image}}
                                                        </div>
                                                        <template slot="content">
                                                            <p style="text-align: left; white-space: normal; word-break: break-all;font-weight: 400;">{{containerTableItem.image}}</p>
                                                        </template>
                                                    </bcs-popover>
                                                </td>
                                                <td>
                                                    <template v-if="containerTableItem.status === 'running'">
                                                        <a href="javascript: void(0);" class="bk-text-button" @click.stop="showTerminal(containerTableItem)">WebConsole</a>
                                                    </template>
                                                    <template v-else>
                                                        <bcs-popover :content="$t('容器状态不是running')" placement="right">
                                                            <a href="javascript: void(0);" class="bk-text-button is-disabled">WebConsole</a>
                                                        </bcs-popover>
                                                    </template>
                                                </td>
                                            </tr>
                                        </template>
                                        <template v-else>
                                            <tr>
                                                <td colspan="4">
                                                    <div class="bk-message-box no-data">
                                                        <bcs-exception type="empty" scene="part"></bcs-exception>
                                                    </div>
                                                </td>
                                            </tr>
                                        </template>
                                    </tbody>
                                </table>
                                <div class="biz-page-box biz-cluster-node-overview-page" v-if="containerTablePageConf.show">
                                    <bk-pagination
                                        :show-limit="false"
                                        :current.sync="containerTablePageConf.curPage"
                                        :count.sync="containerTablePageConf.count"
                                        :limit="containerTablePageConf.pageSize"
                                        :limit-list="containerTablePageConf.limitList"
                                        @change="pageChange">
                                    </bk-pagination>
                                </div>
                            </div>
                        </bk-tab-panel>
                    </bk-tab> -->
                </div>
            </div>
        </div>
    </div>
</template>

<script>
    import moment from 'moment'
    import ECharts from 'vue-echarts/components/ECharts.vue'
    import 'echarts/lib/chart/line'
    import 'echarts/lib/component/tooltip'
    import 'echarts/lib/component/legend'

    import { nodeOverview } from '@/common/chart-option'
    import { catchErrorHandler, formatBytes } from '@/common/util'

    import { createChartOption } from './node-overview-chart-opts'

    export default {
        components: {
            chart: ECharts
        },
        data () {
            return {
                PROJECT_MESOS: PROJECT_MESOS,
                tabActiveName: 'container',

                cpuLine: nodeOverview.cpu,
                cpuChartOptsK8S: createChartOption(this),

                memoryLine: nodeOverview.memory,
                memChartOptsK8S: createChartOption(this),

                networkLine: nodeOverview.network,
                networkChartOptsK8S: createChartOption(this),

                storageLine: nodeOverview.storage,
                diskioChartOptsK8S: createChartOption(this),

                bkMessageInstance: null,
                cpuToggleRangeStr: this.$t('1小时'),
                memToggleRangeStr: this.$t('1小时'),
                networkToggleRangeStr: this.$t('1小时'),
                storageToggleRangeStr: this.$t('1小时'),
                nodeInfo: {},
                containerTableLoading: false,
                containerTableList: [],
                containerTablePageConf: {
                    count: 1,
                    totalPage: 1,
                    pageSize: 5,
                    curPage: 1,
                    show: false,
                    limitList: [5, 10, 20, 100]
                },
                containerTableCurPageData: [],
                labelList: [],
                labelListLoading: true,
                exceptionCode: null,
                terminalWins: {}
            }
        },
        computed: {
            projectId () {
                return this.$route.params.projectId
            },
            projectCode () {
                return this.$route.params.projectCode
            },
            clusterId () {
                return this.$route.params.clusterId
            },
            nodeId () {
                return this.$route.params.nodeId
            },
            onlineProjectList () {
                return this.$store.state.sideMenu.onlineProjectList
            },
            curProject () {
                return this.$store.state.curProject
            }
        },
        created () {
            this.cpuLine.series[0].data
                = this.memoryLine.series[0].data = this.memoryLine.series[1].data
                = this.networkLine.series[0].data = this.networkLine.series[1].data
                = this.storageLine.series[0].data = this.storageLine.series[1].data
                = [0]
            nodeOverview.storage.series[0].data = [9, 0, 22, 40, 12, 31, 2, 12, 18, 27, 27]
        },
        async mounted () {
            const cpuRef = this.$refs.cpuLine1
            cpuRef && cpuRef.showLoading({
                text: this.$t('正在加载中...'),
                color: '#30d878',
                maskColor: 'rgba(255, 255, 255, 0.8)'
            })

            const memRef = this.$refs.memoryLine1
            memRef && memRef.showLoading({
                text: this.$t('正在加载中...'),
                color: '#30d878',
                maskColor: 'rgba(255, 255, 255, 0.8)'
            })

            const diskioRef = this.$refs.storageLine1
            diskioRef && diskioRef.showLoading({
                text: this.$t('正在加载中...'),
                color: '#30d878',
                maskColor: 'rgba(255, 255, 255, 0.8)'
            })

            const netRef = this.$refs.networkLine1
            netRef && netRef.showLoading({
                text: this.$t('正在加载中...'),
                color: '#30d878',
                maskColor: 'rgba(255, 255, 255, 0.8)'
            })

            await this.fetchNodeInfo()
            // this.fetchNodeContainers()

            this.fetchDataK8S('cpu_summary', '1')
            this.fetchDataK8S('mem', '1')
            this.fetchDataK8S('net', '1')
            this.fetchDataK8S('io', '1')
        },
        methods: {
            /**
             * 打开到终端入口
             *
             * @param {Object} container 当前容器
             */
            async showTerminal (container) {
                const clusterId = this.$route.params.clusterId
                const containerId = container.container_id
                const url = `${DEVOPS_BCS_API_URL}/web_console/projects/${this.projectId}/clusters/${clusterId}/?container_id=${containerId}`
                if (this.terminalWins.hasOwnProperty(containerId)) {
                    const win = this.terminalWins[containerId]
                    if (!win.closed) {
                        this.terminalWins[containerId].focus()
                    } else {
                        const win = window.open(url, '', 'width=990,height=618')
                        this.terminalWins[containerId] = win
                    }
                } else {
                    const win = window.open(url, '', 'width=990,height=618')
                    this.terminalWins[containerId] = win
                }
            },

            /**
             * 获取上方的信息
             */
            async fetchNodeInfo () {
                const { projectId, clusterId, nodeId } = this
                try {
                    const res = await this.$store.dispatch('cluster/nodeInfo', {
                        projectId,
                        clusterId,
                        nodeId
                    })

                    const nodeInfo = {}

                    const data = res.data || {}

                    nodeInfo.id = data.id || ''
                    nodeInfo.provider = data.provider || '--'
                    nodeInfo.dockerVersion = data.dockerVersion || '--'
                    nodeInfo.osVersion = data.osVersion || '--'
                    nodeInfo.domainname = data.domainname || '--'
                    nodeInfo.machine = data.machine || '--'
                    nodeInfo.nodename = data.nodename || '--'
                    nodeInfo.release = data.release || '--'
                    nodeInfo.sysname = data.sysname || '--'
                    nodeInfo.version = data.version || '--'
                    nodeInfo.cpu_count = data.cpu_count || '--'
                    nodeInfo.memory = data.memory ? formatBytes(data.memory) : '--'
                    nodeInfo.disk = data.disk ? formatBytes(data.disk) : '--'

                    this.nodeInfo = Object.assign({}, nodeInfo)
                } catch (e) {
                    catchErrorHandler(e, this)
                }
            },

            /**
             * 获取下方容器选项卡表格数据
             */
            async fetchNodeContainers () {
                const { projectId, clusterId, nodeId } = this
                this.containerTableLoading = true
                try {
                    const res = await this.$store.dispatch('cluster/getNodeContainerList', {
                        projectId,
                        clusterId,
                        nodeId
                    })
                    this.containerTableList = res.data || []
                    this.initPageConf()
                    this.containerTableCurPageData = this.getDataByPage(this.containerTablePageConf.curPage)
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    this.containerTableLoading = false
                }
            },

            /**
             * 获取下方 tab 标签的数据
             */
            async fetchLabel () {
                this.labelListLoading = true
                try {
                    const res = await this.$store.dispatch('cluster/getNodeLabel', {
                        projectId: this.projectId,
                        nodeIds: [this.nodeInfo.id]
                    })

                    const labelList = []
                    const labels = res.data || {}
                    Object.keys(labels).forEach(key => {
                        labelList.push({
                            key: key,
                            val: labels[key]
                        })
                    })
                    this.labelList.splice(0, this.labelList.length, ...labelList)
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    this.labelListLoading = false
                }
            },

            /**
             * 初始化下方容器选项卡表格翻页条
             */
            initPageConf () {
                const total = this.containerTableList.length
                this.containerTablePageConf.show = total > 0
                this.containerTablePageConf.totalPage = Math.ceil(total / this.containerTablePageConf.pageSize) || 1
                this.containerTablePageConf.count = total
            },

            /**
             * 获取当前这一页的数据
             *
             * @param {number} page 当前页
             *
             * @return {Array} 当前页数据
             */
            getDataByPage (page) {
                let startIndex = (page - 1) * this.containerTablePageConf.pageSize
                let endIndex = page * this.containerTablePageConf.pageSize
                if (startIndex < 0) {
                    startIndex = 0
                }
                if (endIndex > this.containerTableList.length) {
                    endIndex = this.containerTableList.length
                }
                const data = this.containerTableList.slice(startIndex, endIndex)
                return data
            },

            /**
             * 翻页回调
             *
             * @param {number} page 当前页
             */
            pageChange (page) {
                this.containerTablePageConf.curPage = page
                const data = this.getDataByPage(page)
                this.containerTableCurPageData.splice(0, this.containerTableCurPageData.length, ...data)
            },

            /**
             * 获取中间图表数据 k8s
             *
             * @param {string} idx 标识，cpu / memory / network / storage
             * @param {string} range 时间范围，1: 1 小时，2: 24 小时，3：近 7 天
             */
            async fetchDataK8S (idx, range) {
                const params = {
                    startAt: null,
                    endAt: moment().format('YYYY-MM-DD HH:mm:ss'),
                    projectId: this.projectId,
                    resId: this.nodeId,
                    clusterId: this.clusterId
                }

                // 1 小时
                if (range === '1') {
                    params.startAt = moment().subtract(1, 'hours').format('YYYY-MM-DD HH:mm:ss')
                } else if (range === '2') { // 24 小时
                    params.startAt = moment().subtract(1, 'days').format('YYYY-MM-DD HH:mm:ss')
                } else if (range === '3') { // 近 7 天
                    params.startAt = moment().subtract(7, 'days').format('YYYY-MM-DD HH:mm:ss')
                }

                try {
                    if (idx === 'net') {
                        const res = await Promise.all([
                            this.$store.dispatch('cluster/nodeNetReceive', Object.assign({}, params)),
                            this.$store.dispatch('cluster/nodeNetTransmit', Object.assign({}, params))
                        ])
                        this.renderNetChart(res[0].data.result, res[1].data.result)
                    } else {
                        let url = ''
                        let renderFn = ''
                        if (idx === 'cpu_summary') {
                            url = 'cluster/nodeCpuUsage'
                            renderFn = 'renderCpuChart'
                        }

                        if (idx === 'mem') {
                            url = 'cluster/nodeMemUsage'
                            renderFn = 'renderMemChart'
                        }

                        if (idx === 'io') {
                            url = 'cluster/nodeDiskioUsage'
                            renderFn = 'renderDiskioChart'
                        }

                        const res = await this.$store.dispatch(url, params)
                        this[renderFn](res.data.result || [])
                    }
                } catch (e) {
                    catchErrorHandler(e, this)
                }
            },

            /**
             * 渲染 cpu 图表
             *
             * @param {Array} list 图表数据
             */
            renderCpuChart (list) {
                const chartNode = this.$refs.cpuLine1
                if (!chartNode) {
                    return
                }

                const cpuChartOptsK8S = Object.assign({}, this.cpuChartOptsK8S)
                cpuChartOptsK8S.series.splice(0, cpuChartOptsK8S.series.length, ...[])

                const data = list.length ? list : [{ values: [[parseInt(String(+new Date()).slice(0, 10), 10), '0']] }]

                data.forEach(item => {
                    item.values.forEach(d => {
                        d[0] = parseInt(d[0] + '000', 10)
                    })
                    cpuChartOptsK8S.series.push(
                        {
                            type: 'line',
                            showSymbol: false,
                            smooth: true,
                            hoverAnimation: false,
                            areaStyle: {
                                normal: {
                                    opacity: 0.2
                                }
                            },
                            itemStyle: {
                                normal: {
                                    color: '#30d878'
                                }
                            },
                            data: item.values
                        }
                    )
                })

                const label = this.$t('CPU使用率')
                chartNode.mergeOptions({
                    tooltip: {
                        formatter (params, ticket, callback) {
                            let ret = ''

                            if (params[0].value[1] === '-') {
                                ret = '<div>No Data</div>'
                            } else {
                                ret = `
                                    <div>${moment(parseInt(params[0].value[0], 10)).format('YYYY-MM-DD HH:mm:ss')}</div>
                                    <div>${label}：${parseFloat(params[0].value[1]).toFixed(2)}%</div>
                                `
                            }

                            return ret
                        }
                    }
                })

                chartNode.hideLoading()
            },

            /**
             * 渲染 mem 图表
             *
             * @param {Array} list 图表数据
             */
            renderMemChart (list) {
                const chartNode = this.$refs.memoryLine1
                if (!chartNode) {
                    return
                }

                const memChartOptsK8S = Object.assign({}, this.memChartOptsK8S)
                memChartOptsK8S.series.splice(0, memChartOptsK8S.series.length, ...[])

                const data = list.length ? list : [{ values: [[parseInt(String(+new Date()).slice(0, 10), 10), '0']] }]

                data.forEach(item => {
                    item.values.forEach(d => {
                        d[0] = parseInt(d[0] + '000', 10)
                    })
                    memChartOptsK8S.series.push(
                        {
                            type: 'line',
                            showSymbol: false,
                            smooth: true,
                            hoverAnimation: false,
                            areaStyle: {
                                normal: {
                                    opacity: 0.2
                                }
                            },
                            itemStyle: {
                                normal: {
                                    color: '#3a84ff'
                                }
                            },
                            data: item.values
                        }
                    )
                })

                const label = this.$t('内存使用率')
                chartNode.mergeOptions({
                    tooltip: {
                        formatter (params, ticket, callback) {
                            let ret = ''

                            if (params[0].value[1] === '-') {
                                ret = '<div>No Data</div>'
                            } else {
                                ret = `
                                    <div>${moment(parseInt(params[0].value[0], 10)).format('YYYY-MM-DD HH:mm:ss')}</div>
                                    <div>${label}：${parseFloat(params[0].value[1]).toFixed(2)}%</div>
                                `
                            }

                            return ret
                        }
                    }
                })

                chartNode.hideLoading()
            },

            /**
             * 渲染 diskio 图表
             *
             * @param {Array} list 图表数据
             */
            renderDiskioChart (list) {
                const chartNode = this.$refs.storageLine1
                if (!chartNode) {
                    return
                }

                const diskioChartOptsK8S = Object.assign({}, this.diskioChartOptsK8S)
                diskioChartOptsK8S.series.splice(0, diskioChartOptsK8S.series.length, ...[])

                const data = list.length ? list : [{ values: [[parseInt(String(+new Date()).slice(0, 10), 10), '0']] }]

                data.forEach(item => {
                    item.values.forEach(d => {
                        d[0] = parseInt(d[0] + '000', 10)
                    })
                    diskioChartOptsK8S.series.push(
                        {
                            type: 'line',
                            showSymbol: false,
                            smooth: true,
                            hoverAnimation: false,
                            areaStyle: {
                                normal: {
                                    opacity: 0.2
                                }
                            },
                            itemStyle: {
                                normal: {
                                    color: '#ffbe21'
                                }
                            },
                            data: item.values
                        }
                    )
                })

                const label = this.$t('磁盘IO')
                chartNode.mergeOptions({
                    tooltip: {
                        formatter (params, ticket, callback) {
                            let ret = ''

                            if (params[0].value[1] === '-') {
                                ret = '<div>No Data</div>'
                            } else {
                                ret = `
                                    <div>${moment(parseInt(params[0].value[0], 10)).format('YYYY-MM-DD HH:mm:ss')}</div>
                                    <div>${label}：${parseFloat(params[0].value[1]).toFixed(2)}%</div>
                                `
                            }

                            return ret
                        }
                    }
                })

                chartNode.hideLoading()
            },

            /**
             * 渲染 net 图表
             *
             * @param {Array} listReceive net 入流量数据
             * @param {Array} listTransmit net 出流量数据
             */
            renderNetChart (listReceive, listTransmit) {
                const chartNode = this.$refs.networkLine1
                if (!chartNode) {
                    return
                }

                const networkChartOptsK8S = Object.assign({}, this.networkChartOptsK8S)
                networkChartOptsK8S.series.splice(0, networkChartOptsK8S.series.length, ...[])

                networkChartOptsK8S.yAxis.splice(0, networkChartOptsK8S.yAxis.length, ...[
                    {
                        axisLabel: {
                            formatter (value, index) {
                                return `${formatBytes(value)}`
                            }
                        }
                    }
                ])

                const dataReceive = listReceive.length
                    ? listReceive
                    : [{ values: [[parseInt(String(+new Date()).slice(0, 10), 10), '0']] }]

                const dataTransmit = listTransmit.length
                    ? listTransmit
                    : [{ values: [[parseInt(String(+new Date()).slice(0, 10), 10), '0']] }]

                dataReceive.forEach(item => {
                    item.values.forEach(d => {
                        d[0] = parseInt(d[0] + '000', 10)
                        d.push('receive')
                    })
                    networkChartOptsK8S.series.push(
                        {
                            type: 'line',
                            showSymbol: false,
                            smooth: true,
                            hoverAnimation: false,
                            areaStyle: {
                                normal: {
                                    opacity: 0.2
                                }
                            },
                            itemStyle: {
                                normal: {
                                    color: '#853cff'
                                }
                            },
                            data: item.values
                        }
                    )
                })

                dataTransmit.forEach(item => {
                    item.values.forEach(d => {
                        d[0] = parseInt(d[0] + '000', 10)
                        d.push('transmit')
                    })
                    networkChartOptsK8S.series.push(
                        {
                            type: 'line',
                            showSymbol: false,
                            smooth: true,
                            hoverAnimation: false,
                            areaStyle: {
                                normal: {
                                    opacity: 0.2
                                }
                            },
                            itemStyle: {
                                normal: {
                                    color: '#3dda80'
                                }
                            },
                            data: item.values
                        }
                    )
                })

                const labelReceive = this.$t('入流量')
                const labelTransmit = this.$t('出流量')
                chartNode.mergeOptions({
                    tooltip: {
                        formatter (params, ticket, callback) {
                            let ret = ''
                                + `<div>${moment(parseInt(params[0].value[0], 10)).format('YYYY-MM-DD HH:mm:ss')}</div>`

                            params.forEach(p => {
                                if (p.value[2] === 'receive') {
                                    ret += `<div>${labelReceive}：${formatBytes(p.value[1])}</div>`
                                } else if (p.value[2] === 'transmit') {
                                    ret += `<div>${labelTransmit}：${formatBytes(p.value[1])}</div>`
                                }
                            })

                            return ret
                        }
                    }
                })

                chartNode.hideLoading()
            },

            /**
             * 切换时间范围
             *
             * @param {Object} dropdownRef dropdown 标识
             * @param {string} toggleRangeStr 标识
             * @param {string} idx 标识，cpu / memory / network / storage
             * @param {string} range 时间范围，1: 1 小时，2: 24 小时，3：近 7 天
             */
            toggleRange (dropdownRef, toggleRangeStr, idx, range) {
                if (range === '1') {
                    this[toggleRangeStr] = this.$t('1小时')
                } else if (range === '2') {
                    this[toggleRangeStr] = this.$t('24小时')
                } else if (range === '3') {
                    this[toggleRangeStr] = this.$t('近7天')
                }

                this.$refs[dropdownRef].hide()

                let ref = null
                if (idx === 'cpu_summary') {
                    ref = this.$refs.cpuLine1
                }
                if (idx === 'mem') {
                    ref = this.$refs.memoryLine1
                }
                if (idx === 'io') {
                    ref = this.$refs.storageLine1
                }
                if (idx === 'net') {
                    ref = this.$refs.networkLine1
                }
                ref && ref.showLoading({
                    text: this.$t('正在加载中...'),
                    color: '#30d878',
                    maskColor: 'rgba(255, 255, 255, 0.8)'
                })

                this.fetchDataK8S(idx, range)
            },

            /**
             * 刷新当前 router
             */
            refreshCurRouter () {
                typeof this.$parent.refreshRouterView === 'function' && this.$parent.refreshRouterView()
            },

            /**
             * 返回节点管理
             */
            goNode () {
                this.$router.back()
                // const { params } = this.$route
                // if (params.backTarget) {
                //     this.$router.push({
                //         name: params.backTarget,
                //         params: {
                //             projectId: this.projectId,
                //             projectCode: this.projectCode
                //         }
                //     })
                // } else {
                //     this.$router.push({
                //         name: 'clusterNode',
                //         params: {
                //             projectId: this.projectId,
                //             projectCode: this.projectCode,
                //             clusterId: this.clusterId
                //         }
                //     })
                // }
            },

            /**
             * 跳转到容器详情
             *
             * @param {Object} container 当前容器对象
             */
            async goContainerDetail (container) {
                this.$router.push({
                    name: 'containerDetailForNode',
                    params: {
                        projectId: this.projectId,
                        projectCode: this.projectCode,
                        clusterId: this.clusterId,
                        containerId: container.container_id,
                        nodeId: this.nodeId
                    }
                })
            },

            tabChanged (name) {
                if (this.tabActiveName === name) {
                    return
                }
                this.tabActiveName = name

                clearTimeout(this.taskgroupTimer)
                this.taskgroupTimer = null

                this.openTaskgroup = Object.assign({}, {})

                if (name === 'label') {
                    this.labelList.splice(0, this.labelList.length, ...[])
                    this.fetchLabel()
                } else if (name === 'container') {
                    this.containerTableList.splice(0, this.containerTableList.length, ...[])
                    this.containerTablePageConf.curPage = 1
                    this.fetchNodeContainers()
                }
            }
        }
    }
</script>

<style scoped lang="postcss">
    @import '@/css/variable.css';

    .biz-cluster-node-overview {
        padding: 20px;
    }

    .biz-cluster-node-overview-title {
        display: inline-block;
        height: 60px;
        line-height: 60px;
        font-size: 16px;
        margin-left: 20px;
        cursor: pointer;

        .back {
            font-size: 16px;
            font-weight: 700;
            position: relative;
            top: 1px;
            color: $iconPrimaryColor;
        }
    }

    .biz-cluster-node-overview-wrapper {
        background-color: $bgHoverColor;
        display: inline-block;
        width: 100%;
    }

    .biz-cluster-node-overview-header {
        display: flex;
        border: 1px solid $borderWeightColor;
        border-radius: 2px;

        .header-item {
            font-size: 14px;
            flex: 1;
            height: 75px;
            border-right: 1px solid $borderWeightColor;
            padding-left: 20px;

            &:last-child {
                border-right: none;
            }

            .key-label {
                font-weight: 700;
                padding-top: 13px;
                padding-bottom: 5px;
            }

            .value-label {
                max-width: 130px;
                padding-top: 4px;
                overflow: hidden;
                text-overflow: ellipsis;
                white-space: nowrap;
            }
        }
    }

    .biz-cluster-node-overview-chart-wrapper {
        margin-top: 20px;
        background-color: #fff;
        box-shadow: 1px 0 2px rgba(0, 0, 0, 0.1);
        border: 1px solid $borderWeightColor;
        font-size: 0;
        border-radius: 2px;

        .biz-cluster-node-overview-chart {
            display: inline-block;
            width: 100%;

            .part {
                width: 50%;
                float: left;
                height: 250px;

                &.top-left {
                    border-right: 1px solid $borderWeightColor;
                    border-bottom: 1px solid $borderWeightColor;
                }

                &.top-right {
                    border-bottom: 1px solid $borderWeightColor;
                }

                &.bottom-left {
                    border-right: 1px solid $borderWeightColor;
                }

                .info {
                    font-size: 14px;
                    display: flex;
                    padding: 20px 30px;

                    .left,
                    .right {
                        flex: 1;
                    }

                    .left {
                        font-weight: 700;
                    }

                    .right {
                        text-align: right;
                    }
                }
            }
        }
    }

    .echarts {
        width: 100%;
        height: 180px;
    }

    .biz-cluster-node-overview-table-wrapper {
        margin-top: 20px;
    }

    .biz-cluster-node-overview-table {
        border-bottom: none;

        .name {
            width: 400px;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        .mirror {
            width: 500px;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        i {
            top: 1px;
            position: relative;
            margin-right: 7px;

            &.running {
                color: $iconSuccessColor;
            }

            &.warning {
                color: $iconWarningColor;
            }

            &.danger {
                color: $failColor;
            }
        }
    }

    .biz-cluster-node-overview-page {
        border-top: 1px solid #e6e6e6;
        padding: 20px 40px 20px 0;
    }

    @media screen and (max-width: $mediaWidth) {
        .biz-cluster-node-overview-header {
            .header-item {
                div {
                    &:last-child {
                        width: 100px;
                    }
                }
            }
        }

        .biz-cluster-node-overview-table {
            border-bottom: none;

            .name {
                width: 300px;
            }

            .mirror {
                width: 400px;
            }
        }
    }

</style>
