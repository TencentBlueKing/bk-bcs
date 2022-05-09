/**
 * @file 应用 instance 详情页的 mixin
 */

import moment from 'moment'
import ECharts from 'vue-echarts/components/ECharts.vue'
import Clipboard from 'clipboard'
import 'echarts/lib/chart/line'
import 'echarts/lib/component/tooltip'
import 'echarts/lib/component/legend'
import yamljs from 'js-yaml'
import { Decimal } from 'decimal.js'

import { instanceDetailChart } from '@/common/chart-option'
import { randomInt, catchErrorHandler, chartColors, formatBytes, copyText } from '@/common/util'
import ace from '@/components/ace-editor'
import BcsLog from '@/components/bcs-log/index'
import { createChartOption } from '../pod-chart-opts'

export default {
    // props: {
    //     curProject: {
    //         type: Object
    //     }
    // },
    components: {
        chart: ECharts,
        ace,
        BcsLog
    },
    data () {
        return {
            terminalWins: {},
            winHeight: 0,

            yAxisDefaultConf: {
                boundaryGap: [0, '2%'],
                type: 'value',
                axisLine: { show: true, lineStyle: { color: '#dde4eb' } },
                axisTick: { alignWithLabel: true, length: 0, lineStyle: { color: 'red' } },
                axisLabel: {
                    color: '#868b97',
                    formatter (value, index) {
                        return `${value.toFixed(1)}%`
                    }
                },
                splitLine: { show: true, lineStyle: { color: ['#ebf0f5'], type: 'dashed' } }
            },

            cpuLine: instanceDetailChart.cpu,
            podCpuChartOpts: createChartOption(this),
            podCpuChartOptsContainerView: createChartOption(this),

            memLineInternal: instanceDetailChart.memInternal,
            podMemChartOptsContainerView: instanceDetailChart.memInternal,

            podMemChartOptsInternal: createChartOption(this),
            podMemChartOptsInternalContainerView: createChartOption(this),

            memLine: instanceDetailChart.mem,
            podMemChartOpts: createChartOption(this),

            podNetChartOpts: createChartOption(this),
            podNetChartOptsContainerView: createChartOption(this),

            tabActiveName: 'taskgroup',
            instanceInfo: {},
            labelList: [],
            labelListLoading: true,
            annotationList: [],
            annotationListLoading: true,
            metricList: [],
            metricListLoading: true,
            metricListErrorMessage: this.$t('没有数据'),
            openTaskgroup: {},
            taskgroupList: [],
            taskgroupLoading: true,
            taskgroupTimer: null,
            openKeys: [],
            containerIdList: [],
            containerIdNameMap: {},
            eventList: [],
            eventListLoading: false,
            eventPageConf: {
                // 总数
                total: 0,
                // 总页数
                totalPage: 1,
                // 每页多少条
                pageSize: 5,
                // 当前页
                curPage: 1,
                // 是否显示翻页条
                show: false
            },
            logSideDialogConf: {
                isShow: false,
                title: '',
                timer: null,
                width: 820,
                showLogTime: false
            },
            toJsonDialogConf: {
                isShow: false,
                title: '',
                timer: null,
                width: 700,
                loading: false
            },
            logLoading: false,
            logList: [],
            bkMessageInstance: null,
            exceptionCode: null,
            editorConfig: {
                width: '100%',
                height: '100%',
                lang: 'yaml',
                readOnly: true,
                fullScreen: false,
                value: '',
                editor: null
            },
            clipboardInstance: null,
            copyContent: '',
            taskgroupInfoDialogConf: {
                isShow: false,
                title: '',
                timer: null,
                width: 690,
                loading: false
            },
            baseData: {},
            updateData: {},
            restartData: '',
            killData: '',
            instanceInfoLoading: true,
            reschedulerDialogConf: {
                isShow: false,
                width: 450,
                title: '',
                closeIcon: false,
                curRescheduler: null,
                curReschedulerIndex: -1
            },
            curChartView: 'pod',
            curSelectedPod: '',
            cpuToggleRangeStr: this.$t('1小时'),
            memToggleRangeStr: this.$t('1小时'),
            networkToggleRangeStr: this.$t('1小时'),
            cpuContainerToggleRangeStr: this.$t('1小时'),
            memContainerToggleRangeStr: this.$t('1小时'),
            diskContainerToggleRangeStr: this.$t('1小时'),
            bcsLog: {
                show: false,
                loading: false,
                containerList: [],
                podId: '',
                defaultContainer: ''
            }
        }
    },
    computed: {
        projectId () {
            return this.$route.params.projectId
        },
        curProject () {
            return this.$store.state.curProject
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
        namespaceId () {
            return this.$route.params.namespaceId
        },
        templateId () {
            return this.$route.params.templateId
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
    watch: {
        curSelectedPod (v) {
            if (!v) {
                return
            }
            this.$refs.instanceCpuLineContainerView && this.$refs.instanceCpuLineContainerView.showLoading({
                text: this.$t('正在加载'),
                color: '#30d878',
                maskColor: 'rgba(255, 255, 255, 0.8)'
            })
            this.$refs.instanceMemLineContainerView && this.$refs.instanceMemLineContainerView.showLoading({
                text: this.$t('正在加载'),
                color: '#30d878',
                maskColor: 'rgba(255, 255, 255, 0.8)'
            })
            this.$refs.instanceNetLineContainerView && this.$refs.instanceNetLineContainerView.showLoading({
                text: this.$t('正在加载'),
                color: '#30d878',
                maskColor: 'rgba(255, 255, 255, 0.8)'
            })

            let containerCpuRange = '1'
            if (this.cpuContainerToggleRangeStr === this.$t('24小时')) {
                containerCpuRange = '2'
            }
            if (this.cpuContainerToggleRangeStr === this.$t('近7天')) {
                containerCpuRange = '3'
            }
            this.fetchPodCpuUsageContainerView(containerCpuRange)

            let containerMemRange = '1'
            if (this.memContainerToggleRangeStr === this.$t('24小时')) {
                containerMemRange = '2'
            }
            if (this.memContainerToggleRangeStr === this.$t('近7天')) {
                containerMemRange = '3'
            }
            this.fetchPodMemUsageContainerView(containerMemRange)

            let containerDiskRange = '1'
            if (this.diskContainerToggleRangeStr === this.$t('24小时')) {
                containerDiskRange = '2'
            }
            if (this.diskContainerToggleRangeStr === this.$t('近7天')) {
                containerDiskRange = '3'
            }
            this.fetchDiskContainerView(containerDiskRange)
        }
    },
    async mounted () {
        await this.fetchInstanceInfo()
        this.fetchContainerIds()
        this.winHeight = window.innerHeight
        this.clipboardInstance = new Clipboard('.copy-code-btn')
        this.clipboardInstance.on('success', e => {
            this.$bkMessage({
                theme: 'success',
                message: this.$t('复制成功')
            })
        })
        window.addEventListener('resize', this.resizeHandler)
    },
    destroyed () {
        this.bkMessageInstance && this.bkMessageInstance.close()
        clearTimeout(this.taskgroupTimer)
        this.taskgroupTimer = null
        this.openTaskgroup = Object.assign({}, {})
        window.removeEventListener('resize', this.resizeHandler)
    },
    methods: {
        resizeHandler () {
            this.$refs.instanceCpuLine && this.$refs.instanceCpuLine.resize()
            this.$refs.instanceMemLine && this.$refs.instanceMemLine.resize()
            this.$refs.instanceNetLine && this.$refs.instanceNetLine.resize()
            this.$refs.instanceCpuLineContainerView && this.$refs.instanceCpuLineContainerView.resize()
            this.$refs.instanceMemLineContainerView && this.$refs.instanceMemLineContainerView.resize()
            this.$refs.instanceNetLineContainerView && this.$refs.instanceNetLineContainerView.resize()
        },

        /**
         * 分页大小更改
         *
         * @param {number} pageSize pageSize
         */
        changePageSize (pageSize) {
            this.eventPageConf.pageSize = pageSize
            this.eventPageConf.curPage = 1
            this.fetchEvent()
        },

        /**
         *  编辑器初始化之后的回调函数
         *  @param editor - 编辑器对象
         */
        editorInitAfter (editor) {
            this.editorConfig.editor = editor
            this.editorConfig.editor.setStyle('biz-app-container-tojson-ace')
        },

        /**
         * ace editor 全屏
         */
        setFullScreen () {
            this.editorConfig.fullScreen = true
        },

        /**
         * 取消全屏
         */
        cancelFullScreen () {
            this.editorConfig.fullScreen = false
        },

        /**
         * 关闭 to json
         *
         * @param {Object} cluster 当前集群对象
         */
        closeToJson () {
            this.toJsonDialogConf.isShow = false
            this.toJsonDialogConf.title = ''
            this.editorConfig.value = ''
            this.copyContent = ''
        },

        /**
         * to json, for mesos
         */
        async toJson () {
            try {
                this.toJsonDialogConf.isShow = true
                this.toJsonDialogConf.loading = true
                this.toJsonDialogConf.title = `${this.instanceInfo.name}.json`

                const params = {
                    projectId: this.projectId,
                    instanceId: this.instanceId,
                    cluster_id: this.clusterId
                }

                if (String(this.instanceId) === '0') {
                    params.name = this.instanceName
                    params.namespace = this.instanceNamespace
                    params.category = this.instanceCategory
                }

                if (this.CATEGORY) {
                    params.category = this.CATEGORY
                }

                const res = await this.$store.dispatch('app/toJson', params)
                setTimeout(() => {
                    this.toJsonDialogConf.loading = false
                    setTimeout(() => {
                        this.editorConfig.editor.gotoLine(0, 0, true)
                    }, 10)

                    const data = res.data || {}
                    if (Object.keys(data).length) {
                        this.editorConfig.value = JSON.stringify(res.data || {}, null, 4)
                    } else {
                        this.editorConfig.value = this.$t('配置为空')
                    }
                    this.copyContent = this.editorConfig.value
                }, 100)
            } catch (e) {
                console.error(e)
                catchErrorHandler(e, this)
                this.toJsonDialogConf.isShow = false
                this.toJsonDialogConf.loading = false
            }
        },

        /**
         * to yaml, for k8s
         */
        async toYaml () {
            try {
                this.toJsonDialogConf.isShow = true
                this.toJsonDialogConf.loading = true
                this.toJsonDialogConf.title = `${this.instanceInfo.name}.yaml`

                const params = {
                    projectId: this.projectId,
                    instanceId: this.instanceId,
                    cluster_id: this.clusterId
                }

                if (String(this.instanceId) === '0') {
                    params.name = this.instanceName
                    params.namespace = this.instanceNamespace
                    params.category = this.instanceCategory
                }

                if (this.CATEGORY) {
                    params.category = this.CATEGORY
                }

                const res = await this.$store.dispatch('app/toJson', params)
                setTimeout(() => {
                    this.toJsonDialogConf.loading = false
                    setTimeout(() => {
                        this.editorConfig.editor.gotoLine(0, 0, true)
                    }, 10)

                    const data = res.data || {}
                    if (Object.keys(data).length) {
                        this.editorConfig.value = yamljs.dump(res.data || {})
                    } else {
                        this.editorConfig.value = this.$t('配置为空')
                    }
                    this.copyContent = this.editorConfig.value
                }, 100)
            } catch (e) {
                console.error(e)
                catchErrorHandler(e, this)
                this.toJsonDialogConf.isShow = false
                this.toJsonDialogConf.loading = false
            }
        },

        /**
         * 获取实例详情信息，上方数据
         */
        async fetchInstanceInfo () {
            this.instanceInfoLoading = true
            const params = {
                projectId: this.projectId,
                instanceId: this.instanceId,
                cluster_id: this.clusterId
            }

            if (String(this.instanceId) === '0') {
                params.name = this.instanceName
                params.namespace = this.instanceNamespace
                params.category = this.instanceCategory
            }

            if (this.CATEGORY) {
                params.category = this.CATEGORY
            }
            this.$refs.instanceCpuLine && this.$refs.instanceCpuLine.showLoading({
                text: this.$t('正在加载'),
                color: '#30d878',
                maskColor: 'rgba(255, 255, 255, 0.8)'
            })
            this.$refs.instanceMemLine && this.$refs.instanceMemLine.showLoading({
                text: this.$t('正在加载'),
                color: '#30d878',
                maskColor: 'rgba(255, 255, 255, 0.8)'
            })
            this.$refs.instanceNetLine && this.$refs.instanceNetLine.showLoading({
                text: this.$t('正在加载'),
                color: '#30d878',
                maskColor: 'rgba(255, 255, 255, 0.8)'
            })

            try {
                const res = await this.$store.dispatch('app/getInstanceInfo', params)

                this.instanceInfo = Object.assign({}, res.data || {})

                const createTimeMoment = moment(this.instanceInfo.create_time)
                this.instanceInfo.createTime = createTimeMoment.isValid()
                    ? createTimeMoment.format('YYYY-MM-DD HH:mm:ss')
                    : '--'

                const updateTimeMoment = moment(this.instanceInfo.update_time)
                this.instanceInfo.updateTime = updateTimeMoment.isValid()
                    ? updateTimeMoment.format('YYYY-MM-DD HH:mm:ss')
                    : '--'

                await this.fetchTaskgroup(true)
            } catch (e) {
                catchErrorHandler(e, this)
            } finally {
                this.instanceInfoLoading = false
            }
        },

        /**
         * 获取下方 tab taskgroup 的数据
         *
         * @param {boolean} isLoadContainerMetrics 是否需要加载图表
         */
        async fetchTaskgroup (isLoadContainerMetrics) {
            this.taskgroupLoading = true
            try {
                const params = {
                    projectId: this.projectId,
                    instanceId: this.instanceId,
                    cluster_id: this.clusterId
                }

                if (String(this.instanceId) === '0') {
                    params.name = this.instanceName
                    params.namespace = this.instanceNamespace
                    params.category = this.instanceCategory
                }

                if (this.CATEGORY) {
                    params.category = this.CATEGORY
                }

                const res = await this.$store.dispatch('app/getTaskgroupList', params)

                this.taskgroupList.splice(0, this.taskgroupList.length, ...[])

                const list = res.data || []
                list.forEach(item => {
                    let diffStr = ''
                    if (item.current_time && item.start_time) {
                        const timeDiff = moment.duration(moment(item.current_time).diff(moment(item.start_time)))
                        const arr = [
                            moment(item.current_time).diff(moment(item.start_time), 'days'),
                            timeDiff.get('hour'),
                            timeDiff.get('minute'),
                            timeDiff.get('second')
                        ]
                        diffStr = (arr[0] !== 0 ? (arr[0] + this.$t('天1')) : '')
                            + (arr[1] !== 0 ? (arr[1] + this.$t('小时1')) : '')
                            + (arr[2] !== 0 ? (arr[2] + this.$t('分1')) : '')
                            + (arr[3] !== 0 ? (arr[3] + this.$t('秒1')) : '')
                    }

                    this.taskgroupList.push({
                        ...item,
                        isOpen: false,
                        containerList: [],
                        containerLoading: false,
                        surviveTime: diffStr
                    })
                })
                clearTimeout(this.taskgroupTimer)
                this.taskgroupTimer = null
                this.taskgroupTimer = setTimeout(() => {
                    this.loopTaskgroup()
                }, 5000)
            } catch (e) {
                catchErrorHandler(e, this)
            } finally {
                this.taskgroupLoading = false
            }
        },

        /**
         * 获取所有的 container id
         */
        async fetchContainerIds () {
            try {
                const params = {
                    projectId: this.projectId,
                    instanceId: this.instanceId,
                    cluster_id: this.clusterId
                }

                if (String(this.instanceId) === '0') {
                    params.name = this.instanceName
                    params.namespace = this.instanceNamespace
                    params.category = this.instanceCategory
                }

                if (this.CATEGORY) {
                    params.category = this.CATEGORY
                }

                const res = await this.$store.dispatch('app/getContainerIds', params)
                const containerIdNameMap = {}
                const containerIdList = []
                const list = res.data || []
                // const list = [{container_id: '2aa0d5444531fe63f56da621f9f254596584c6338383f0370d678d8297d4af23', 'container_name': 'container_name111'}]
                list.forEach(item => {
                    containerIdNameMap[item.container_id] = item.container_name
                    containerIdList.push(item.container_id)
                })

                this.containerIdNameMap = JSON.parse(JSON.stringify(containerIdNameMap))
                this.containerIdList.splice(0, this.containerIdList.length, ...containerIdList)

                this.fetchPodCpuUsage('1')
                this.fetchPodMemUsage('1')
                this.fetchPodNet('1')
                // }
            } catch (e) {
                catchErrorHandler(e, this)
            } finally {
                // this.taskgroupLoading = false
            }
        },

        /**
         * 获取 POD CPU使用率
         *
         * @param {string} range 时间范围，1: 1 小时，2: 24 小时，3：近 7 天
         */
        async fetchPodCpuUsage (range) {
            const idList = this.taskgroupList.map(item => item.name)
            if (!idList || !idList.length) {
                this.renderPodCpuChart([])
                return
            }

            try {
                const params = {
                    projectId: this.projectId,
                    data: {
                        pod_name_list: idList,
                        namespace: this.instanceInfo.namespace_name,
                        end_at: moment().format('YYYY-MM-DD HH:mm:ss')
                    },
                    clusterId: this.clusterId
                }

                // 1 小时
                if (range === '1') {
                    params.data.start_at = moment().subtract(1, 'hours').format('YYYY-MM-DD HH:mm:ss')
                } else if (range === '2') { // 24 小时
                    params.data.start_at = moment().subtract(1, 'days').format('YYYY-MM-DD HH:mm:ss')
                } else if (range === '3') { // 近 7 天
                    params.data.start_at = moment().subtract(7, 'days').format('YYYY-MM-DD HH:mm:ss')
                }

                const res = await this.$store.dispatch('app/podCpuUsage', params)
                this.renderPodCpuChart(res.data.result || [])
            } catch (e) {
                catchErrorHandler(e, this)
            } finally {
                this.$refs.instanceCpuLine && this.$refs.instanceCpuLine.hideLoading()
            }
        },

        /**
         * 渲染 pod cpu 图表
         *
         * @param {Array} list 数据
         */
        renderPodCpuChart (list) {
            const chartNode = this.$refs.instanceCpuLine
            if (!chartNode) {
                return
            }

            const podCpuChartOpts = Object.assign({}, this.podCpuChartOpts)
            podCpuChartOpts.series.splice(0, podCpuChartOpts.series.length, ...[])

            const data = list.length ? list : [{
                metric: { pod_name: '--' },
                values: [[parseInt(String(+new Date()).slice(0, 10), 10), '10']]
            }]

            if (list.length) {
                podCpuChartOpts.yAxis.splice(0, podCpuChartOpts.yAxis.length, ...[
                    {
                        ...this.yAxisDefaultConf,
                        axisLabel: {
                            color: '#868b97',
                            formatter (value, index) {
                                const valueLen = String(value).length
                                return `${Decimal(value).toPrecision(valueLen > 3 ? 3 : valueLen)}%`
                            }
                        }
                    }
                ])
            }

            data.forEach(item => {
                item.values.forEach(d => {
                    d[0] = parseInt(d[0] + '000', 10)
                    d.push(item.metric.pod_name)
                })
                podCpuChartOpts.series.push(
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

            podCpuChartOpts.tooltip.formatter = (params, ticket, callback) => {
                let ret = ''
                if (params.every(param => param.value[2] === '--')) {
                    ret = '<div>No Data</div>'
                } else {
                    let date = params[0].value[0]
                    if (String(parseInt(date, 10)).length === 10) {
                        date = parseInt(date, 10) + '000'
                    }

                    ret += `<div>${moment(parseInt(date, 10)).format('YYYY-MM-DD HH:mm:ss')}</div>`
                    params.forEach(p => {
                        // ret += `<div>${p.value[2]}：${parseFloat(p.value[1]).toPrecision(3)}%</div>`
                        const valueLen = String(p.value[1]).length
                        ret += `<div>${p.value[2]}：${Decimal(p.value[1]).toPrecision(valueLen > 3 ? 3 : valueLen)}%</div>`
                    })
                }

                return ret
            }

            // chartNode.mergeOptions({
            //     tooltip: {
            //         formatter (params, ticket, callback) {
            //             let ret = ''
            //             if (params.every(param => param.value[2] === '--')) {
            //                 ret = '<div>No Data</div>'
            //             } else {
            //                 let date = params[0].value[0]
            //                 if (String(parseInt(date, 10)).length === 10) {
            //                     date = parseInt(date, 10) + '000'
            //                 }

            //                 ret += `<div>${moment(parseInt(date, 10)).format('YYYY-MM-DD HH:mm:ss')}</div>`
            //                 params.forEach(p => {
            //                     // ret += `<div>${p.value[2]}：${parseFloat(p.value[1]).toPrecision(3)}%</div>`
            //                     const valueLen = String(p.value[1]).length
            //                     ret += `<div>${p.value[2]}：${Decimal(p.value[1]).toPrecision(valueLen > 3 ? 3 : valueLen)}%</div>`
            //                 })
            //             }

            //             return ret
            //         }
            //     }
            // })

            chartNode.hideLoading()
        },

        /**
         * 获取 POD 内存使用量
         *
         * @param {string} range 时间范围，1: 1 小时，2: 24 小时，3：近 7 天
         */
        async fetchPodMemUsage (range) {
            const idList = this.taskgroupList.map(item => item.name)
            if (!idList || !idList.length) {
                this.renderPodMemChart([])
                return
            }
            try {
                const params = {
                    projectId: this.projectId,
                    data: {
                        pod_name_list: idList,
                        namespace: this.instanceInfo.namespace_name,
                        end_at: moment().format('YYYY-MM-DD HH:mm:ss')
                    },
                    clusterId: this.clusterId
                }

                // 1 小时
                if (range === '1') {
                    params.data.start_at = moment().subtract(1, 'hours').format('YYYY-MM-DD HH:mm:ss')
                } else if (range === '2') { // 24 小时
                    params.data.start_at = moment().subtract(1, 'days').format('YYYY-MM-DD HH:mm:ss')
                } else if (range === '3') { // 近 7 天
                    params.data.start_at = moment().subtract(7, 'days').format('YYYY-MM-DD HH:mm:ss')
                }

                const res = await this.$store.dispatch('app/podMemUsage', params)
                this.renderPodMemChart(res.data.result || [])
            } catch (e) {
                catchErrorHandler(e, this)
            } finally {
                this.$refs.instanceMemLine && this.$refs.instanceMemLine.hideLoading()
            }
        },

        /**
         * 渲染 pod 内存使用量
         *
         * @param {Array} list 数据
         */
        renderPodMemChart (list) {
            const chartNode = this.$refs.instanceMemLine
            if (!chartNode) {
                return
            }
            const chartOpts = Object.assign({}, this.podMemChartOptsInternal)

            chartOpts.series.splice(0, chartOpts.series.length, ...[])

            const data = list.length ? list : [{
                metric: { pod_name: '--' },
                values: [[parseInt(String(+new Date()).slice(0, 10), 10), '10']]
            }]

            if (list.length) {
                chartOpts.yAxis.splice(0, chartOpts.yAxis.length, ...[
                    {
                        ...this.yAxisDefaultConf,
                        axisLabel: {
                            color: '#868b97',
                            formatter (value, index) {
                                return `${formatBytes(value)}`
                            }
                        }
                    }
                ])
            }

            data.forEach(item => {
                item.values.forEach(d => {
                    d[0] = parseInt(d[0] + '000', 10)
                    d.push(item.metric.pod_name)
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

            chartOpts.tooltip.formatter = (params, ticket, callback) => {
                let ret = ''
                if (params.every(param => param.value[2] === '--')) {
                    ret = '<div>No Data</div>'
                } else {
                    let date = params[0].value[0]
                    if (String(parseInt(date, 10)).length === 10) {
                        date = parseInt(date, 10) + '000'
                    }
                    ret += `<div>${moment(parseInt(date, 10)).format('YYYY-MM-DD HH:mm:ss')}</div>`
                    params.forEach(p => {
                        ret += `<div>${p.value[2]}：${formatBytes(p.value[1])}</div>`
                    })
                }

                return ret
            }

            // chartNode.mergeOptions({
            //     tooltip: {
            //         formatter (params, ticket, callback) {
            //             let ret = ''
            //             if (params.every(param => param.value[2] === '--')) {
            //                 ret = '<div>No Data</div>'
            //             } else {
            //                 let date = params[0].value[0]
            //                 if (String(parseInt(date, 10)).length === 10) {
            //                     date = parseInt(date, 10) + '000'
            //                 }
            //                 ret += `<div>${moment(parseInt(date, 10)).format('YYYY-MM-DD HH:mm:ss')}</div>`
            //                 params.forEach(p => {
            //                     ret += `<div>${p.value[2]}：${formatBytes(p.value[1])}</div>`
            //                 })
            //             }

            //             return ret
            //         }
            //     }
            // })

            chartNode.hideLoading()
        },

        /**
         * 获取 POD网络接收发送数据
         *
         * @param {string} range 时间范围，1: 1 小时，2: 24 小时，3：近 7 天
         */
        async fetchPodNet (range) {
            const idList = this.taskgroupList.map(item => item.name)
            if (!idList || !idList.length) {
                this.renderPodNetChart([], [])
                return
            }
            try {
                const params = {
                    projectId: this.projectId,
                    data: {
                        pod_name_list: idList,
                        namespace: this.instanceInfo.namespace_name,
                        end_at: moment().format('YYYY-MM-DD HH:mm:ss')
                    },
                    clusterId: this.clusterId
                }

                // 1 小时
                if (range === '1') {
                    params.data.start_at = moment().subtract(1, 'hours').format('YYYY-MM-DD HH:mm:ss')
                } else if (range === '2') { // 24 小时
                    params.data.start_at = moment().subtract(1, 'days').format('YYYY-MM-DD HH:mm:ss')
                } else if (range === '3') { // 近 7 天
                    params.data.start_at = moment().subtract(7, 'days').format('YYYY-MM-DD HH:mm:ss')
                }

                const res = await Promise.all([
                    this.$store.dispatch('app/podNetReceive', Object.assign({}, params)),
                    this.$store.dispatch('app/podNetTransmit', Object.assign({}, params))
                ])
                if (res[0] && res[1]) {
                    this.renderPodNetChart(res[0].data.result, res[1].data.result)
                }
            } catch (e) {
                catchErrorHandler(e, this)
            } finally {
                this.$refs.instanceNetLine && this.$refs.instanceNetLine.hideLoading()
            }
        },

        /**
         * 渲染 Pod网络接收发送数据 图表
         *
         * @param {Array} listReceive net 入流量数据
         * @param {Array} listTransmit net 出流量数据
         */
        renderPodNetChart (listReceive, listTransmit) {
            const chartNode = this.$refs.instanceNetLine
            if (!chartNode) {
                return
            }

            const podNetChartOpts = Object.assign({}, this.podNetChartOpts)
            podNetChartOpts.series.splice(0, podNetChartOpts.series.length, ...[])

            podNetChartOpts.yAxis.splice(0, podNetChartOpts.yAxis.length, ...[
                {
                    ...this.yAxisDefaultConf,
                    axisLabel: {
                        color: '#868b97',
                        formatter (value, index) {
                            return `${formatBytes(value)}`
                        }
                    }
                }
            ])

            const dataReceive = listReceive.length
                ? listReceive
                : [{
                    metric: { pod_name: '--' },
                    values: [[parseInt(String(+new Date()).slice(0, 10), 10), '10']]
                }]

            const dataTransmit = listTransmit.length
                ? listTransmit
                : [{
                    metric: { pod_name: '--' },
                    values: [[parseInt(String(+new Date()).slice(0, 10), 10), '10']]
                }]

            dataReceive.forEach(item => {
                item.values.forEach(d => {
                    d[0] = parseInt(d[0] + '000', 10)
                    d.push('receive')
                    d.push(item.metric.pod_name)
                })
                podNetChartOpts.series.push(
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
                    d.push(item.metric.pod_name)
                })
                podNetChartOpts.series.push(
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

            const labelReceive = this.$t('入流量')
            const labelTransmit = this.$t('出流量')

            podNetChartOpts.tooltip.formatter = (params, ticket, callback) => {
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
                    if (p.value[2] === 'receive') {
                        ret += `<div>${p.value[3]}-${labelReceive}：${formatBytes(p.value[1])}</div>`
                    } else if (p.value[2] === 'transmit') {
                        ret += `<div>${p.value[3]}-${labelTransmit}：${formatBytes(p.value[1])}</div>`
                    }
                })

                return ret
            }

            // chartNode.mergeOptions({
            //     tooltip: {
            //         formatter (params, ticket, callback) {
            //             if (params[0].value[3] === '--') {
            //                 return '<div>No Data</div>'
            //             }

            //             let date = params[0].value[0]
            //             if (String(parseInt(date, 10)).length === 10) {
            //                 date = parseInt(date, 10) + '000'
            //             }

            //             let ret = ''
            //                     + `<div>${moment(parseInt(date, 10)).format('YYYY-MM-DD HH:mm:ss')}</div>`

            //             params.forEach(p => {
            //                 if (p.value[2] === 'receive') {
            //                     ret += `<div>${p.value[3]}-${labelReceive}：${formatBytes(p.value[1])}</div>`
            //                 } else if (p.value[2] === 'transmit') {
            //                     ret += `<div>${p.value[3]}-${labelTransmit}：${formatBytes(p.value[1])}</div>`
            //                 }
            //             })

            //             return ret
            //         }
            //     }
            // })

            chartNode.hideLoading()
        },

        /**
         * 切换 pod 视图下拉框改变事件、容器视图
         *
         * @param {string} paramName 选择的 pod name
         */
        changeSelectedPod (podName) {
            this.curSelectedPod = podName
        },

        /**
         * 切换 pod 视图、容器视图
         *
         * @param {string} idx 视图标识
         */
        async toggleView (idx) {
            if (this.curChartView === idx) {
                return
            }
            this.curChartView = idx
            // pod 视图
            if (this.curChartView === 'pod') {
                this.$refs.instanceCpuLine && this.$refs.instanceCpuLine.showLoading({
                    text: this.$t('正在加载'),
                    color: '#30d878',
                    maskColor: 'rgba(255, 255, 255, 0.8)'
                })
                this.$refs.instanceMemLine && this.$refs.instanceMemLine.showLoading({
                    text: this.$t('正在加载'),
                    color: '#30d878',
                    maskColor: 'rgba(255, 255, 255, 0.8)'
                })
                this.$refs.instanceNetLine && this.$refs.instanceNetLine.showLoading({
                    text: this.$t('正在加载'),
                    color: '#30d878',
                    maskColor: 'rgba(255, 255, 255, 0.8)'
                })

                let podCpuRange = '1'
                if (this.cpuToggleRangeStr === this.$t('24小时')) {
                    podCpuRange = '2'
                }
                if (this.cpuToggleRangeStr === this.$t('近7天')) {
                    podCpuRange = '3'
                }
                this.fetchPodCpuUsage(podCpuRange)

                let podMemRange = '1'
                if (this.memToggleRangeStr === this.$t('24小时')) {
                    podMemRange = '2'
                }
                if (this.memToggleRangeStr === this.$t('近7天')) {
                    podMemRange = '3'
                }
                this.fetchPodMemUsage(podMemRange)

                let podNetRange = '1'
                if (this.networkToggleRangeStr === this.$t('24小时')) {
                    podNetRange = '2'
                }
                if (this.networkToggleRangeStr === this.$t('近7天')) {
                    podNetRange = '3'
                }
                this.fetchPodNet(podNetRange)
                this.curSelectedPod = ''
            } else { // 容器视图
                if (this.taskgroupList.length) {
                    this.curSelectedPod = this.taskgroupList[0].name
                } else {
                    this.curSelectedPod = 'null'
                }
            }
        },

        /**
         * 获取 POD CPU使用率 容器视图
         *
         * @param {string} range 时间范围，1: 1 小时，2: 24 小时，3：近 7 天
         */
        async fetchPodCpuUsageContainerView (range) {
            if (this.curSelectedPod === 'null') {
                this.renderPodCpuChartContainerView([])
                return
            }
            try {
                const params = {
                    projectId: this.projectId,
                    pod_name: this.curSelectedPod,
                    clusterId: this.clusterId,
                    end_at: moment().format('YYYY-MM-DD HH:mm:ss'),
                    namespace: this.instanceInfo.namespace_name
                }

                // 1 小时
                if (range === '1') {
                    params.start_at = moment().subtract(1, 'hours').format('YYYY-MM-DD HH:mm:ss')
                } else if (range === '2') { // 24 小时
                    params.start_at = moment().subtract(1, 'days').format('YYYY-MM-DD HH:mm:ss')
                } else if (range === '3') { // 近 7 天
                    params.start_at = moment().subtract(7, 'days').format('YYYY-MM-DD HH:mm:ss')
                }

                const res = await this.$store.dispatch('app/podCpuUsageContainerView', params)
                this.renderPodCpuChartContainerView(res.data.result || [])
            } catch (e) {
                catchErrorHandler(e, this)
            } finally {
                this.$refs.instanceCpuLineContainerView && this.$refs.instanceCpuLineContainerView.hideLoading()
            }
        },

        /**
         * 渲染 pod cpu 图表
         *
         * @param {Array} list 数据
         */
        renderPodCpuChartContainerView (list) {
            const chartNode = this.$refs.instanceCpuLineContainerView
            if (!chartNode) {
                return
            }

            const podCpuChartOptsContainerView = Object.assign({}, this.podCpuChartOptsContainerView)
            podCpuChartOptsContainerView.series.splice(0, podCpuChartOptsContainerView.series.length, ...[])

            const data = list.length ? list : [{
                metric: { container_name: '--' },
                values: [[parseInt(String(+new Date()).slice(0, 10), 10), '10']]
            }]

            if (list.length) {
                podCpuChartOptsContainerView.yAxis.splice(0, podCpuChartOptsContainerView.yAxis.length, ...[
                    {
                        ...this.yAxisDefaultConf,
                        axisLabel: {
                            color: '#868b97',
                            formatter (value, index) {
                                const valueLen = String(value).length
                                return `${Decimal(value).toPrecision(valueLen > 3 ? 3 : valueLen)}%`
                            }
                        }
                    }
                ])
            }

            data.forEach(item => {
                item.values.forEach(d => {
                    d[0] = parseInt(d[0] + '000', 10)
                    d.push(item.metric.container_name)
                })
                podCpuChartOptsContainerView.series.push(
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

            podCpuChartOptsContainerView.tooltip.formatter = (params, ticket, callback) => {
                let ret = ''
                if (params.every(param => param.value[2] === '--')) {
                    ret = '<div>No Data</div>'
                } else {
                    let date = params[0].value[0]
                    if (String(parseInt(date, 10)).length === 10) {
                        date = parseInt(date, 10) + '000'
                    }
                    ret += `<div>${moment(parseInt(date, 10)).format('YYYY-MM-DD HH:mm:ss')}</div>`
                    params.forEach(p => {
                        // ret += `<div>${p.value[2]}：${parseFloat(p.value[1]).toPrecision(3)}%</div>`
                        const valueLen = String(p.value[1]).length
                        ret += `<div>${p.value[2]}：${Decimal(p.value[1]).toPrecision(valueLen > 3 ? 3 : valueLen)}%</div>`
                    })
                }

                return ret
            }

            // chartNode.mergeOptions({
            //     tooltip: {
            //         formatter (params, ticket, callback) {
            //             console.error(params)
            //             let ret = ''
            //             if (params.every(param => param.value[2] === '--')) {
            //                 ret = '<div>No Data</div>'
            //             } else {
            //                 let date = params[0].value[0]
            //                 if (String(parseInt(date, 10)).length === 10) {
            //                     date = parseInt(date, 10) + '000'
            //                 }
            //                 ret += `<div>${moment(parseInt(date, 10)).format('YYYY-MM-DD HH:mm:ss')}</div>`
            //                 params.forEach(p => {
            //                     // ret += `<div>${p.value[2]}：${parseFloat(p.value[1]).toPrecision(3)}%</div>`
            //                     const valueLen = String(p.value[1]).length
            //                     ret += `<div>${p.value[2]}：${Decimal(p.value[1]).toPrecision(valueLen > 3 ? 3 : valueLen)}%</div>`
            //                 })
            //             }

            //             return ret
            //         }
            //     }
            // })

            chartNode.hideLoading()
        },

        /**
         * 获取 内存使用量 容器视图
         *
         * @param {string} range 时间范围，1: 1 小时，2: 24 小时，3：近 7 天
         */
        async fetchPodMemUsageContainerView (range) {
            if (this.curSelectedPod === 'null') {
                this.renderPodMemChartContainerView([])
                return
            }
            try {
                const params = {
                    projectId: this.projectId,
                    pod_name: this.curSelectedPod,
                    clusterId: this.clusterId,
                    end_at: moment().format('YYYY-MM-DD HH:mm:ss'),
                    namespace: this.instanceInfo.namespace_name
                }

                // 1 小时
                if (range === '1') {
                    params.start_at = moment().subtract(1, 'hours').format('YYYY-MM-DD HH:mm:ss')
                } else if (range === '2') { // 24 小时
                    params.start_at = moment().subtract(1, 'days').format('YYYY-MM-DD HH:mm:ss')
                } else if (range === '3') { // 近 7 天
                    params.start_at = moment().subtract(7, 'days').format('YYYY-MM-DD HH:mm:ss')
                }

                const res = await this.$store.dispatch('app/podMemUsageContainerView', params)
                this.renderPodMemChartContainerView(res.data.result || [])
            } catch (e) {
                catchErrorHandler(e, this)
            } finally {
                this.$refs.instanceMemLineContainerView && this.$refs.instanceMemLineContainerView.hideLoading()
            }
        },

        /**
         * 渲染 pod 内存使用量 容器视图
         *
         * @param {Array} list 数据
         */
        renderPodMemChartContainerView (list) {
            const chartNode = this.$refs.instanceMemLineContainerView
            if (!chartNode) {
                return
            }

            const chartOpts = Object.assign({}, this.podMemChartOptsInternalContainerView)

            chartOpts.series.splice(0, chartOpts.series.length, ...[])

            const data = list.length ? list : [{
                metric: { container_name: '--' },
                values: [[parseInt(String(+new Date()).slice(0, 10), 10), '10']]
            }]

            if (list.length) {
                chartOpts.yAxis.splice(0, chartOpts.yAxis.length, ...[
                    {
                        ...this.yAxisDefaultConf,
                        axisLabel: {
                            color: '#868b97',
                            formatter (value, index) {
                                return `${formatBytes(value)}`
                            }
                        }
                    }
                ])
            }

            data.forEach(item => {
                item.values.forEach(d => {
                    d[0] = parseInt(d[0] + '000', 10)
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
                                color: '#3a84ff'
                            }
                        },
                        data: item.values
                    }
                )
            })

            chartOpts.tooltip.formatter = (params, ticket, callback) => {
                let ret = ''
                if (params.every(param => param.value[2] === '--')) {
                    ret = '<div>No Data</div>'
                } else {
                    let date = params[0].value[0]
                    if (String(parseInt(date, 10)).length === 10) {
                        date = parseInt(date, 10) + '000'
                    }
                    ret += `<div>${moment(parseInt(date, 10)).format('YYYY-MM-DD HH:mm:ss')}</div>`
                    params.forEach(p => {
                        ret += `<div>${p.value[2]}：${formatBytes(p.value[1])}</div>`
                    })
                }

                return ret
            }

            // chartNode.mergeOptions({
            //     tooltip: {
            //         formatter (params, ticket, callback) {
            //             let ret = ''
            //             if (params.every(param => param.value[2] === '--')) {
            //                 ret = '<div>No Data</div>'
            //             } else {
            //                 let date = params[0].value[0]
            //                 if (String(parseInt(date, 10)).length === 10) {
            //                     date = parseInt(date, 10) + '000'
            //                 }
            //                 ret += `<div>${moment(parseInt(date, 10)).format('YYYY-MM-DD HH:mm:ss')}</div>`
            //                 params.forEach(p => {
            //                     ret += `<div>${p.value[2]}：${formatBytes(p.value[1])}</div>`
            //                 })
            //             }

            //             return ret
            //         }
            //     }
            // })

            chartNode.hideLoading()
        },

        /**
         * 获取 容器磁盘读写数据 容器视图
         *
         * @param {string} range 时间范围，1: 1 小时，2: 24 小时，3：近 7 天
         */
        async fetchDiskContainerView (range) {
            if (this.curSelectedPod === 'null') {
                this.renderDiskChartContainerView([], [])
                return
            }
            try {
                const params = {
                    projectId: this.projectId,
                    pod_name: this.curSelectedPod,
                    clusterId: this.clusterId,
                    end_at: moment().format('YYYY-MM-DD HH:mm:ss'),
                    namespace: this.instanceInfo.namespace_name
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
                    this.$store.dispatch('app/podDiskWriteContainerView', Object.assign({}, params)),
                    this.$store.dispatch('app/podDiskReadContainerView', Object.assign({}, params))
                ])
                this.renderDiskChartContainerView(res[0].data.result, res[1].data.result)
            } catch (e) {
                catchErrorHandler(e, this)
            } finally {
                this.$refs.instanceNetLineContainerView && this.$refs.instanceNetLineContainerView.hideLoading()
            }
        },

        /**
         * 渲染 容器磁盘读写数据 图表
         *
         * @param {Array} listWrite 容器磁盘写数据
         * @param {Array} listRead 容器磁盘读数据
         */
        renderDiskChartContainerView (listWrite, listRead) {
            const chartNode = this.$refs.instanceNetLineContainerView
            if (!chartNode) {
                return
            }

            const podNetChartOptsContainerView = Object.assign({}, this.podNetChartOptsContainerView)
            podNetChartOptsContainerView.series.splice(0, podNetChartOptsContainerView.series.length, ...[])
            podNetChartOptsContainerView.yAxis.splice(0, podNetChartOptsContainerView.yAxis.length, ...[
                {
                    ...this.yAxisDefaultConf,
                    axisLabel: {
                        color: '#868b97',
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
                podNetChartOptsContainerView.series.push(
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
                podNetChartOptsContainerView.series.push(
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

            podNetChartOptsContainerView.tooltip.formatter = (params, ticket, callback) => {
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

            // chartNode.mergeOptions({
            //     tooltip: {
            //         formatter (params, ticket, callback) {
            //             if (params[0].value[3] === '--') {
            //                 return '<div>No Data</div>'
            //             }

            //             let date = params[0].value[0]
            //             if (String(parseInt(date, 10)).length === 10) {
            //                 date = parseInt(date, 10) + '000'
            //             }

            //             let ret = ''
            //                     + `<div>${moment(parseInt(date, 10)).format('YYYY-MM-DD HH:mm:ss')}</div>`

            //             params.forEach(p => {
            //                 if (p.value[2] === 'write') {
            //                     ret += `<div>${p.value[3]}-${labelWrite}：${formatBytes(p.value[1])}</div>`
            //                 } else if (p.value[2] === 'read') {
            //                     ret += `<div>${p.value[3]}-${labelRead}：${formatBytes(p.value[1])}</div>`
            //                 }
            //             })

            //             return ret
            //         }
            //     }
            // })

            chartNode.hideLoading()
        },

        // ------------------------------------------------------------------------------------------------------- //

        /**
         * 获取 cpu 图表数据
         */
        async fetchContainerMetricsCpu () {
            if (!this.containerIdList || !this.containerIdList.length) {
                return
            }
            try {
                const params = {
                    projectId: this.projectId,
                    res_id_list: this.containerIdList,
                    metric: 'cpu_summary',
                    cluster_id: this.clusterId
                }
                if (this.CATEGORY) {
                    params.category = this.CATEGORY
                }

                const res = await this.$store.dispatch('app/getAllContainerMetrics', params)

                setTimeout(() => {
                    this.renderCpuChart(
                        res.data.list && res.data.list.length
                            ? res.data.list
                            : [
                                {
                                    container_name: 'noData', metrics: [{ usage: 0, time: new Date().getTime() }]
                                }
                            ]
                    )
                }, 0)
            } catch (e) {
                catchErrorHandler(e, this)
            } finally {
                this.$refs.instanceCpuLine && this.$refs.instanceCpuLine.hideLoading()
            }
        },

        /**
         * 渲染 cpu 图表
         *
         * @param {Array} data 数据
         */
        renderCpuChart (data) {
            const seriesEmpty = []
            const series = []
            const seriesLen = data.length
            const ref = this.$refs.instanceCpuLine
            if (!ref) {
                return
            }

            for (let i = 0; i < seriesLen; i++) {
                const chartData = []
                const emptyData = []

                const curColor = chartColors[i % 10]

                data[i].metrics.forEach(metric => {
                    chartData.push({
                        value: [metric.time, metric.usage, curColor]
                    })
                    emptyData.push(0)
                })

                const name = data[i].container_name || this.containerIdNameMap[data[i].id]

                series.push({
                    type: 'line',
                    name: name,
                    showSymbol: false,
                    hoverAnimation: false,
                    lineStyle: {
                        normal: {
                            color: curColor,
                            opacity: randomInt(7, 10) / 10
                        }
                    },
                    data: chartData
                })
                seriesEmpty.push({
                    type: 'line',
                    name: name,
                    showSymbol: false,
                    hoverAnimation: false,
                    data: emptyData
                })
            }

            ref.mergeOptions({
                series: seriesEmpty
            })
            ref.mergeOptions({
                series: series
            })
        },

        /**
         * 获取 mem 图表数据
         *
         * @param {Object} ref chart ref
         * @param {string} metric 标识是 cpu 还是内存图表
         */
        async fetchContainerMetricsMem (ref, metric) {
            if (!this.containerIdList || !this.containerIdList.length) {
                return
            }
            try {
                const params = {
                    projectId: this.projectId,
                    res_id_list: this.containerIdList,
                    metric: 'mem',
                    cluster_id: this.clusterId
                }
                if (this.CATEGORY) {
                    params.category = this.CATEGORY
                }

                const res = await this.$store.dispatch('app/getAllContainerMetrics', params)

                setTimeout(() => {
                    this.renderMemChartInternal(
                        res.data.list && res.data.list.length
                            ? res.data.list
                            : [
                                {
                                    container_name: 'noData', metrics: [{ rss_pct: 0, time: new Date().getTime() }]
                                }
                            ]
                    )
                }, 0)
            } catch (e) {
                catchErrorHandler(e, this)
            } finally {
                this.$refs.instanceMemLine && this.$refs.instanceMemLine.hideLoading()
            }
        },

        /**
         * 渲染 mem 图表，内部版
         *
         * @param {Array} data 数据
         */
        renderMemChartInternal (data) {
            const seriesEmpty = []
            const series = []
            const seriesLen = data.length
            const ref = this.$refs.instanceMemLine
            if (!ref) {
                return
            }

            for (let i = 0; i < seriesLen; i++) {
                const chartData = []
                const emptyData = []

                const curColor = chartColors[i % 10]

                data[i].metrics.forEach(metric => {
                    chartData.push({
                        value: [metric.time, metric.rss_pct, curColor]
                    })
                    emptyData.push(0)
                })

                // this.containerIdNameMap[data[i].id] 不存在时，data[i].container_name 就是 noData
                // const name = this.containerIdNameMap[data[i].id] || data[i].container_name
                const name = data[i].container_name || this.containerIdNameMap[data[i].id]
                series.push({
                    type: 'line',
                    name: name,
                    showSymbol: false,
                    hoverAnimation: false,
                    lineStyle: {
                        normal: {
                            color: curColor,
                            opacity: 1
                        }
                    },
                    data: chartData
                })
                seriesEmpty.push({
                    type: 'line',
                    name: name,
                    showSymbol: false,
                    hoverAnimation: false,
                    data: emptyData
                })
            }

            ref.mergeOptions({
                series: seriesEmpty
            })
            ref.mergeOptions({
                series: series
            })
        },

        /**
         * 渲染 mem 图表，非内部版
         *
         * @param {Array} data 数据
         */
        renderMemChart (data) {
            const seriesEmpty = []
            const series = []
            const seriesLen = data.length
            const ref = this.$refs.instanceMemLine
            if (!ref) {
                return
            }

            for (let i = 0; i < seriesLen; i++) {
                const chartData = []
                const emptyData = []

                const curColor = chartColors[i % 10]

                data[i].metrics.forEach(metric => {
                    chartData.push({
                        value: [metric.time, metric.used, curColor]
                    })
                    emptyData.push(0)
                })

                // this.containerIdNameMap[data[i].id] 不存在时，data[i].container_name 就是 noData
                // const name = this.containerIdNameMap[data[i].id] || data[i].container_name
                const name = data[i].container_name || this.containerIdNameMap[data[i].id]
                series.push({
                    type: 'line',
                    name: name,
                    showSymbol: false,
                    hoverAnimation: false,
                    lineStyle: {
                        normal: {
                            color: curColor,
                            opacity: 1
                        }
                    },
                    data: chartData
                })
                seriesEmpty.push({
                    type: 'line',
                    name: name,
                    showSymbol: false,
                    hoverAnimation: false,
                    data: emptyData
                })
            }

            ref.mergeOptions({
                series: seriesEmpty
            })
            ref.mergeOptions({
                series: series
            })
        },

        /**
         * 显示 taskgroup 详情
         *
         * @param {Object} taskgroup 当前 taskgroup 对象
         * @param {number} index 当前 taskgroup 对象在 taskgroupList 里的索引
         */
        async showTaskgroupInfo (taskgroup, index) {
            this.taskgroupInfoDialogConf.isShow = true
            this.taskgroupInfoDialogConf.loading = true
            this.taskgroupInfoDialogConf.title = taskgroup.name

            try {
                this.baseData = Object.assign({}, {})
                this.updateData = Object.assign({}, {})
                this.restartData = ''

                const params = {
                    projectId: this.projectId,
                    instanceId: this.instanceId,
                    taskgroupName: taskgroup.name,
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
                    this.killData = ''
                } else {
                    this.killData = Object.assign({}, {})
                }

                const res = await this.$store.dispatch('app/getTaskgroupInfo', params)

                this.baseData = Object.assign({}, res.data.base_info || {})
                this.baseData.lastUpdateTime = this.baseData.last_update_time
                    ? moment(this.baseData.last_update_time).format('YYYY-MM-DD HH:mm:ss')
                    : ''
                this.baseData.createTime = this.baseData.start_time
                    ? moment(this.baseData.start_time).format('YYYY-MM-DD HH:mm:ss')
                    : ''

                let diffStr = ''
                if (this.baseData.current_time && this.baseData.start_time) {
                    const timeDiff = moment.duration(
                        moment(this.baseData.current_time, 'YYYY-MM-DD HH:mm:ss').diff(
                            moment(this.baseData.start_time, 'YYYY-MM-DD HH:mm:ss')
                        )
                    )
                    const arr = [
                        moment(this.baseData.current_time).diff(moment(this.baseData.start_time), 'days'),
                        timeDiff.get('hour'),
                        timeDiff.get('minute'),
                        timeDiff.get('second')
                    ]
                    diffStr = (arr[0] !== 0 ? (arr[0] + this.$t('天')) : '')
                        + (arr[1] !== 0 ? (arr[1] + this.$t('小时')) : '')
                        + (arr[2] !== 0 ? (arr[2] + this.$t('分')) : '')
                        + (arr[3] !== 0 ? (arr[3] + this.$t('秒')) : '')
                }

                this.baseData.surviveTime = diffStr

                this.updateData = Object.assign({}, res.data.update_strategy || {})
                this.restartData = res.data.restart_policy || ''
                if (this.CATEGORY) {
                    this.killData = res.data.kill_policy || ''
                } else {
                    this.killData = Object.assign({}, res.data.kill_policy || {})
                }
            } catch (e) {
                catchErrorHandler(e, this)
            } finally {
                this.taskgroupInfoDialogConf.loading = false
            }
        },

        /**
         * 展开/收起 taskgroup 里的表格
         *
         * @param {Object} taskgroup 当前 taskgroup 对象
         * @param {number} index 当前 taskgroup 对象在 taskgroupList 里的索引
         */
        async toggleContainers (taskgroup, index) {
            taskgroup.isOpen = !taskgroup.isOpen
            this.$set(this.taskgroupList, index, taskgroup)
            if (!taskgroup.isOpen) {
                return
            }

            taskgroup.containerLoading = true
            const containerList = await this.getTaskGroupContainer(taskgroup)
            taskgroup.containerList = containerList
            taskgroup.containerLoading = false
            this.$set(this.taskgroupList, index, taskgroup)
        },

        async getTaskGroupContainer (taskgroup) {
            try {
                const params = {
                    projectId: this.projectId,
                    instanceId: this.instanceId,
                    taskgroupName: taskgroup.name,
                    cluster_id: this.clusterId
                }

                if (String(this.instanceId) === '0') {
                    params.name = this.instanceName
                    params.namespace = this.instanceNamespace
                    params.category = this.instanceCategory
                }

                if (this.CATEGORY) {
                    params.category = this.CATEGORY
                }

                const res = await this.$store.dispatch('app/getContainterList', params)
                let containerList = res.data || []

                const containerIds = containerList.map(container => container.container_id).join(',')
                // 区分企业版和内部版
                if (containerIds && this.$INTERNAL) {
                    const logParams = {
                        projectId: this.projectId,
                        container_ids: containerIds
                    }
                    const logRes = await this.$store.dispatch('app/getContaintersLogLinks', logParams)
                    const logLinks = logRes.data || {}
                    containerList = containerList.map(container => ({
                        ...container,
                        ...(logLinks[container.container_id] || {})
                    }))
                }

                return containerList
            } catch (e) {
                catchErrorHandler(e, this)
                return []
            }
        },

        /**
         * 显示重新调度确认框
         *
         * @param {Object} taskgroup 当前 taskgroup 对象
         * @param {number} index 当前 taskgroup 对象在 taskgroupList 里的索引
         */
        async showRescheduler (taskgroup, index) {
            if (taskgroup.status.toLowerCase() === 'lost') {
                this.reschedulerDialogConf.isShow = true
                this.reschedulerDialogConf.title = this.$t('当前taskgroup处于lost状态，请确认上面容器已经不再运行')
                this.reschedulerDialogConf.curRescheduler = Object.assign({}, taskgroup)
                this.reschedulerDialogConf.curReschedulerIndex = index
            } else {
                await this.rescheduler(taskgroup, index)
            }
        },

        /**
         * 隐藏重新调度确认框
         */
        hideRescheduler () {
            this.reschedulerDialogConf.isShow = false
            setTimeout(() => {
                this.reschedulerDialogConf.title = ''
            }, 500)
        },

        /**
         * 重新调度确认框的确认
         */
        reschedulerConfirm () {
            this.hideRescheduler()
            this.rescheduler(this.reschedulerDialogConf.curRescheduler, this.reschedulerDialogConf.curReschedulerIndex)
        },

        /**
         * 重新调度
         *
         * @param {Object} taskgroup 当前 taskgroup 对象
         * @param {number} index 当前 taskgroup 对象在 taskgroupList 里的索引
         */
        async rescheduler (taskgroup, index) {
            clearTimeout(this.taskgroupTimer)
            this.taskgroupTimer = null

            const statusTmp = taskgroup.status

            taskgroup.isOpen = false
            taskgroup.status = 'Starting'

            this.$set(this.taskgroupList, index, taskgroup)
            try {
                const params = {
                    projectId: this.projectId,
                    instanceId: this.instanceId,
                    taskgroupName: taskgroup.name,
                    cluster_id: this.clusterId
                }

                if (String(this.instanceId) === '0') {
                    params.name = this.instanceName
                    params.namespace = this.instanceNamespace
                    params.category = this.instanceCategory
                }

                if (this.CATEGORY) {
                    params.category = this.CATEGORY
                }

                await this.$store.dispatch('app/reschedulerTaskgroup', params)
                clearTimeout(this.taskgroupTimer)
                this.taskgroupTimer = null
                this.taskgroupTimer = setTimeout(() => {
                    this.loopTaskgroup()
                }, 5000)
            } catch (e) {
                taskgroup.status = statusTmp
                this.$set(this.taskgroupList, index, taskgroup)
                catchErrorHandler(e, this)
            } finally {
                this.reschedulerDialogConf.curRescheduler = null
                this.reschedulerDialogConf.curReschedulerIndex = -1
            }
        },

        /**
         * 轮询 taskgroup
         */
        async loopTaskgroup () {
            try {
                const params = {
                    projectId: this.projectId,
                    instanceId: this.instanceId,
                    cluster_id: this.clusterId
                }

                if (String(this.instanceId) === '0') {
                    params.name = this.instanceName
                    params.namespace = this.instanceNamespace
                    params.category = this.instanceCategory
                }

                if (this.CATEGORY) {
                    params.category = this.CATEGORY
                }

                const res = await this.$store.dispatch('app/getTaskgroupList', params)

                this.openTaskgroup = Object.assign({}, {})
                this.taskgroupList.forEach(item => {
                    if (item.isOpen) {
                        this.openTaskgroup[item.name] = item.containerList
                    }
                })

                this.taskgroupList.splice(0, this.taskgroupList.length, ...[])

                const list = res.data || []
                list.forEach(item => {
                    let diffStr = ''
                    if (item.current_time && item.start_time) {
                        const timeDiff = moment.duration(
                            moment(item.current_time, 'YYYY-MM-DD HH:mm:ss').diff(
                                moment(item.start_time, 'YYYY-MM-DD HH:mm:ss')
                            )
                        )
                        const arr = [
                            moment(item.current_time).diff(moment(item.start_time), 'days'),
                            timeDiff.get('hour'),
                            timeDiff.get('minute'),
                            timeDiff.get('second')
                        ]

                        diffStr = (arr[0] !== 0 ? (arr[0] + this.$t('天')) : '')
                            + (arr[1] !== 0 ? (arr[1] + this.$t('小时')) : '')
                            + (arr[2] !== 0 ? (arr[2] + this.$t('分')) : '')
                            + (arr[3] !== 0 ? (arr[3] + this.$t('秒')) : '')
                    }

                    this.taskgroupList.push({
                        ...item,
                        isOpen: !!this.openTaskgroup[item.name],
                        containerList: this.openTaskgroup[item.name] || [],
                        surviveTime: diffStr
                    })
                })

                this.taskgroupTimer = setTimeout(() => {
                    this.loopTaskgroup()
                }, 5000)
            } catch (e) {
                console.error(e, this)
            }
        },

        /**
         * 打开到终端入口
         *
         * @param {Object} container 当前容器
         */
        async showTerminal (container, taskgroup) {
            const cluster = this.instanceInfo
            const clusterId = cluster.cluster_id
            const containerId = container.container_id
            const url = `${DEVOPS_BCS_API_URL}/web_console/projects/${this.projectId}/clusters/${clusterId}/?namespace=${this.instanceInfo.namespace_name}&pod_name=${taskgroup.name}&container_name=${container.name}`
            if (this.terminalWins.hasOwnProperty(containerId)) {
                const win = this.terminalWins[containerId]
                if (!win.closed) {
                    this.terminalWins[containerId].focus()
                } else {
                    const win = window.open(url, '_blank')
                    this.terminalWins[containerId] = win
                }
            } else {
                const win = window.open(url, '_blank')
                this.terminalWins[containerId] = win
            }
        },

        /**
         * 显示容器日志
         *
         * @param {Object} container 当前容器
         */
        async showLog (container) {
            this.logSideDialogConf.isShow = true
            this.logSideDialogConf.title = container.name
            this.logSideDialogConf.container = container
            this.logLoading = true
            try {
                const params = {
                    projectId: this.projectId,
                    containerId: container.container_id,
                    with_localtime: 1,
                    cluster_id: this.clusterId
                }

                if (String(this.instanceId) === '0') {
                    params.name = this.instanceName
                    params.namespace = this.instanceNamespace
                    params.category = this.instanceCategory
                }

                if (this.CATEGORY) {
                    params.category = this.CATEGORY
                }

                const res = await this.$store.dispatch('app/getContainterLog', params)
                this.logList.splice(0, this.logList.length, ...(res.data || []))
            } catch (e) {
                catchErrorHandler(e, this)
            } finally {
                this.logLoading = false
            }
        },

        async handleShowLog (taskGroup, index) {
            this.bcsLog.show = true
            this.bcsLog.loading = true
            const { containerList = [] } = this.taskgroupList[index]
            if (!containerList.length) {
                const list = await this.getTaskGroupContainer(taskGroup)
                taskGroup.containerList = list
                this.$set(this.taskgroupList, index, taskGroup)
            }
            this.bcsLog.containerList = this.taskgroupList[index].containerList.map(item => ({
                id: item.name,
                name: item.name
            }))
            this.bcsLog.podId = this.taskgroupList[index].name
            this.bcsLog.defaultContainer = this.taskgroupList[index].containerList[0]?.name
            this.bcsLog.loading = false
        },

        /**
         * 关闭日志
         *
         * @param {Object} cluster 当前集群对象
         */
        closeLog () {
            this.logSideDialogConf.isShow = false
            this.logSideDialogConf.title = ''
            this.logSideDialogConf.container = null
            this.logSideDialogConf.showLogTime = false
            this.logList.splice(0, this.logList.length, ...[])
        },

        /**
         * 刷新日志
         */
        async refreshLog () {
            this.logLoading = true
            try {
                const params = {
                    projectId: this.projectId,
                    containerId: this.logSideDialogConf.container.container_id,
                    with_localtime: 1,
                    cluster_id: this.clusterId
                }

                if (String(this.instanceId) === '0') {
                    params.name = this.instanceName
                    params.namespace = this.instanceNamespace
                    params.category = this.instanceCategory
                }

                if (this.CATEGORY) {
                    params.category = this.CATEGORY
                }

                const res = await this.$store.dispatch('app/getContainterLog', params)
                // this.logList.unshift(...(res.data || []))
                this.logList.splice(0, this.logList.length, ...(res.data || []))
            } catch (e) {
                catchErrorHandler(e, this)
            } finally {
                this.logLoading = false
            }
        },

        /**
         * 获取下方 tab 标签的数据
         */
        async fetchLabel () {
            this.labelListLoading = true
            try {
                const params = {
                    projectId: this.projectId,
                    instanceId: this.instanceId,
                    instanceName: this.instanceInfo.name,
                    cluster_id: this.clusterId
                }

                if (String(this.instanceId) === '0') {
                    params.name = this.instanceName
                    params.namespace = this.instanceNamespace
                    params.category = this.instanceCategory
                }

                if (this.CATEGORY) {
                    params.category = this.CATEGORY
                }

                const res = await this.$store.dispatch('app/getLabelList', params)
                const list = res.data || []
                this.labelList.splice(0, this.labelList.length, ...list)
            } catch (e) {
                catchErrorHandler(e, this)
            } finally {
                this.labelListLoading = false
            }
        },

        /**
         * 获取下方 tab 备注的数据
         */
        async fetchAnnotation () {
            this.annotationListLoading = true
            try {
                const params = {
                    projectId: this.projectId,
                    instanceId: this.instanceId,
                    instanceName: this.instanceInfo.name,
                    cluster_id: this.clusterId
                }

                if (String(this.instanceId) === '0') {
                    params.name = this.instanceName
                    params.namespace = this.instanceNamespace
                    params.category = this.instanceCategory
                }

                if (this.CATEGORY) {
                    params.category = this.CATEGORY
                }

                const res = await this.$store.dispatch('app/getAnnotationList', params)
                const list = res.data || []
                this.annotationList.splice(0, this.annotationList.length, ...list)
            } catch (e) {
                catchErrorHandler(e, this)
            } finally {
                this.annotationListLoading = false
            }
        },

        /**
         * 获取下方 tab metric 数据
         */
        async fetchMetric () {
            this.metricListLoading = true
            this.metricListErrorMessage = this.$t('没有数据')
            try {
                const params = {
                    projectId: this.projectId,
                    instanceId: this.instanceId,
                    cluster_id: this.clusterId
                }

                if (String(this.instanceId) === '0') {
                    params.name = this.instanceName
                    params.namespace = this.instanceNamespace
                    params.category = this.instanceCategory
                }

                if (this.CATEGORY) {
                    params.category = this.CATEGORY
                }

                const res = await this.$store.dispatch('app/getMetricList', params)
                const list = res.data || []
                this.metricList.splice(0, this.metricList.length, ...list)
            } catch (e) {
                this.metricListErrorMessage = e.message || e.data.msg || e.statusText
                catchErrorHandler(e, this)
            } finally {
                this.metricListLoading = false
            }
        },

        /**
         * 获取下方 tab 事件的数据
         *
         * @param {number} offset 起始页码
         * @param {number} limit 偏移量
         */
        async fetchEvent (offset = 0, limit = this.eventPageConf.pageSize) {
            this.eventListLoading = true
            try {
                const params = {
                    projectId: this.projectId,
                    instanceId: this.instanceId,
                    cluster_id: this.clusterId,
                    offset,
                    limit
                }

                if (String(this.instanceId) === '0') {
                    params.name = this.instanceName
                    params.namespace = this.instanceNamespace
                    params.category = this.instanceCategory
                }

                if (this.CATEGORY) {
                    params.category = this.CATEGORY
                }

                const res = await this.$store.dispatch('app/getEventList', params)

                const count = res.data.total || 0
                const list = []
                res.data.data.forEach(item => {
                    list.push({
                        eventTime: moment(item.eventTime).format('YYYY-MM-DD HH:mm:ss'),
                        component: item.component,
                        obj: item.extraInfo.name,
                        level: item.level,
                        describe: `${item.type}：${item.describe}`
                    })
                })
                this.eventList.splice(0, this.eventList.length, ...list)
                this.eventPageConf.total = count
                this.eventPageConf.totalPage = Math.ceil(count / this.eventPageConf.pageSize)
                if (this.eventPageConf.totalPage < this.eventPageConf.curPage) {
                    this.eventPageConf.curPage = 1
                }
            } catch (e) {
                catchErrorHandler(e, this)
            } finally {
                this.eventListLoading = false
            }
        },

        /**
         * 翻页
         *
         * @param {number} page 页码
         */
        eventPageChange (page) {
            this.fetchEvent(this.eventPageConf.pageSize * (page - 1), this.eventPageConf.pageSize)
        },

        /**
         * 选项卡切换事件
         *
         * @param {string} name 选项卡标识
         */
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
            } else if (name === 'annotation') {
                this.annotationList.splice(0, this.annotationList.length, ...[])
                this.fetchAnnotation()
            } else if (name === 'taskgroup') {
                this.fetchTaskgroup(false)
            } else if (name === 'event') {
                this.fetchEvent()
            } else if (name === 'metric') {
                this.fetchMetric()
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
            if (idx === 'podCpu') {
                ref = this.$refs.instanceCpuLine
                hook = 'fetchPodCpuUsage'
            }
            if (idx === 'podMem') {
                ref = this.$refs.instanceMemLine
                hook = 'fetchPodMemUsage'
            }
            if (idx === 'podNet') {
                ref = this.$refs.instanceNetLine
                hook = 'fetchPodNet'
            }
            ref && ref.showLoading({
                text: this.$t('正在加载中...'),
                color: '#30d878',
                maskColor: 'rgba(255, 255, 255, 0.8)'
            })

            this[hook](range)
        },

        /**
         * 切换时间范围
         *
         * @param {Object} dropdownRef dropdown 标识
         * @param {string} toggleRangeStr 标识
         * @param {string} idx 标识，cpu / memory / network / storage
         * @param {string} range 时间范围，1: 1 小时，2: 24 小时，3：近 7 天
         */
        toggleContainerRange (dropdownRef, toggleRangeStr, idx, range) {
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
            if (idx === 'containerCpu') {
                ref = this.$refs.instanceCpuLineContainerView
                hook = 'fetchPodCpuUsageContainerView'
            }
            if (idx === 'containerMem') {
                ref = this.$refs.instanceMemLineContainerView
                hook = 'fetchPodMemUsageContainerView'
            }
            if (idx === 'containerDisk') {
                ref = this.$refs.instanceNetLineContainerView
                hook = 'fetchDiskContainerView'
            }

            ref && ref.showLoading({
                text: this.$t('正在加载中...'),
                color: '#30d878',
                maskColor: 'rgba(255, 255, 255, 0.8)'
            })

            this[hook](range)
        },

        handleShowEditorSearch () {
            this.$refs.yamlEditor && this.$refs.yamlEditor.showSearchBox()
        },

        handleCopyContent (value) {
            copyText(value)
            this.$bkMessage({
                theme: 'success',
                message: this.$t('复制成功')
            })
        }
    }
}
