<template>
    <div class="biz-content">
        <div class="biz-top-bar">
            <div class="biz-app-instance-title">
                <i class="bcs-icon bcs-icon-arrows-left back" @click="goNodeOverview"></i>
                <span @click="refreshCurRouter">{{containerInfo.container_name || '--'}}</span>
            </div>
            <bk-guide></bk-guide>
        </div>
        <div class="biz-content-wrapper biz-app-instance">
            <app-exception v-if="exceptionCode" :type="exceptionCode.code" :text="exceptionCode.msg"></app-exception>
            <div v-else class="biz-app-instance-wrapper">
                <div class="biz-app-instance-header">
                    <div class="header-item">
                        <div class="key-label">{{$t('主机名称：')}}</div>
                        <bcs-popover :delay="500" placement="bottom-start">
                            <div class="value-label">{{containerInfo.host_name || '--'}}</div>
                            <template slot="content">
                                <p style="text-align: left; white-space: normal;word-break: break-all;font-weight: 400;">{{containerInfo.host_name || '--'}}</p>
                            </template>
                        </bcs-popover>
                    </div>
                    <div class="header-item">
                        <div class="key-label">{{$t('主机IP：')}}</div>
                        <bcs-popover :delay="500" placement="bottom">
                            <div class="value-label">{{containerInfo.host_ip || '--'}}</div>
                            <template slot="content">
                                <p style="text-align: left; white-space: normal;word-break: break-all;font-weight: 400;">{{containerInfo.host_ip || '--'}}</p>
                            </template>
                        </bcs-popover>
                    </div>
                    <div class="header-item">
                        <div class="key-label">{{$t('容器IP：')}}</div>
                        <bcs-popover :delay="500" placement="bottom">
                            <div class="value-label">{{containerInfo.container_ip || '--'}}</div>
                            <template slot="content">
                                <p style="text-align: left; white-space: normal;word-break: break-all;font-weight: 400;">{{containerInfo.container_ip || '--'}}</p>
                            </template>
                        </bcs-popover>
                    </div>
                    <div class="header-item">
                        <div class="key-label">{{$t('容器ID：')}}</div>
                        <bcs-popover :delay="500" placement="bottom">
                            <div class="value-label">{{containerInfo.container_id || '--'}}</div>
                            <template slot="content">
                                <p style="text-align: left; white-space: normal;word-break: break-all;font-weight: 400;">{{containerInfo.container_id || '--'}}</p>
                            </template>
                        </bcs-popover>
                    </div>
                    <div class="header-item">
                        <div class="key-label">{{$t('镜像：')}}</div>
                        <bcs-popover :delay="500" placement="bottom">
                            <div class="value-label">{{containerInfo.image || '--'}}</div>
                            <template slot="content">
                                <p style="text-align: left; white-space: normal;word-break: break-all;font-weight: 400;">{{containerInfo.image || '--'}}</p>
                            </template>
                        </bcs-popover>
                    </div>
                    <div class="header-item">
                        <div class="key-label">{{$t('网络模式：')}}</div>
                        <bcs-popover :delay="500" placement="bottom">
                            <div class="value-label">{{containerInfo.network_mode || '--'}}</div>
                            <template slot="content">
                                <p style="text-align: left; white-space: normal;word-break: break-all;font-weight: 400;">{{containerInfo.network_mode || '--'}}</p>
                            </template>
                        </bcs-popover>
                    </div>
                </div>
                <div class="biz-app-instance-chart-wrapper">
                    <div class="biz-app-instance-chart-k8s">
                        <div class="part top-left">
                            <div class="info">
                                <div class="left">{{$t('CPU使用率')}}</div>
                            </div>
                            <chart :options="containerCpuChartOpts" ref="containerCpuLine" auto-resize></chart>
                        </div>
                        <div class="part top-left">
                            <div class="info">
                                <div class="left">{{$t('内存使用量')}}</div>
                            </div>
                            <chart :options="containerMemChartOptsInternal" ref="containerMemLine" auto-resize></chart>
                        </div>
                        <div class="part top-right">
                            <div class="info">
                                <div class="left">{{$t('磁盘IO')}}</div>
                            </div>
                            <chart :options="containerDiskChartOptsInternal" ref="containerDiskLine" auto-resize></chart>
                        </div>
                    </div>
                </div>
                <div class="biz-app-container-table-wrapper">
                    <bk-tab :type="'fill'" class="biz-tab-container" :active-name="tabActiveName" @tab-changed="tabChanged">
                        <bk-tab-panel name="ports" :title="$t('端口映射')">
                            <table class="bk-table has-table-hover biz-table biz-app-container-ports-table">
                                <thead>
                                    <tr>
                                        <th style="text-align: left;padding-left: 27px; width: 300px">
                                            Name
                                        </th>
                                        <th style="width: 150px">Host Port</th>
                                        <th style="width: 150px">Container Port</th>
                                        <th style="width: 100px">Protocol</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    <template v-if="portList.length">
                                        <tr v-for="(port, index) in portList" :key="index">
                                            <td style="text-align: left;padding-left: 27px;">
                                                <bcs-popover placement="top" :delay="500">
                                                    <p class="port-name">{{port.name}}</p>
                                                    <template slot="content">
                                                        <p style="text-align: left; white-space: normal;word-break: break-all;">{{port.name}}</p>
                                                    </template>
                                                </bcs-popover>
                                            </td>
                                            <td>{{port.hostPort}}</td>
                                            <td>{{port.containerPort}}</td>
                                            <td>
                                                <bcs-popover placement="top" :delay="500">
                                                    <p class="port-protocol">{{port.protocol}}</p>
                                                    <template slot="content">
                                                        <p style="text-align: left; white-space: normal;word-break: break-all;">{{port.protocol}}</p>
                                                    </template>
                                                </bcs-popover>
                                            </td>
                                        </tr>
                                    </template>
                                    <template v-else>
                                        <tr>
                                            <td colspan="4">
                                                <div class="bk-message-box no-data">
                                                    <p class="message empty-message">{{$t('该应用的网络模式无需端口映射')}}</p>
                                                </div>
                                            </td>
                                        </tr>
                                    </template>
                                </tbody>
                            </table>
                        </bk-tab-panel>
                        <bk-tab-panel name="commands" :title="$t('命令')">
                            <table class="bk-table has-table-hover biz-table biz-app-container-commands-table">
                                <thead>
                                    <tr>
                                        <th style="text-align: left;padding-left: 27px; width: 300px">
                                            Command
                                        </th>
                                        <th style="width: 200px">Args</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    <template v-if="commandList.length">
                                        <tr v-for="(command, index) in commandList" :key="index">
                                            <td style="text-align: left;padding-left: 27px;">
                                                <bcs-popover placement="top" :delay="500">
                                                    <p class="command-name">{{command.command}}</p>
                                                    <template slot="content">
                                                        <p style="text-align: left; white-space: normal;word-break: break-all;">{{command.command}}</p>
                                                    </template>
                                                </bcs-popover>
                                            </td>
                                            <td>
                                                <bcs-popover placement="top" :delay="500">
                                                    <p class="command-args">{{command.args}}</p>
                                                    <template slot="content">
                                                        <p style="text-align: left; white-space: normal;word-break: break-all;">{{command.args}}</p>
                                                    </template>
                                                </bcs-popover>
                                            </td>
                                        </tr>
                                    </template>
                                    <template v-else>
                                        <tr>
                                            <td colspan="2">
                                                <div class="bk-message-box no-data">
                                                    <bcs-exception type="empty" scene="part"></bcs-exception>
                                                </div>
                                            </td>
                                        </tr>
                                    </template>
                                </tbody>
                            </table>
                        </bk-tab-panel>
                        <bk-tab-panel name="volumes" :title="$t('挂载卷')">
                            <table class="bk-table has-table-hover biz-table biz-app-container-volumes-table">
                                <thead>
                                    <tr>
                                        <th style="text-align: left;padding-left: 27px; width: 250px">
                                            Host Path
                                        </th>
                                        <th style="width: 250px">Mount Path</th>
                                        <th style="width: 140px">ReadOnly</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    <template v-if="volumeList.length">
                                        <tr v-for="(volume, index) in volumeList" :key="index">
                                            <td style="text-align: left;padding-left: 27px;">
                                                <bcs-popover placement="top" :delay="500">
                                                    <p class="volume-host">{{volume.hostPath}}</p>
                                                    <template slot="content">
                                                        <p style="text-align: left; white-space: normal;word-break: break-all;">{{volume.hostPath}}</p>
                                                    </template>
                                                </bcs-popover>
                                            </td>
                                            <td>
                                                <bcs-popover placement="top" :delay="500">
                                                    <p class="volume-mount">{{volume.mountPath}}</p>
                                                    <template slot="content">
                                                        <p style="text-align: left; white-space: normal;word-break: break-all;">{{volume.mountPath}}</p>
                                                    </template>
                                                </bcs-popover>
                                            </td>
                                            <td>{{volume.readOnly}}</td>
                                        </tr>
                                    </template>
                                    <template v-else>
                                        <tr>
                                            <td colspan="3">
                                                <div class="bk-message-box no-data">
                                                    <bcs-exception type="empty" scene="part"></bcs-exception>
                                                </div>
                                            </td>
                                        </tr>
                                    </template>
                                </tbody>
                            </table>
                        </bk-tab-panel>
                        <bk-tab-panel name="env_args" :title="$t('环境变量')">
                            <table class="bk-table has-table-hover biz-table biz-app-container-env-table">
                                <thead>
                                    <tr>
                                        <th style="text-align: left;padding-left: 27px; width: 150px">
                                            Key
                                        </th>
                                        <th style="width: 350px">Value</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    <template v-if="envList.length">
                                        <tr v-for="(env, index) in envList" :key="index">
                                            <td style="text-align: left;padding-left: 27px;">
                                                <bcs-popover placement="top" :delay="500">
                                                    <p class="env-key">{{env.key}}</p>
                                                    <template slot="content">
                                                        <p style="text-align: left; white-space: normal;word-break: break-all;">{{env.key}}</p>
                                                    </template>
                                                </bcs-popover>
                                            </td>
                                            <td>
                                                <bcs-popover placement="top" :delay="500">
                                                    <p class="env-value">{{env.value}}</p>
                                                    <template slot="content">
                                                        <p style="text-align: left; white-space: normal;word-break: break-all;">{{env.value}}</p>
                                                    </template>
                                                </bcs-popover>
                                            </td>
                                        </tr>
                                    </template>
                                    <template v-else>
                                        <tr>
                                            <td colspan="2">
                                                <div class="bk-message-box no-data">
                                                    <bcs-exception type="empty" scene="part"></bcs-exception>
                                                </div>
                                            </td>
                                        </tr>
                                    </template>
                                </tbody>
                            </table>
                        </bk-tab-panel>
                        <bk-tab-panel name="health_check" :title="$t('健康检查')">
                            <table class="bk-table has-table-hover biz-table biz-app-container-health-table">
                                <thead>
                                    <tr>
                                        <th style="text-align: left;padding-left: 27px; width: 150px">
                                            Type
                                        </th>
                                        <th style="width: 140px">Result</th>
                                        <th style="width: 350px">Message</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    <template v-if="healthList.length">
                                        <tr v-for="(health, index) in healthList" :key="index">
                                            <td style="text-align: left;padding-left: 27px;">
                                                <bcs-popover placement="top" :delay="500">
                                                    <p class="health-type">{{health.type}}</p>
                                                    <template slot="content">
                                                        <p style="text-align: left; white-space: normal;word-break: break-all;">{{health.type}}</p>
                                                    </template>
                                                </bcs-popover>
                                            </td>
                                            <td>{{health.result}}</td>
                                            <td>
                                                <bcs-popover placement="top" :delay="500">
                                                    <p class="health-message">{{health.message}}</p>
                                                    <template slot="content">
                                                        <p style="text-align: left; white-space: normal;word-break: break-all;">{{health.message}}</p>
                                                    </template>
                                                </bcs-popover>
                                            </td>
                                        </tr>
                                    </template>
                                    <template v-else>
                                        <tr>
                                            <td colspan="3">
                                                <div class="bk-message-box no-data">
                                                    <bcs-exception type="empty" scene="part"></bcs-exception>
                                                </div>
                                            </td>
                                        </tr>
                                    </template>
                                </tbody>
                            </table>
                        </bk-tab-panel>
                        <bk-tab-panel name="labels" :title="$t('标签')">
                            <table class="bk-table has-table-hover biz-table biz-app-container-label-table">
                                <thead>
                                    <tr>
                                        <th style="text-align: left;padding-left: 27px; width: 150px">
                                            Key
                                        </th>
                                        <th style="width: 350px">Value</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    <template v-if="labelList.length">
                                        <tr v-for="(label, index) in labelList" :key="index">
                                            <td style="text-align: left;padding-left: 27px;">
                                                <bcs-popover placement="top" :delay="500">
                                                    <p class="label-key">{{label.key}}</p>
                                                    <template slot="content">
                                                        <p style="text-align: left; white-space: normal;word-break: break-all;">{{label.key}}</p>
                                                    </template>
                                                </bcs-popover>
                                            </td>
                                            <td>
                                                <bcs-popover placement="top" :delay="500">
                                                    <p class="label-value">{{label.val}}</p>
                                                    <template slot="content">
                                                        <p style="text-align: left; white-space: normal;word-break: break-all;">{{label.val}}</p>
                                                    </template>
                                                </bcs-popover>
                                            </td>
                                        </tr>
                                    </template>
                                    <template v-else>
                                        <tr>
                                            <td colspan="2">
                                                <div class="bk-message-box no-data">
                                                    <bcs-exception type="empty" scene="part"></bcs-exception>
                                                </div>
                                            </td>
                                        </tr>
                                    </template>
                                </tbody>
                            </table>
                        </bk-tab-panel>
                        <bk-tab-panel name="resources" :title="$t('资源限制')">
                            <table class="bk-table has-table-hover biz-table biz-app-container-resource-table">
                                <thead>
                                    <tr>
                                        <th style="text-align: left;padding-left: 27px; width: 150px">
                                            Cpu
                                        </th>
                                        <th style="width: 350px">Memory</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    <template v-if="resourceList.length">
                                        <tr v-for="(resource, index) in resourceList" :key="index">
                                            <td style="text-align: left;padding-left: 27px;">
                                                <p class="resource-cpu">{{parseFloat(resource.cpu)}}</p>
                                            </td>
                                            <td>
                                                <p class="resource-mem">{{parseFloat(resource.memory)}}</p>
                                            </td>
                                        </tr>
                                    </template>
                                    <template v-else>
                                        <tr>
                                            <td colspan="2">
                                                <div class="bk-message-box no-data">
                                                    <bcs-exception type="empty" scene="part"></bcs-exception>
                                                </div>
                                            </td>
                                        </tr>
                                    </template>
                                </tbody>
                            </table>
                        </bk-tab-panel>
                    </bk-tab>
                </div>
            </div>
        </div>
    </div>
</template>

<script>
    import ECharts from 'vue-echarts/components/ECharts.vue'
    import 'echarts/lib/chart/line'
    import 'echarts/lib/component/tooltip'
    import 'echarts/lib/component/legend'
    import moment from 'moment'
    import { Decimal } from 'decimal.js'

    import { containerDetailChart } from '@/common/chart-option'
    import { catchErrorHandler, formatBytes } from '@/common/util'

    import { createChartOption } from '@/views/app/container-chart-opts'

    export default {
        components: {
            chart: ECharts
        },
        data () {
            return {
                containerInfo: {},

                containerCpuChartOpts: createChartOption(this),

                containerMemChartOptsInternal: createChartOption(this),

                memLine: containerDetailChart.mem,

                containerDiskChartOptsInternal: createChartOption(this),

                diskLine: containerDetailChart.disk,

                tabActiveName: 'ports',
                portList: [],
                commandList: [],
                volumeList: [],
                envList: [],
                healthList: [],
                labelList: [],
                resourceList: [],
                bkMessageInstance: null,
                exceptionCode: null,
                projectIdTimer: null
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
            containerId () {
                return this.$route.params.containerId
            },
            nodeId () {
                return this.$route.params.nodeId
            },
            curProject () {
                return this.$store.state.curProject
            }
        },
        async mounted () {
            await this.fetchContainerInfo()
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
            this.$refs.containerDiskLine && this.$refs.containerDiskLine.showLoading({
                text: this.$t('正在加载'),
                color: '#30d878',
                maskColor: 'rgba(255, 255, 255, 0.8)'
            })
            this.fetchContainerCpuUsage()
            this.fetchContainerMemUsage()
            this.fetchContainerDisk()
        },
        destroyed () {
            this.bkMessageInstance && this.bkMessageInstance.close()
        },
        methods: {
            /**
             * 获取容器详情信息，上方数据和下方
             */
            async fetchContainerInfo () {
                try {
                    const res = await this.$store.dispatch('cluster/getContainerInfo', {
                        projectId: this.projectId,
                        clusterId: this.clusterId,
                        containerId: this.containerId
                    })
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

                    const envList = []
                    const envArgs = this.containerInfo.env_args || {}
                    Object.keys(envArgs).forEach(key => {
                        envList.push({
                            key: key,
                            value: envArgs[key]
                        })
                    })
                    this.envList.splice(0, this.envList.length, ...envList)

                    const healthList = this.containerInfo.health_check || []
                    this.healthList.splice(0, this.healthList.length, ...healthList)

                    const labelList = this.containerInfo.labels || []
                    this.labelList.splice(0, this.labelList.length, ...labelList)

                    const resourceList = []
                    const resources = this.containerInfo.resources || {}
                    const limits = resources.limits || {}
                    if (limits.cpu || limits.memory) {
                        resourceList.push({
                            cpu: limits.cpu,
                            memory: limits.memory
                        })
                    }
                    this.resourceList.splice(0, this.resourceList.length, ...resourceList)
                } catch (e) {
                    catchErrorHandler(e, this)
                }
            },

            /**
             * 获取 容器CPU使用率
             */
            async fetchContainerCpuUsage () {
                try {
                    const params = {
                        projectId: this.projectId,
                        container_ids: this.containerId.split(','),
                        clusterId: this.clusterId,
                        namespace: this.containerInfo.namespace
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
                    values: [[parseInt(String(+new Date()).slice(0, 10), 10), '--']]
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
                            let ret = ''

                            if (params[0].value[1] === '--') {
                                ret = '<div>No Data</div>'
                            } else {
                                let thresholdStr = ''
                                const valueLen0 = String(params[0].value[1]).length
                                if (params[1] && params[1].seriesName === 'threshold') {
                                    const valueLen1 = String(params[1].value[1]).length
                                    thresholdStr = `<div style="color: #fd9c9c;">Limit: ${Decimal(params[1].value[1]).toPrecision(valueLen1 > 3 ? 3 : valueLen1)}%</div>`
                                }
                                ret = `
                                    <div>${moment(parseInt(params[0].value[0], 10)).format('YYYY-MM-DD HH:mm:ss')}</div>
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
             */
            async fetchContainerMemUsage () {
                try {
                    const params = {
                        projectId: this.projectId,
                        container_ids: this.containerId.split(','),
                        clusterId: this.clusterId,
                        namespace: this.containerInfo.namespace
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
                    values: [[parseInt(String(+new Date()).slice(0, 10), 10), '--']]
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
                            let ret = ''

                            if (params[0].value[1] === '--') {
                                ret = '<div>No Data</div>'
                            } else {
                                let thresholdStr = ''
                                if (params[1] && params[1].seriesName === 'threshold') {
                                    thresholdStr = `<div style="color: #fd9c9c;">Limit: ${formatBytes(params[1].value[1])}</div>`
                                }
                                ret = `
                                    <div>${moment(parseInt(params[0].value[0], 10)).format('YYYY-MM-DD HH:mm:ss')}</div>
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
             */
            async fetchContainerDisk () {
                try {
                    const params = {
                        projectId: this.projectId,
                        container_ids: this.containerId.split(','),
                        clusterId: this.clusterId,
                        namespace: this.containerInfo.namespace
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
                        values: [[parseInt(String(+new Date()).slice(0, 10), 10), '--']]
                    }]

                const dataRead = listRead.length
                    ? listRead
                    : [{
                        metric: { container_name: '--' },
                        values: [[parseInt(String(+new Date()).slice(0, 10), 10), '--']]
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
                            if (params[0].value[1] === '--') {
                                return '<div>No Data</div>'
                            }

                            let ret = ''
                                + `<div>${moment(parseInt(params[0].value[0], 10)).format('YYYY-MM-DD HH:mm:ss')}</div>`

                            params.forEach(p => {
                                if (p.value[2] === 'write') {
                                    ret += `<div>${params[0].value[3]}-${labelWrite}：${formatBytes(p.value[1])}</div>`
                                } else if (p.value[2] === 'read') {
                                    ret += `<div>${params[0].value[3]}-${labelRead}：${formatBytes(p.value[1])}</div>`
                                }
                            })

                            return ret
                        }
                    }
                })

                chartNode.hideLoading()
            },

            /**
             * 刷新当前 router
             */
            refreshCurRouter () {
                typeof this.$parent.refreshRouterView === 'function' && this.$parent.refreshRouterView()
            },

            /**
             * 返回 node overview
             */
            goNodeOverview () {
                this.$router.push({
                    name: 'clusterNodeOverview',
                    params: {
                        projectId: this.projectId,
                        projectCode: this.projectCode,
                        clusterId: this.clusterId,
                        nodeId: this.nodeId
                    }
                })
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
            }
        }
    }
</script>

<style scoped lang="postcss">
    @import '@/css/variable.css';
    @import '@/css/mixins/ellipsis.css';

    .biz-app-instance {
        padding: 20px;
    }

    .biz-app-instance-title {
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

    .biz-app-instance-wrapper {
        background-color: $bgHoverColor;
        display: inline-block;
        width: 100%;
    }

    .biz-app-instance-header {
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
                @mixin ellipsis 180px;
                padding-top: 4px;
            }
        }
    }

    .biz-app-instance-chart-wrapper {
        margin-top: 20px;
        background-color: #fff;
        box-shadow: 1px 0 2px rgba(0, 0, 0, 0.1);
        border: 1px solid $borderWeightColor;
        font-size: 0;
        border-radius: 2px;

        .biz-app-instance-chart-k8s {
            display: flex;
            width: 100%;
            .part {
                flex: 1;
                height: 250px;
                &.top-left {
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

        .biz-app-instance-chart {
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

                .right {

                    .system,
                    .user {
                        display: inline-block;
                        font-size: 14px;

                        .circle {
                            display: inline-block;
                            width: 14px;
                            height: 14px;
                            border-radius: 50%;
                            position: relative;
                            top: 2px;
                        }
                    }

                    .system {
                        .circle {
                            border: 3px solid #3a84ff;
                        }
                    }

                    .user {
                        margin-left: 30px;

                        .circle {
                            border: 3px solid #30d873;
                        }
                    }
                }
            }
        }
    }

    .echarts {
        width: 100%;
        height: 180px;
    }

    .biz-app-container-table-wrapper {
        margin-top: 20px;
    }

    .biz-app-container-ports-table,
    .biz-app-container-commands-table,
    .biz-app-container-volumes-table,
    .biz-app-container-health-table,
    .biz-app-container-env-table,
    .biz-app-container-label-table,
    .biz-app-container-resource-table {
        border-bottom: none;

        .no-data {
            min-height: 180px;

            .empty-message {
                margin-top: 50px;
            }
        }
    }

    .biz-app-container-ports-table {
        .port-name {
            @mixin ellipsis 300px;
        }

        .port-protocol {
            @mixin ellipsis 100px;
        }
    }

    .biz-app-container-commands-table {
        .command-name {
            @mixin ellipsis 300px;
        }

        .command-args {
            @mixin ellipsis 200px;
        }
    }

    .biz-app-container-volumes-table {
        .volume-host {
            @mixin ellipsis 250px;
        }

        .volume-mount {
            @mixin ellipsis 250px;
        }
    }

    .biz-app-container-health-table {
        .health-type {
            @mixin ellipsis 150px;
        }

        .health-message {
            @mixin ellipsis 350px;
        }
    }

    @media screen and (max-width: $mediaWidth) {
        .biz-app-instance-header {
            .header-item {
                div {
                    &:last-child {
                        width: 120px;
                    }
                }
            }
        }
    }

</style>
