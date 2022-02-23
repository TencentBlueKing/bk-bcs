/**
 * @file 应用 container 详情页的 mixin
 */

import ECharts from 'vue-echarts/components/ECharts.vue'
import 'echarts/lib/chart/line'
import 'echarts/lib/component/tooltip'
import 'echarts/lib/component/legend'
import moment from 'moment'
import { Decimal } from 'decimal.js'

import { containerDetailChart } from '@/common/chart-option'
import { catchErrorHandler, formatBytes } from '@/common/util'

import { createChartOption } from '../container-chart-opts'

export default {
    props: {
        curProject: {
            type: Object
        }
    },
    components: {
        chart: ECharts
    },
    data () {
        return {
            containerInfo: {},

            cpuLine: containerDetailChart.cpu,
            containerCpuChartOpts: createChartOption(this),

            memLineInternal: containerDetailChart.memInternal,
            containerMemChartOptsInternal: createChartOption(this),

            memLine: containerDetailChart.mem,
            containerMemChartOpts: createChartOption(this),

            netLine: containerDetailChart.net,

            diskLineInternal: containerDetailChart.diskInternal,
            containerDiskChartOptsInternal: createChartOption(this),

            diskLine: containerDetailChart.disk,
            containerDiskChartOpts: createChartOption(this),

            tabActiveName: 'ports',
            portList: [],
            commandList: [],
            volumeList: [],
            envList: [],
            healthList: [],
            labelList: [],
            resourceList: [],
            contentLoading: false,
            bkMessageInstance: null,
            exceptionCode: null,
            envTabLoading: true,
            cpuToggleRangeStr: this.$t('1小时'),
            memToggleRangeStr: this.$t('1小时'),
            diskToggleRangeStr: this.$t('1小时')
        }
    },
    computed: {
        projectId () {
            return this.$route.params.projectId
        },
        projectCode () {
            return this.$route.params.projectCode
        },
        instanceId () {
            const instanceId = this.$route.params.instanceId === undefined
                ? 0
                : this.$route.params.instanceId
            return instanceId
        },
        taskgroupName () {
            return this.$route.params.taskgroupName
        },
        namespaceId () {
            return this.$route.params.namespaceId
        },
        containerId () {
            return this.$route.params.containerId
        },
        instanceName () {
            return this.$route.params.instanceName
        },
        instanceNamespace () {
            return this.$route.params.instanceNamespace
        },
        instanceCategory () {
            return this.$route.params.instanceCategory
        },
        searchParamsList () {
            return this.$route.params.searchParamsList
        },
        isEn () {
            return this.$store.state.isEn
        },
        clusterId () {
            return this.$route.query.cluster_id || ''
        }
    },
    mounted () {
        this.fetchContainerInfo()
    },
    destroyed () {
        this.bkMessageInstance && this.bkMessageInstance.close()
    },
    methods: {
        /**
         * 获取容器详情信息，上方数据和下方
         */
        async fetchContainerInfo () {
            this.contentLoading = true
            try {
                let url = ''
                const params = {
                    projectId: this.projectId,
                    instanceId: this.instanceId,
                    taskgroupName: this.taskgroupName,
                    containerId: this.containerId,
                    cluster_id: this.clusterId
                }

                if (String(this.instanceId) === '0') {
                    params.name = this.instanceName
                    params.namespace = this.instanceNamespace
                    params.category = this.instanceCategory
                }

                // k8s
                if (this.CATEGORY) {
                    params.category = this.CATEGORY
                    url = 'app/getContainerInfoK8s'
                } else {
                    url = 'app/getContainerInfoMesos'
                }

                const res = await this.$store.dispatch(url, params)
                this.containerInfo = Object.assign({}, res.data || {})

                const portList = this.containerInfo.ports || []
                this.portList.splice(0, this.portList.length, ...portList)

                const commandList = []
                const commands = this.containerInfo.commands || {}
                if (commands.command || commands.args) {
                    commandList.push(this.containerInfo.commands)
                }
                this.commandList.splice(0, this.commandList.length, ...commandList)

                const volumeList = this.containerInfo.volumes || []
                this.volumeList.splice(0, this.volumeList.length, ...volumeList)

                const envList = this.containerInfo.env_args || []
                this.envList.splice(0, this.envList.length, ...envList)

                // mesos，k8s 先隐藏健康检查
                if (!this.CATEGORY) {
                    const healthList = this.containerInfo.health_check || []
                    this.healthList.splice(0, this.healthList.length, ...healthList)
                }

                const labelList = this.containerInfo.labels || []
                this.labelList.splice(0, this.labelList.length, ...labelList)

                const resourceList = []
                const resources = this.containerInfo.resources || {}
                const requests = resources.requests || {}
                const cpuRequests = requests.cpu ? `requests: ${requests.cpu} | ` : ''
                const memRequests = requests.memory ? `requests: ${requests.memory} | ` : ''
                const limits = resources.limits || {}
                const cpuLimits = limits.cpu ? `limits: ${limits.cpu}` : ''
                const memLimits = limits.memory ? `limits: ${limits.memory}` : ''
                resourceList.push({
                    cpu: cpuRequests + cpuLimits,
                    memory: memRequests + memLimits
                })
                this.resourceList.splice(0, this.resourceList.length, ...resourceList)

                this.$refs.containerCpuLine && this.$refs.containerCpuLine.showLoading({
                    text: this.$t('正在加载'),
                    color: '#30d878',
                    maskColor: 'rgba(255, 255, 255, 0.8)'
                })
                this.$refs.containerMemLine && this.$refs.containerMemLine.showLoading({
                    text: this.$t('正在加载'),
                    color: '#30d878',
                    maskColor: 'rgba(255, 255, 255, 0.8)'
                })
                this.$refs.containerNetLine && this.$refs.containerNetLine.showLoading({
                    text: this.$t('正在加载'),
                    color: '#30d878',
                    maskColor: 'rgba(255, 255, 255, 0.8)'
                })
                this.$refs.containerDiskLine && this.$refs.containerDiskLine.showLoading({
                    text: this.$t('正在加载'),
                    color: '#30d878',
                    maskColor: 'rgba(255, 255, 255, 0.8)'
                })

                let cpuRange = '1'
                if (this.cpuToggleRangeStr === this.$t('24小时')) {
                    cpuRange = '2'
                }
                if (this.cpuToggleRangeStr === this.$t('近7天')) {
                    cpuRange = '3'
                }
                this.fetchContainerCpuUsage(cpuRange)

                let memRange = '1'
                if (this.memToggleRangeStr === this.$t('24小时')) {
                    memRange = '2'
                }
                if (this.memToggleRangeStr === this.$t('近7天')) {
                    memRange = '3'
                }
                this.fetchContainerMemUsage(memRange)

                let diskRange = '1'
                if (this.diskToggleRangeStr === this.$t('24小时')) {
                    diskRange = '2'
                }
                if (this.diskToggleRangeStr === this.$t('近7天')) {
                    diskRange = '3'
                }
                this.fetchContainerDisk(diskRange)
                // }
            } catch (e) {
                catchErrorHandler(e, this)
            } finally {
                this.contentLoading = false
            }
        },

        /**
         * 获取 容器CPU使用率
         *
         * @param {string} range 时间范围，1: 1 小时，2: 24 小时，3：近 7 天
         */
        async fetchContainerCpuUsage (range) {
            try {
                const params = {
                    projectId: this.projectId,
                    container_ids: this.containerId.split(','),
                    namespace: this.containerInfo.namespace,
                    clusterId: this.clusterId,
                    end_at: moment().format('YYYY-MM-DD HH:mm:ss')
                }

                // 1 小时
                if (range === '1') {
                    params.start_at = moment().subtract(1, 'hours').format('YYYY-MM-DD HH:mm:ss')
                } else if (range === '2') { // 24 小时
                    params.start_at = moment().subtract(1, 'days').format('YYYY-MM-DD HH:mm:ss')
                } else if (range === '3') { // 近 7 天
                    params.start_at = moment().subtract(7, 'days').format('YYYY-MM-DD HH:mm:ss')
                }

                const res = await this.$store.dispatch('app/containerCpuUsage', Object.assign({}, params))
                const limitRes = await this.$store.dispatch('app/containerCpuLimit', Object.assign({}, params))

                const limitData = limitRes.data.result || []
                const limitList = []
                limitData.forEach(item => {
                    limitList.push({
                        metric: item.metric,
                        val: parseFloat(item.value[1]) / 100000 * 100
                    })
                })
                this.renderContainerCpuChart(res.data.result || [], limitList)
            } catch (e) {
                catchErrorHandler(e, this)
            } finally {
                this.$refs.containerCpuLine && this.$refs.containerCpuLine.hideLoading()
            }
        },

        /**
         * 渲染 容器CPU使用率
         *
         * @param {Array} list 数据
         * @param {Array} limitList 红线数据
         */
        renderContainerCpuChart (list, limitList) {
            const chartNode = this.$refs.containerCpuLine
            if (!chartNode) {
                return
            }
            const containerCpuChartOpts = Object.assign({}, this.containerCpuChartOpts)
            containerCpuChartOpts.series.splice(0, containerCpuChartOpts.series.length, ...[])

            const data = list.length ? list : [{
                metric: { container_name: '--' },
                values: [[parseInt(String(+new Date()).slice(0, 10), 10), '10']]
            }]

            if (list.length) {
                containerCpuChartOpts.yAxis.splice(0, containerCpuChartOpts.yAxis.length, ...[
                    {
                        axisLabel: {
                            formatter (value, index) {
                                const valueLen = String(value).length
                                return `${Decimal(value).toPrecision(valueLen > 3 ? 3 : valueLen)}%`
                            }
                        }
                    }
                ])
            }

            const redLineData = []
            const hasRedLine = !!limitList.length

            data.forEach(item => {
                item.values.forEach(d => {
                    d[0] = parseInt(d[0] + '000', 10)
                    d.push(item.metric.container_name)
                    if (hasRedLine && list.length) {
                        limitList.forEach(limit => {
                            redLineData.push([d[0], limit.val, limit.metric.container_name])
                        })
                    }
                })
                containerCpuChartOpts.series.push(
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

            limitList.forEach(item => {
                containerCpuChartOpts.series.push({
                    type: 'line',
                    name: 'threshold',
                    smooth: true,
                    showSymbol: false,
                    hoverAnimation: false,
                    itemStyle: {
                        normal: {
                            color: 'red'
                        }
                    },
                    data: redLineData
                })
            })

            chartNode.mergeOptions({
                tooltip: {
                    formatter (params, ticket, callback) {
                        let ret

                        if (params[0].value[2] === '--') {
                            ret = '<div>No Data</div>'
                        } else {
                            let thresholdStr = ''
                            const valueLen0 = String(params[0].value[1]).length
                            if (params[1] && params[1].seriesName === 'threshold') {
                                const valueLen1 = String(params[1].value[1]).length
                                thresholdStr = `<div style="color: #fd9c9c;">Limit: ${Decimal(params[1].value[1]).toPrecision(valueLen1 > 3 ? 3 : valueLen1)}%</div>`
                            }
                            let date = params[0].value[0]
                            if (String(parseInt(date, 10)).length === 10) {
                                date = parseInt(date, 10) + '000'
                            }
                            ret = `
                                <div>${moment(parseInt(date, 10)).format('YYYY-MM-DD HH:mm:ss')}</div>
                                ${thresholdStr}
                                <div>${params[0].value[2]}: ${Decimal(params[0].value[1]).toPrecision(valueLen0 > 3 ? 3 : valueLen0)}%</div>
                            `
                        }

                        return ret
                    }
                }
            })

            chartNode.hideLoading()
        },

        /**
         * 获取 容器内存使用量
         *
         * @param {string} range 时间范围，1: 1 小时，2: 24 小时，3：近 7 天
         */
        async fetchContainerMemUsage (range) {
            try {
                const params = {
                    projectId: this.projectId,
                    container_ids: this.containerId.split(','),
                    namespace: this.containerInfo.namespace,
                    clusterId: this.clusterId,
                    end_at: moment().format('YYYY-MM-DD HH:mm:ss')
                }

                // 1 小时
                if (range === '1') {
                    params.start_at = moment().subtract(1, 'hours').format('YYYY-MM-DD HH:mm:ss')
                } else if (range === '2') { // 24 小时
                    params.start_at = moment().subtract(1, 'days').format('YYYY-MM-DD HH:mm:ss')
                } else if (range === '3') { // 近 7 天
                    params.start_at = moment().subtract(7, 'days').format('YYYY-MM-DD HH:mm:ss')
                }

                const res = await this.$store.dispatch('app/containerMemUsage', Object.assign({}, params))
                const limitRes = await this.$store.dispatch('app/containerMemLimit', Object.assign({}, params))

                const limitData = limitRes.data.result || []
                const limitList = []
                limitData.forEach(item => {
                    limitList.push({
                        metric: item.metric,
                        val: parseFloat(item.value[1])
                    })
                })
                this.renderContainerMemChart(res.data.result || [], limitList)
            } catch (e) {
                catchErrorHandler(e, this)
            } finally {
                this.$refs.containerMemLine && this.$refs.containerMemLine.hideLoading()
            }
        },

        /**
         * 渲染 容器内存使用量
         *
         * @param {Array} list 数据
         * @param {Array} limitList 红线数据
         */
        renderContainerMemChart (list, limitList) {
            const chartNode = this.$refs.containerMemLine
            if (!chartNode) {
                return
            }

            const chartOpts = Object.assign({}, this.containerMemChartOptsInternal)

            chartOpts.series.splice(0, chartOpts.series.length, ...[])

            const data = list.length ? list : [{
                metric: { container_name: '--' },
                values: [[parseInt(String(+new Date()).slice(0, 10), 10), '10']]
            }]

            if (list.length) {
                chartOpts.yAxis.splice(0, chartOpts.yAxis.length, ...[
                    {
                        axisLabel: {
                            formatter (value, index) {
                                return `${formatBytes(value)}`
                            }
                        }
                    }
                ])
            }

            const redLineData = []
            const hasRedLine = !!limitList.length

            data.forEach(item => {
                item.values.forEach(d => {
                    d[0] = parseInt(d[0] + '000', 10)
                    d.push(item.metric.container_name)
                    if (hasRedLine && list.length) {
                        limitList.forEach(limit => {
                            redLineData.push([d[0], limit.val, limit.metric.container_name])
                        })
                    }
                })
                chartOpts.series.push(
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

            limitList.forEach(item => {
                chartOpts.series.push({
                    type: 'line',
                    name: 'threshold',
                    smooth: true,
                    showSymbol: false,
                    hoverAnimation: false,
                    itemStyle: {
                        normal: {
                            color: 'red'
                        }
                    },
                    data: redLineData
                })
            })

            chartNode.mergeOptions({
                tooltip: {
                    formatter (params, ticket, callback) {
                        let ret

                        if (params[0].value[2] === '--') {
                            ret = '<div>No Data</div>'
                        } else {
                            let thresholdStr = ''
                            if (params[1] && params[1].seriesName === 'threshold') {
                                thresholdStr = `<div style="color: #fd9c9c;">Limit: ${formatBytes(params[1].value[1])}</div>`
                            }
                            let date = params[0].value[0]
                            if (String(parseInt(date, 10)).length === 10) {
                                date = parseInt(date, 10) + '000'
                            }
                            ret = `
                                <div>${moment(parseInt(date, 10)).format('YYYY-MM-DD HH:mm:ss')}</div>
                                ${thresholdStr}
                                <div>${params[0].value[2]}: ${formatBytes(params[0].value[1])}</div>
                            `
                        }

                        return ret
                    }
                }
            })

            chartNode.hideLoading()
        },

        /**
         * 获取 容器磁盘读写数据
         *
         * @param {string} range 时间范围，1: 1 小时，2: 24 小时，3：近 7 天
         */
        async fetchContainerDisk (range) {
            try {
                const params = {
                    projectId: this.projectId,
                    container_ids: this.containerId.split(','),
                    namespace: this.containerInfo.namespace,
                    clusterId: this.clusterId,
                    end_at: moment().format('YYYY-MM-DD HH:mm:ss')
                }

                // 1 小时
                if (range === '1') {
                    params.start_at = moment().subtract(1, 'hours').format('YYYY-MM-DD HH:mm:ss')
                } else if (range === '2') { // 24 小时
                    params.start_at = moment().subtract(1, 'days').format('YYYY-MM-DD HH:mm:ss')
                } else if (range === '3') { // 近 7 天
                    params.start_at = moment().subtract(7, 'days').format('YYYY-MM-DD HH:mm:ss')
                }

                const res = await Promise.all([
                    this.$store.dispatch('app/containerDiskWrite', Object.assign({}, params)),
                    this.$store.dispatch('app/containerDiskRead', Object.assign({}, params))
                ])
                this.renderContainerDiskChart(res[0].data.result, res[1].data.result)
            } catch (e) {
                catchErrorHandler(e, this)
            } finally {
                this.$refs.containerDiskLine && this.$refs.containerDiskLine.hideLoading()
            }
        },

        /**
         * 渲染 容器磁盘读写数据 图表
         *
         * @param {Array} listWrite 容器磁盘写数据
         * @param {Array} listRead 容器磁盘读数据
         */
        renderContainerDiskChart (listWrite, listRead) {
            const chartNode = this.$refs.containerDiskLine
            if (!chartNode) {
                return
            }

            const chartOpts = Object.assign({}, this.containerDiskChartOptsInternal)

            chartOpts.series.splice(0, chartOpts.series.length, ...[])

            chartOpts.yAxis.splice(0, chartOpts.yAxis.length, ...[
                {
                    axisLabel: {
                        formatter (value, index) {
                            return `${formatBytes(value)}`
                        }
                    }
                }
            ])

            const dataWrite = listWrite.length
                ? listWrite
                : [{
                    metric: { container_name: '--' },
                    values: [[parseInt(String(+new Date()).slice(0, 10), 10), '10']]
                }]

            const dataRead = listRead.length
                ? listRead
                : [{
                    metric: { container_name: '--' },
                    values: [[parseInt(String(+new Date()).slice(0, 10), 10), '10']]
                }]

            dataWrite.forEach(item => {
                item.values.forEach(d => {
                    d[0] = parseInt(d[0] + '000', 10)
                    d.push('write')
                    d.push(item.metric.container_name)
                })
                chartOpts.series.push(
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

            dataRead.forEach(item => {
                item.values.forEach(d => {
                    d[0] = parseInt(d[0] + '000', 10)
                    d.push('read')
                    d.push(item.metric.container_name)
                })
                chartOpts.series.push(
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

            const labelWrite = this.$t('磁盘写数据')
            const labelRead = this.$t('磁盘读数据')
            chartNode.mergeOptions({
                tooltip: {
                    formatter (params, ticket, callback) {
                        if (params[0].value[3] === '--') {
                            return '<div>No Data</div>'
                        }

                        let date = params[0].value[0]
                        if (String(parseInt(date, 10)).length === 10) {
                            date = parseInt(date, 10) + '000'
                        }

                        let ret = ''
                                + `<div>${moment(parseInt(date, 10)).format('YYYY-MM-DD HH:mm:ss')}</div>`

                        params.forEach(p => {
                            if (p.value[2] === 'write') {
                                ret += `<div>${p.value[3]}-${labelWrite}：${formatBytes(p.value[1])}</div>`
                            } else if (p.value[2] === 'read') {
                                ret += `<div>${p.value[3]}-${labelRead}：${formatBytes(p.value[1])}</div>`
                            }
                        })

                        return ret
                    }
                }
            })

            chartNode.hideLoading()
        },

        /**
         * 获取 cpu 图表数据
         */
        async fetchContainerMetricsCpu () {
            const ref = this.$refs.containerCpuLine
            try {
                const params = {
                    projectId: this.projectId,
                    containerId: this.containerId,
                    metric: 'cpu_summary',
                    cluster_id: this.clusterId
                }
                if (this.CATEGORY) {
                    params.category = this.CATEGORY
                }

                const res = await this.$store.dispatch('app/getContainerMetrics', params)

                setTimeout(() => {
                    this.setCpuData(
                        res.data.list && res.data.list.length
                            ? res.data.list
                            : [
                                {
                                    container_name: 'noData', usage: 0, time: new Date().getTime()
                                }
                            ]
                    )
                }, 0)
            } catch (e) {
                catchErrorHandler(e, this)
            } finally {
                ref && ref.hideLoading()
            }
        },

        /**
         * 设置 cpu 图表数据
         *
         * @param {Array} data 数据
         */
        setCpuData (data) {
            const chartData = []
            const emptyData = []
            const ref = this.$refs.containerCpuLine
            if (!ref) {
                return
            }

            data.forEach(item => {
                chartData.push({
                    value: [item.time, item.usage]
                })
                emptyData.push(0)
            })

            const name = this.containerInfo.container_name || data[0].container_name

            // 先设置 emptyData，再切换数据，这样的话，图表是从中间往两边展开，效果会好一些
            ref.mergeOptions({
                series: [
                    {
                        name: name,
                        data: emptyData
                    }
                ]
            })
            ref.mergeOptions({
                series: [
                    {
                        name: name,
                        data: chartData
                    }
                ]
            })
            ref.hideLoading()
        },

        /**
         * 获取 mem 图表数据
         *
         * @param {string} metric 标识是 cpu 还是内存图表
         */
        async fetchContainerMetricsMem (metric) {
            const ref = this.$refs.containerMemLine
            try {
                const params = {
                    projectId: this.projectId,
                    containerId: this.containerId,
                    metric: 'mem',
                    cluster_id: this.clusterId
                }
                if (this.CATEGORY) {
                    params.category = this.CATEGORY
                }

                const res = await this.$store.dispatch('app/getContainerMetrics', params)

                setTimeout(() => {
                    this.setMemDataInternal(
                        res.data.list && res.data.list.length
                            ? res.data.list
                            : [
                                {
                                    container_name: 'noData', rss_pct: 0, time: new Date().getTime()
                                }
                            ]
                    )
                }, 0)
            } catch (e) {
                catchErrorHandler(e, this)
            } finally {
                ref && ref.hideLoading()
            }
        },

        /**
         * 设置 mem 图表数据，内部版
         *
         * @param {Array} data 数据
         */
        setMemDataInternal (data) {
            const chartData = []
            const emptyData = []
            const ref = this.$refs.containerMemLine
            if (!ref) {
                return
            }

            data.forEach(item => {
                chartData.push({
                    value: [item.time, item.rss_pct]
                })
                emptyData.push(0)
            })

            const name = this.containerInfo.container_name || data[0].container_name

            // 先设置 emptyData，再切换数据，这样的话，图表是从中间往两边展开，效果会好一些
            ref.mergeOptions({
                series: [
                    {
                        name: name,
                        data: emptyData
                    }
                ]
            })
            ref.mergeOptions({
                series: [
                    {
                        name: name,
                        data: chartData
                    }
                ]
            })
            ref.hideLoading()
        },

        /**
         * 设置 mem 图表数据，非内部版
         *
         * @param {Array} data 数据
         */
        setMemData (data) {
            const chartData = []
            const emptyData = []
            const ref = this.$refs.containerMemLine
            if (!ref) {
                return
            }

            data.forEach(item => {
                chartData.push({
                    value: [item.time, item.used]
                })
                emptyData.push(0)
            })

            const name = this.containerInfo.container_name || data[0].container_name

            // 先设置 emptyData，再切换数据，这样的话，图表是从中间往两边展开，效果会好一些
            ref.mergeOptions({
                series: [
                    {
                        name: name,
                        data: emptyData
                    }
                ]
            })
            ref.mergeOptions({
                series: [
                    {
                        name: name,
                        data: chartData
                    }
                ]
            })
            ref.hideLoading()
        },

        /**
         * 获取 net 图表数据
         *
         * @param {string} metric 标识是 cpu 还是内存图表
         */
        async fetchContainerMetricsNet (metric) {
            const ref = this.$refs.containerNetLine
            try {
                const params = {
                    projectId: this.projectId,
                    containerId: this.containerId,
                    metric: 'net',
                    cluster_id: this.clusterId
                }
                if (this.CATEGORY) {
                    params.category = this.CATEGORY
                }

                const res = await this.$store.dispatch('app/getContainerMetrics', params)

                setTimeout(() => {
                    this.setNetData(
                        res.data.list && res.data.list.length
                            ? res.data.list
                            : [
                                {
                                    container_name: 'noData', txbytes: 0, rxbytes: 0, time: new Date().getTime()
                                }
                            ]
                    )
                }, 0)
            } catch (e) {
                catchErrorHandler(e, this)
            } finally {
                ref && ref.hideLoading()
            }
        },

        /**
         * 设置 net 图表数据
         *
         * @param {Array} data 数据
         */
        setNetData (data) {
            const emptyData = []
            const ref = this.$refs.containerNetLine
            if (!ref) {
                return
            }

            const charOpts = {
                legend: {
                    data: []
                },
                series: []
            }

            const name = this.containerInfo.container_name || data[0].container_name

            // 每秒发送的字节数
            const txbyteData = []

            // 每秒接收的字节数
            const rxbyteData = []

            data.forEach(item => {
                txbyteData.push({
                    value: [item.time, item.txbytes, 'tx', name]
                })
                rxbyteData.push({
                    value: [item.time, item.rxbytes, 'rx', name]
                })
                emptyData.push(0)
            })

            charOpts.legend.data.push(this.$t('发送'))
            charOpts.legend.data.push(this.$t('接收'))

            charOpts.series.push(
                {
                    type: 'line',
                    name: this.$t('发送'),
                    data: txbyteData
                },
                {
                    type: 'line',
                    name: this.$t('接收'),
                    data: rxbyteData
                }
            )

            // 先设置 emptyData，再切换数据，这样的话，图表是从中间往两边展开，效果会好一些
            ref.mergeOptions({
                series: [
                    {
                        data: emptyData
                    },
                    {
                        data: emptyData
                    }
                ]
            })

            ref.mergeOptions(charOpts)
            ref.hideLoading()
        },

        /**
         * 获取 disk 图表数据
         *
         * @param {string} metric 标识是 cpu 还是内存图表
         */
        async fetchContainerMetricsDisk (metric) {
            const ref = this.$refs.containerDiskLine
            try {
                const params = {
                    projectId: this.projectId,
                    containerId: this.containerId,
                    metric: 'disk'
                }
                if (this.CATEGORY) {
                    params.category = this.CATEGORY
                }

                const res = await this.$store.dispatch('app/getContainerMetrics', params)

                setTimeout(() => {
                    this.setDiskDataInternal(
                        res.data.list && res.data.list.length
                            ? res.data.list
                            : [
                                {
                                    device_name: 'noData', metrics: [{ used_pct: 0, time: new Date().getTime() }]
                                }
                            ]
                    )
                }, 0)
            } catch (e) {
                catchErrorHandler(e, this)
            } finally {
                ref && ref.hideLoading()
            }
        },

        /**
         * 设置 disk 图表数据，内部版
         *
         * @param {Array} data 数据
         */
        setDiskDataInternal (data) {
            const emptyData = []
            const ref = this.$refs.containerDiskLine
            if (!ref) {
                return
            }

            const charOpts = {
                legend: {
                    data: []
                },
                series: []
            }
            data.forEach(item => {
                const metrics = item.metrics
                const chartData = []
                metrics.forEach(metric => {
                    chartData.push({
                        value: [metric.time, metric.used_pct]
                    })
                    emptyData.push(0)
                })

                charOpts.series.push(
                    {
                        type: 'line',
                        name: item.device_name,
                        data: chartData
                    }
                )
            })

            // 先设置 emptyData，再切换数据，这样的话，图表是从中间往两边展开，效果会好一些
            ref.mergeOptions({
                series: [
                    {
                        data: emptyData
                    }
                ]
            })
            ref.mergeOptions(charOpts)
            ref.hideLoading()
        },

        /**
         * 设置 disk 图表数据，非内部版
         *
         * @param {Array} data 数据
         */
        setDiskData (data) {
            const emptyData = []
            const ref = this.$refs.containerDiskLine
            if (!ref) {
                return
            }

            const charOpts = {
                legend: {
                    data: []
                },
                series: []
            }

            data.forEach(item => {
                const metrics = item.metrics
                const chartData = []
                metrics.forEach(metric => {
                    chartData.push({
                        value: [metric.time, metric.read_bytes, metric.write_bytes, metric.container_name]
                    })
                    emptyData.push(0)
                })

                charOpts.series.push(
                    {
                        type: 'line',
                        name: item.device_name,
                        data: chartData
                    }
                )
            })

            // 先设置 emptyData，再切换数据，这样的话，图表是从中间往两边展开，效果会好一些
            ref.mergeOptions({
                series: [
                    {
                        data: emptyData
                    }
                ]
            })
            ref.mergeOptions(charOpts)
            ref.hideLoading()
        },

        /**
         * 选项卡切换事件
         *
         * @param {string} name 选项卡标识
         */
        async tabChanged (name) {
            if (this.tabActiveName === name) {
                return
            }
            this.tabActiveName = name

            if (this.CATEGORY && this.tabActiveName === 'env_args') {
                try {
                    this.envTabLoading = true
                    const params = {
                        projectId: this.projectId,
                        instanceId: this.instanceId,
                        taskgroupName: this.taskgroupName,
                        containerId: this.containerId,
                        cluster_id: this.clusterId
                    }

                    if (String(this.instanceId) === '0') {
                        params.name = this.instanceName
                        params.namespace = this.instanceNamespace
                        params.category = this.instanceCategory
                    }

                    // k8s
                    if (this.CATEGORY) {
                        params.category = this.CATEGORY
                    }

                    const res = await this.$store.dispatch('app/getEnvInfo', params)
                    const envList = []
                    envList.splice(0, 0, ...(res.data || []))
                    this.envList.splice(0, this.envList.length, ...envList)
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    this.envTabLoading = false
                }
            }
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
            let hook = ''
            if (idx === 'cpu') {
                ref = this.$refs.containerCpuLine
                hook = 'fetchContainerCpuUsage'
            }
            if (idx === 'mem') {
                ref = this.$refs.containerMemLine
                hook = 'fetchContainerMemUsage'
            }
            if (idx === 'disk') {
                ref = this.$refs.containerDiskLine
                hook = 'fetchContainerDisk'
            }
            ref && ref.showLoading({
                text: this.$t('正在加载中...'),
                color: '#30d878',
                maskColor: 'rgba(255, 255, 255, 0.8)'
            })

            this[hook](range)
        }
    }
}
