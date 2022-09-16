<template>
    <chart :options="renderChartOpts" ref="chartNode" auto-resize></chart>
</template>

<script>
    import moment from 'moment'

    import ECharts from 'vue-echarts/components/ECharts.vue'
    import 'echarts/lib/chart/line'
    import 'echarts/lib/component/tooltip'

    export default {
        components: {
            chart: ECharts
        },
        props: {
            chartType: {
                type: String
            },
            showLoading: {
                type: Boolean,
                default: true
            },
            data: {
                type: Array,
                default: () => []
            },
            // matrix，数据为二维数组
            // vector，数据为一维数组
            resultType: {
                type: String,
                default: 'matrix'
            }
        },
        data () {
            return {
                cpuChartOpts: {
                    tooltip: {
                        trigger: 'axis',
                        axisPointer: {
                            type: 'line',
                            animation: false,
                            label: {
                                backgroundColor: '#6a7985'
                            }
                        }
                    },
                    grid: {
                        show: false,
                        top: '4%',
                        left: '4%',
                        right: '5%',
                        bottom: '3%',
                        containLabel: true
                    },
                    xAxis: [
                        {
                            type: 'time',
                            boundaryGap: false,
                            axisLine: {
                                show: true,
                                lineStyle: {
                                    color: '#dde4eb'
                                }
                            },
                            axisTick: {
                                alignWithLabel: true,
                                length: 5,
                                lineStyle: {
                                    color: '#ebf0f5'
                                }
                            },
                            axisLabel: {
                                color: '#868b97',
                                formatter (value, index) {
                                    if (String(parseInt(value, 10)).length === 10) {
                                        value = parseInt(value, 10) + '000'
                                    }
                                    return moment(parseInt(value, 10)).format('HH:mm')
                                }
                            },
                            splitLine: {
                                show: true,
                                lineStyle: {
                                    color: ['#ebf0f5'],
                                    type: 'dashed'
                                }
                            }
                        }
                    ],
                    yAxis: [
                        {
                            boundaryGap: [0, '2%'],
                            type: 'value',
                            axisLine: {
                                show: true,
                                lineStyle: {
                                    color: '#dde4eb'
                                }
                            },
                            axisTick: {
                                alignWithLabel: true,
                                length: 0,
                                lineStyle: {
                                    color: 'red'
                                }
                            },
                            axisLabel: {
                                color: '#868b97',
                                formatter (value, index) {
                                    return `${value.toFixed(1)}%`
                                }
                            },
                            splitLine: {
                                show: true,
                                lineStyle: {
                                    color: ['#ebf0f5'],
                                    type: 'dashed'
                                }
                            }
                        }
                    ],
                    series: []
                },
                memChartOpts: {
                    tooltip: {
                        trigger: 'axis',
                        axisPointer: {
                            type: 'line',
                            animation: false,
                            label: {
                                backgroundColor: '#6a7985'
                            }
                        }
                    },
                    grid: {
                        show: false,
                        top: '4%',
                        left: '4%',
                        right: '5%',
                        bottom: '3%',
                        containLabel: true
                    },
                    xAxis: [
                        {
                            type: 'time',
                            boundaryGap: false,
                            axisLine: {
                                show: true,
                                lineStyle: {
                                    color: '#dde4eb'
                                }
                            },
                            axisTick: {
                                alignWithLabel: true,
                                length: 5,
                                lineStyle: {
                                    color: '#ebf0f5'
                                }
                            },
                            axisLabel: {
                                color: '#868b97',
                                formatter (value, index) {
                                    if (String(parseInt(value, 10)).length === 10) {
                                        value = parseInt(value, 10) + '000'
                                    }
                                    return moment(parseInt(value, 10)).format('HH:mm')
                                }
                            },
                            splitLine: {
                                show: true,
                                lineStyle: {
                                    color: ['#ebf0f5'],
                                    type: 'dashed'
                                }
                            }
                        }
                    ],
                    yAxis: [
                        {
                            boundaryGap: [0, '2%'],
                            type: 'value',
                            axisLine: {
                                show: true,
                                lineStyle: {
                                    color: '#dde4eb'
                                }
                            },
                            axisTick: {
                                alignWithLabel: true,
                                length: 0,
                                lineStyle: {
                                    color: 'red'
                                }
                            },
                            axisLabel: {
                                color: '#868b97',
                                formatter (value, index) {
                                    return `${(value).toFixed(1)}%`
                                }
                            },
                            splitLine: {
                                show: true,
                                lineStyle: {
                                    color: ['#ebf0f5'],
                                    type: 'dashed'
                                }
                            }
                        }
                    ],
                    series: [
                    ]
                },
                diskChartOpts: {
                    tooltip: {
                        trigger: 'axis',
                        axisPointer: {
                            type: 'line',
                            animation: false,
                            label: {
                                backgroundColor: '#6a7985'
                            }
                        }
                    },
                    grid: {
                        show: false,
                        top: '4%',
                        left: '4%',
                        right: '5%',
                        bottom: '3%',
                        containLabel: true
                    },
                    xAxis: [
                        {
                            type: 'time',
                            boundaryGap: false,
                            axisLine: {
                                show: true,
                                lineStyle: {
                                    color: '#dde4eb'
                                }
                            },
                            axisTick: {
                                alignWithLabel: true,
                                length: 5,
                                lineStyle: {
                                    color: '#ebf0f5'
                                    // color: '#868b97'
                                }
                            },
                            axisLabel: {
                                color: '#868b97',
                                formatter (value, index) {
                                    if (String(parseInt(value, 10)).length === 10) {
                                        value = parseInt(value, 10) + '000'
                                    }
                                    return moment(parseInt(value, 10)).format('HH:mm')
                                }
                            },
                            splitLine: {
                                show: true,
                                lineStyle: {
                                    color: ['#ebf0f5'],
                                    type: 'dashed'
                                }
                            }
                        }
                    ],
                    yAxis: [
                        {
                            boundaryGap: [0, '2%'],
                            type: 'value',
                            axisLine: {
                                show: true,
                                lineStyle: {
                                    color: '#dde4eb'
                                }
                            },
                            axisTick: {
                                alignWithLabel: true,
                                length: 0,
                                lineStyle: {
                                    color: 'red'
                                }
                            },
                            axisLabel: {
                                color: '#868b97',
                                formatter (value, index) {
                                    return `${(value).toFixed(1)}%`
                                }
                            },
                            splitLine: {
                                show: true,
                                // show: false,
                                lineStyle: {
                                    color: ['#ebf0f5'],
                                    type: 'dashed'
                                }
                            }
                        }
                    ],
                    series: [
                    ]
                },
                renderChartOpts: null
            }
        },
        computed: {
            curCluster () {
                return this.$store.state.cluster.curCluster
            },
            isEn () {
                return this.$store.state.isEn
            }
        },
        watch: {
            // showLoading: {
            //     handler (newVal, oldVal) {
            //         if (newVal === true) {
            //             const chartNode = this.$refs.chartNode
            //             chartNode && chartNode.showLoading({
            //                 text: this.$t('正在加载中...'),
            //                 color: '#30d878',
            //                 maskColor: 'rgba(255, 255, 255, 0.8)'
            //             })
            //         }
            //     },
            //     immediate: true
            // }
            data (v) {
                setTimeout(() => {
                    this.renderMatrixChart(this.data)
                }, 0)
            }
        },
        created () {
            if (this.chartType === 'cpu') {
                this.renderChartOpts = Object.assign({}, this.cpuChartOpts)
            }
            if (this.chartType === 'mem') {
                this.renderChartOpts = Object.assign({}, this.memChartOpts)
            }
            if (this.chartType === 'disk') {
                this.renderChartOpts = Object.assign({}, this.diskChartOpts)
            }
        },
        mounted () {
            if (this.showLoading) {
                const chartNode = this.$refs.chartNode
                chartNode && chartNode.showLoading({
                    text: this.$t('正在加载中...'),
                    color: '#30d878',
                    maskColor: 'rgba(255, 255, 255, 0.8)'
                })
            } else {
                this.renderMatrixChart(this.data)
            }
            window.addEventListener('resize', this.resizeHandler)
        },
        destroyed () {
            window.removeEventListener('resize', this.resizeHandler)
        },
        methods: {
            resizeHandler () {
                this.$refs.chartNode && this.$refs.chartNode.resize()
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
             * 渲染 matrix 图表，数据为二维数组
             *
             * @param {Array} list 数据
             */
            renderMatrixChart (list) {
                const chartNode = this.$refs.chartNode
                if (!chartNode) {
                    return
                }

                const renderChartOpts = Object.assign({}, this.renderChartOpts)

                // const data = list.length ? list : [{ time: +new Date(), used: 0, total: 0 }]
                const data = list.length ? list : [{ values: [[parseInt(String(+new Date()).slice(0, 10), 10), '0']] }]

                data.forEach(item => {
                    // [
                    //     {
                    //         value: [item.time, this.conversionPercent(item.remain_cpu, item.total_cpu)]
                    //     }
                    // ]
                    let color = ''
                    if (this.chartType === 'cpu') {
                        color = '#30d878'
                    }

                    if (this.chartType === 'mem') {
                        color = '#3a84ff'
                    }

                    if (this.chartType === 'disk') {
                        color = '#853cff'
                    }

                    renderChartOpts.series.push(
                        {
                            type: 'line',
                            smooth: true,
                            showSymbol: false,
                            hoverAnimation: false,
                            areaStyle: {
                                normal: {
                                    opacity: 0.2
                                }
                            },
                            itemStyle: {
                                normal: {
                                    color: color
                                }
                            },
                            data: item.values
                        }
                    )
                })

                let label = ''
                if (this.chartType === 'cpu') {
                    label = this.$t('CPU使用率')
                }

                if (this.chartType === 'mem') {
                    label = this.$t('内存使用率')
                }

                if (this.chartType === 'disk') {
                    label = this.$t('磁盘使用率')
                }

                chartNode.mergeOptions({
                    tooltip: {
                        formatter (params, ticket, callback) {
                            let date = params[0].value[0]
                            if (String(parseInt(date, 10)).length === 10) {
                                date = parseInt(date, 10) + '000'
                            }
                            return `
                                <div>${moment(parseInt(date, 10)).format('YYYY-MM-DD HH:mm:ss')}</div>
                                <div>${label}：${parseFloat(params[0].value[1]).toFixed(2)}%</div>
                            `
                        }
                    }
                })

                chartNode.hideLoading()
            }
        }
    }
</script>
<style lang="postcss">
    .echarts {
        width: 100%;
        height: 250px;
    }
</style>
