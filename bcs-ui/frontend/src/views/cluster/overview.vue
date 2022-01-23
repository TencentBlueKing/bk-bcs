<template>
    <div class="biz-content">
        <div class="biz-content-wrapper biz-cluster-overview" v-bkloading="{ isLoading: showLoading }">
            <app-exception
                v-if="exceptionCode && !showLoading"
                :type="exceptionCode.code"
                :text="exceptionCode.msg">
            </app-exception>
            <div v-if="!exceptionCode && !showLoading" class="biz-cluster-overview-wrapper">
                <div class="biz-cluster-tab-content">
                    <div class="biz-cluster-overview-chart">
                        <div class="chart-box top">
                            <div class="info">
                                <div class="left">
                                    {{$t('CPU使用率')}}
                                </div>
                                <div class="right">
                                    <div><span>{{cpuUsagePercent}}</span><sup>%</sup></div>
                                    <div class="cpu"><span>{{cpuUsage}} of {{cpuTotal}}</span></div>
                                </div>
                            </div>
                            <cluster-overview-chart
                                :result-type="cpuChartResultType"
                                :show-loading="cpuChartLoading"
                                :chart-type="'cpu'"
                                :data="cpuChartData">
                            </cluster-overview-chart>
                        </div>

                        <div class="chart-box top">
                            <div class="info">
                                <div class="left">
                                    {{$t('内存使用率')}}
                                </div>
                                <div class="right">
                                    <div><span>{{memUsagePercent}}</span><sup>%</sup></div>
                                    <div class="memory"><span>{{memUsage}} of {{memTotal}}</span></div>
                                </div>
                            </div>
                            <cluster-overview-chart
                                :result-type="memChartResultType"
                                :show-loading="memChartLoading"
                                :chart-type="'mem'"
                                :data="memChartData">
                            </cluster-overview-chart>
                        </div>

                        <div class="chart-box top">
                            <div class="info">
                                <div class="left">{{$t('磁盘使用率')}}</div>
                                <div class="right">
                                    <div><span>{{diskUsagePercent}}</span><sup>%</sup></div>
                                    <div class="disk"><span>{{diskUsage}} of {{diskTotal}}</span></div>
                                </div>
                            </div>
                            <cluster-overview-chart
                                :result-type="diskChartResultType"
                                :show-loading="diskChartLoading"
                                :chart-type="'disk'"
                                :data="diskChartData">
                            </cluster-overview-chart>
                        </div>
                    </div>

                    <div class="biz-cluster-overview-chart">
                        <div class="chart-box bottom">
                            <div class="info">
                                <div class="left">{{$t('节点')}}</div>
                                <div class="right">
                                    <div>
                                        <i class="bcs-icon bcs-icon-circle"></i>
                                        <span>{{$t('使用中')}}</span>
                                        <span>{{nodeActived}}</span>
                                    </div>
                                    <div>
                                        <i class="bcs-icon bcs-icon-circle"></i>
                                        <span>{{$t('未使用')}}</span>
                                        <span>{{nodeDisabled}}</span>
                                    </div>
                                </div>
                            </div>
                            <Ring :percent="nodePercent" :size="210" :text="'none'"
                                :stroke-width="10" :fill-width="10" :fill-color="'#3ede78'"
                                :percent-change-handler="percentChangeHandler('node')"
                                :ext-cls="'biz-cluster-ring'">
                                <div slot="text" class="ring-text-inner">
                                    <div class="number">{{nodePercentStr}}</div>
                                </div>
                            </Ring>
                        </div>

                        <div class="chart-box bottom">
                            <div class="info">
                                <div class="left">{{$t('命名空间')}}</div>
                                <div class="right">
                                    <div>
                                        <i class="bcs-icon bcs-icon-circle"></i>
                                        <span>{{$t('已使用')}}</span>
                                        <span>{{namespaceActived}}</span>
                                    </div>
                                    <div>
                                        <i class="bcs-icon bcs-icon-circle"></i>
                                        <span>{{$t('总量')}}</span>
                                        <span>{{namespaceTotal}}</span>
                                    </div>
                                </div>
                            </div>
                            <Ring :percent="namespacePercent" :size="210" :text="'none'"
                                :stroke-width="10" :fill-width="10" :fill-color="'#3ede78'"
                                :percent-change-handler="percentChangeHandler('namespace')"
                                :ext-cls="'biz-cluster-ring'">
                                <div slot="text" class="ring-text-inner">
                                    <div class="number">{{namespacePercentStr}}</div>
                                </div>
                            </Ring>
                        </div>

                        <div class="chart-box bottom">
                            <div class="info">
                                <div class="left">{{$t('集群IP')}}</div>
                                <div class="right">
                                    <div>
                                        <i class="bcs-icon bcs-icon-circle"></i>
                                        <span>{{$t('可用')}}</span>
                                        <span>{{ipTotal}}</span>
                                    </div>
                                    <div>
                                        <i class="bcs-icon bcs-icon-circle"></i>
                                        <span>{{$t('已使用')}}</span>
                                        <span>{{ipUsed}}</span>
                                    </div>
                                </div>
                            </div>
                            <Ring :percent="ipPercent" :size="210" :text="'none'"
                                :stroke-width="10" :fill-width="10" :fill-color="'#3ede78'"
                                :percent-change-handler="percentChangeHandler('ip')"
                                :ext-cls="'biz-cluster-ring'">
                                <div slot="text" class="ring-text-inner">
                                    <div class="number">{{ipPercentStr}}</div>
                                </div>
                            </Ring>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script>
    import Ring from '@/components/ring'
    import { catchErrorHandler, formatBytes } from '@/common/util'

    import ClusterOverviewChart from './cluster-overview-chart'

    export default {
        components: {
            Ring,
            ClusterOverviewChart
        },
        data () {
            return {
                nodePercent: 0,
                nodePercentStr: 0,
                nodeActived: '',
                nodeDisabled: '',
                namespacePercent: 0,
                namespacePercentStr: 0,
                namespaceActived: '',
                namespaceTotal: '',
                ipPercent: 0,
                ipPercentStr: 0,
                ipUsed: '',
                ipTotal: '',
                bkMessageInstance: null,
                curClusterInPage: {},
                exceptionCode: null,
                showLoading: false,

                cpuUsagePercent: 0,
                cpuUsage: 0,
                cpuTotal: 0,
                memUsagePercent: 0,
                memUsage: 0,
                memTotal: 0,
                diskUsagePercent: 0,
                diskUsage: 0,
                diskTotal: 0,
                cpuChartData: [],
                cpuChartLoading: true,
                cpuChartResultType: 'matrix',
                memChartData: [],
                memChartLoading: true,
                memChartResultType: 'matrix',
                diskChartData: [],
                diskChartLoading: true,
                diskChartResultType: 'matrix'
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
            curClusterId () {
                return this.$store.state.curClusterId
            },
            curCluster () {
                const data = this.$store.state.cluster.clusterList.find(item => item.cluster_id === this.clusterId) || {}
                this.curClusterInPage = Object.assign({}, data)
                return JSON.parse(JSON.stringify(data))
            },
            isEn () {
                return this.$store.state.isEn
            },
            curProject () {
                return this.$store.state.curProject
            },
            clusterPerm () {
                return this.$store.state.cluster.clusterPerm
            }
        },
        async created () {
            if (!this.clusterPerm[this.curCluster?.clusterID]?.policy?.view) {
                await this.$store.dispatch('getResourcePermissions', {
                    project_id: this.projectId,
                    policy_code: 'view',
                    // eslint-disable-next-line camelcase
                    resource_code: this.curCluster?.cluster_id,
                    resource_name: this.curCluster?.name,
                    resource_type: `cluster_${this.curCluster?.environment === 'prod' ? 'prod' : 'test'}`
                }).catch(err => {
                    this.exceptionCode = {
                        code: err.code,
                        msg: err.message
                    }
                })
            }
        },
        async mounted () {
            if (this.curCluster.project_id && this.curCluster.cluster_id) {
                await this.fetchClusterOverview()
                this.fetchClusterMetrics()
                this.prepareChartData()
            }
        },
        destroyed () {
        },
        methods: {
            /**
             * 集群使用率概览
             */
            async fetchClusterOverview () {
                if (!this.curCluster.project_id || !this.curCluster.cluster_id) return

                try {
                    const res = await this.$store.dispatch('cluster/clusterOverview', {
                        projectId: this.curCluster.project_id,
                        clusterId: this.curCluster.cluster_id
                    })
                    const data = res.data || {}
                    const cpu = data.cpu_usage || {}
                    this.cpuUsage = parseFloat(cpu.used || 0).toFixed(2)
                    this.cpuTotal = parseFloat(cpu.total || 0).toFixed(2)
                    this.cpuUsagePercent = this.conversionPercentUsed(cpu.used, cpu.total)

                    const mem = data.memory_usage || {}
                    this.memUsage = formatBytes(mem.used_bytes || 0)
                    this.memTotal = formatBytes(mem.total_bytes || 0)
                    this.memUsagePercent = this.conversionPercentUsed(mem.used_bytes, mem.total_bytes)

                    const disk = data.disk_usage || {}
                    this.diskUsage = formatBytes(disk.used_bytes || 0)
                    this.diskTotal = formatBytes(disk.total_bytes || 0)
                    this.diskUsagePercent = this.conversionPercentUsed(disk.used_bytes, disk.total_bytes)
                } catch (e) {
                    catchErrorHandler(e, this)
                }
            },

            /**
             * 转换百分比
             *
             * @param {number} used 使用量
             * @param {number} total 总量
             *
             * @return {number} 百分比数字
             */
            conversionPercentUsed (used, total) {
                if (!total || parseFloat(total) === 0) {
                    return 0
                }

                let ret = parseFloat(used) / parseFloat(total) * 100
                if (ret !== 0 && ret !== 100) {
                    ret = ret.toFixed(2)
                }
                return ret
            },

            /**
             * 转换百分比
             *
             * @param {number} remain 剩下的数量
             * @param {number} total 总量
             *
             * @return {number} 百分比数字
             */
            conversionPercent (remain, total) {
                if (!remain || !total) {
                    return 0
                }
                return total === 0 ? 0 : ((total - remain) / total * 100).toFixed(2)
            },

            /**
             * 构建图表数据
             */
            async prepareChartData () {
                if (!this.curCluster.project_id || !this.curCluster.cluster_id) return
                try {
                    this.cpuChartLoading = true
                    this.memChartLoading = true
                    this.diskChartLoading = true
                    const promises = [
                        this.$store.dispatch('cluster/clusterCpuUsage', {
                            projectId: this.curCluster.project_id, // 这里用 this.curCluster 来获取是为了使计算属性生效
                            clusterId: this.curCluster.cluster_id
                        }),
                        this.$store.dispatch('cluster/clusterMemUsage', {
                            projectId: this.curCluster.project_id, // 这里用 this.curCluster 来获取是为了使计算属性生效
                            clusterId: this.curCluster.cluster_id
                        }),
                        this.$store.dispatch('cluster/clusterDiskUsage', {
                            projectId: this.curCluster.project_id, // 这里用 this.curCluster 来获取是为了使计算属性生效
                            clusterId: this.curCluster.cluster_id
                        })
                    ]

                    const res = await Promise.all(promises)
                    // Promise.all 返回的顺序与 promises 数组的顺序是一致的
                    // 如果为空，那么在 res.data.result 这里就是空数组
                    // 如果不是空数组，那么 res.data.result 里的任何都是有数据的，所以不判断里面的了
                    this.cpuChartData.splice(0, this.cpuChartData.length, ...(res[0].data.result || []))
                    this.cpuChartResultType = res[0].data.resultType
                    this.cpuChartLoading = false

                    this.memChartData.splice(0, this.memChartData.length, ...(res[1].data.result || []))
                    this.memChartResultType = res[1].data.resultType
                    this.memChartLoading = false

                    this.diskChartData.splice(0, this.diskChartData.length, ...(res[2].data.result || []))
                    this.diskChartResultType = res[2].data.resultType
                    this.diskChartLoading = false
                } catch (e) {
                    catchErrorHandler(e, this)
                }
            },

            /**
             * 获取下面三个圈的数据
             */
            async fetchClusterMetrics () {
                if (!this.curCluster.project_id || !this.curCluster.cluster_id) return

                try {
                    const res = await this.$store.dispatch('cluster/getClusterMetrics', {
                        projectId: this.curCluster.project_id,
                        clusterId: this.curCluster.cluster_id
                    })

                    const nodeActived = res.data.node.actived || 0
                    const nodeTotal = res.data.node.total || 0
                    if (nodeTotal === 0) {
                        this.nodePercent = 0
                        this.nodePercentStr = 0
                    } else {
                        const nodePercent = nodeActived * 100 / nodeTotal
                        this.nodePercent = nodePercent
                        this.nodePercentStr = nodePercent === 100 ? '100%' : nodePercent.toFixed(1) + '%'
                    }
                    this.nodeActived = this.isEn ? `${nodeActived} set` : `${nodeActived}台`
                    this.nodeDisabled = this.isEn ? `${res.data.node.disabled || 0} set` : `${res.data.node.disabled || 0}台`

                    const namespaceActived = res.data.namespace.actived || 0
                    const namespaceTotal = res.data.namespace.total || 0
                    if (namespaceTotal === 0) {
                        this.namespacePercent = 0
                        this.namespacePercentStr = 0
                    } else {
                        const namespacePercent = namespaceActived * 100 / namespaceTotal
                        this.namespacePercent = namespacePercent
                        this.namespacePercentStr = namespacePercent === 100 ? '100%' : namespacePercent.toFixed(1) + '%'
                    }
                    this.namespaceActived = this.isEn ? namespaceActived : `${namespaceActived}个`
                    this.namespaceTotal = this.isEn ? namespaceTotal : `${namespaceTotal}个`

                    const ipUsed = res.data.ip_resource.used || 0
                    const ipTotal = res.data.ip_resource.total || 0
                    if (ipTotal === 0) {
                        this.ipPercent = 0
                        this.ipPercentStr = 0
                    } else {
                        const ipPercent = ipUsed * 100 / ipTotal
                        this.ipPercent = ipPercent
                        this.ipPercentStr = ipPercent === 100 ? '100%' : ipPercent.toFixed(1) + '%'
                    }
                    this.ipUsed = this.isEn ? ipUsed : `${ipUsed}个`
                    this.ipTotal = this.isEn ? ipTotal : `${ipTotal}个`
                } catch (e) {
                    catchErrorHandler(e, this)
                }
            },

            /**
             * 刷新当前 router
             */
            refreshCurRouter () {
                typeof this.$parent.refreshRouterView === 'function' && this.$parent.refreshRouterView()
            },

            /**
             * 返回集群首页列表
             */
            goIndex () {
                const { params } = this.$route
                if (params.backTarget) {
                    this.$router.push({
                        name: params.backTarget,
                        params: {
                            projectId: this.projectId,
                            projectCode: this.projectCode
                        }
                    })
                } else {
                    this.$router.push({
                        name: 'clusterMain',
                        params: {
                            projectId: this.projectId,
                            projectCode: this.projectCode
                        }
                    })
                }
            },

            /**
             * 切换到节点管理
             */
            goNode () {
                this.$router.push({
                    name: 'clusterNode',
                    params: {
                        projectId: this.projectId,
                        projectCode: this.projectCode,
                        clusterId: this.clusterId,
                        backTarget: this.$route.params.backTarget
                    }
                })
            },

            /**
             * ring 组件百分比变化回调函数
             *
             * @param {string} indicator 标识当前是哪个 ring 组件
             */
            percentChangeHandler (indicator) {
                return percent => (this[`${indicator}PercentStr`] = `${percent}%`)
            },

            /**
             * 切换到集群信息列表
             */
            goInfo () {
                this.$router.push({
                    name: 'clusterInfo',
                    params: {
                        projectId: this.projectId,
                        projectCode: this.projectCode,
                        clusterId: this.clusterId,
                        backTarget: this.$route.params.backTarget
                    }
                })
            }
        }
    }
</script>
<style scoped lang="postcss">
    @import '@/css/variable.css';
    @import '@/css/mixins/clearfix.css';

    .biz-cluster-overview {
        padding: 0px;
    }

    .biz-cluster-overview-title {
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
            color: $iconPrimaryColor
        }
    }

    .biz-cluster-overview-wrapper {
        background-color: $bgHoverColor;
        display: inline-block;
        width: 100%;
        border-radius: 2px;
    }

    .biz-cluster-tab-header {
        display: flex;
        height: 60px;
        line-height: 60px;
        font-size: 0;
        border-bottom: 1px solid $borderWeightColor;

        .header-item {
            font-size: 14px;
            display: flex;
            justify-content: center;
            align-items: center;
            width: 140px;
            border: none;
            cursor: pointer;

            i {
                font-size: 16px;
                margin-right: 8px;
            }

            &.active {
                color: $iconPrimaryColor;
                background-color: #fff;
                border-right: 1px solid $borderWeightColor;
                font-weight: 700;
                cursor: default;

                i {
                    font-weight: 700;
                }
            }
        }
    }

    .biz-cluster-tab-content {
        background-color: #fff;
        font-size: 14px;
    }

    .biz-cluster-overview-chart {
        background-color: #fff;
        display: flex;
        justify-content: center;

        &:nth-child(1) {
            border-bottom: 1px solid $borderWeightColor;
        }

        .chart-box {
            @mixin clearfix;
            height: 360px;
            padding: 20px;
            position: relative;
            flex: 1;
            border-left: 1px solid $borderWeightColor;

            &:nth-child(1) {
                border-left: none;
            }
        }

        .ring-text-inner {
            position: absolute;
            transform: translate(-50%, -50%);
            left: 50%;
            top: 50%;
            text-align: center;
        }

        .number {
            font-size: 50px;
        }

        .label {
            font-size: 14px;
        }

        .biz-cluster-ring {
            position: absolute;
            transform: translate(-50%, -50%);
            left: 50%;
            top: 56%;
        }

        .info {
            display: inline-block;
            width: 100%;
        }

        .chart-box.bottom {
            @mixin clearfix;
            font-size: 14px;

            .left {
                font-weight: 700;
                float: left;
            }

            .right {
                float: right;
                font-size: 14px;

                i {
                    font-weight: 700;
                    vertical-align: middle;
                    margin-right: 7px;
                }

                span {
                    display: inline-block;
                    width: 50px;

                    &:last-child {
                        text-align: right;
                        font-weight: 700;
                    }
                }

                div:first-child {
                    margin-bottom: 10px;

                    i {
                        color: #3ede78;
                    }
                }

                div:last-child {
                    i {
                        color: $borderColor;
                    }
                }
            }
        }

        .chart-box.top {
            .left {
                font-weight: 700;
                float: left;
            }

            .right {
                text-align: right;
                float: right;
                font-size: 14px;

                div:first-child {
                    font-size: 32px;

                    sup {
                        font-size: 20px;
                    }
                }

                div:last-child {
                    font-weight: 700;

                    &.cpu {
                        color: #3ede78;
                    }

                    &.memory {
                        color: #3a84ff;
                    }

                    &.disk {
                        color: #853cff;
                    }
                }
            }
        }

        .echarts {
            width: 100%;
            height: 250px;
        }
    }
</style>
