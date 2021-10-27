/**
 * Tencent is pleased to support the open source community by making 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition) available.
 * Copyright (C) 2017-2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

import moment from 'moment'

const STYLE_STR = 'text-align: left; white-space: normal;word-break: break-all;'

/**
 * 集群总览页面
 *
 * @type {Object}
 */
export const overview = {
    // cpu Allocation 图表
    cpu: {
        tooltip: {
            trigger: 'axis',
            axisPointer: {
                type: 'line',
                animation: false,
                label: {
                    backgroundColor: '#6a7985'
                }
            },
            formatter (params, ticket, callback) {
                let ret

                if (params[0].value[2] === '无数据') {
                    ret = '<div>No Data</div>'
                } else {
                    ret = `
                        <div>${moment(params[0].value[0]).format('YYYY-MM-DD HH:mm:ss')}</div>
                        <div>CPU Usage：${params[0].value[1]}%</div>
                    `
                }

                return ret
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
                        return moment(value).format('HH:mm')
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
                        return `${value}%`
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
            {
                type: 'line',
                name: 'CPU Usage',
                // showSymbol: true,
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
                        color: '#30d878'
                    }
                }
            }
        ]
    },
    // memory Allocation 图表
    memory: {
        tooltip: {
            trigger: 'axis',
            axisPointer: {
                type: 'line',
                animation: false,
                label: {
                    backgroundColor: '#6a7985'
                }
            },
            formatter (params, ticket, callback) {
                let ret

                if (params[0].value[2] === '无数据') {
                    ret = '<div>No Data</div>'
                } else {
                    ret = `
                        <div>${moment(params[0].value[0]).format('YYYY-MM-DD HH:mm:ss')}</div>
                        <div>Memory Usage：${(params[0].value[1]).toFixed(2)}GB</div>
                    `
                }

                return ret
            }
        },
        grid: {
            show: false,
            top: '4%',
            left: '0%',
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
                        return moment(value).format('HH:mm')
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
                        return `${(value).toFixed(2)}GB`
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
            {
                type: 'line',
                name: 'memoryAllocation',
                // showSymbol: true,
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
                        color: '#3a84ff'
                    }
                }
            }
        ]
    },
    // disk Allocation 图表
    disk: {
        tooltip: {
            trigger: 'axis',
            axisPointer: {
                type: 'line',
                animation: false,
                label: {
                    backgroundColor: '#6a7985'
                }
            },
            formatter (params, ticket, callback) {
                let ret

                if (params[0].value[2] === '无数据') {
                    ret = '<div>No Data</div>'
                } else {
                    ret = `
                        <div>${moment(params[0].value[0]).format('YYYY-MM-DD HH:mm:ss')}</div>
                        <div>Disk Usage：${(params[0].value[1]).toFixed(2)}GB</div>
                    `
                }

                return ret
            }
        },
        grid: {
            // top: '4%',
            // left: '0',
            // right: '5%',
            // bottom: '3%',
            show: false,
            top: '4%',
            left: '0%',
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
                        return moment(value).format('HH:mm')
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
                        return `${(value).toFixed(2)}GB`
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
            {
                type: 'line',
                name: 'memoryAllocation',
                // showSymbol: true,
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
                        color: '#853cff'
                    }
                }
            }
        ]
    }
}

/**
 * 节点详情页面
 *
 * @type {Object}
 */
export const nodeOverview = {
    // cpu 使用率图表
    cpu: {
        tooltip: {
            trigger: 'axis',
            confine: true,
            axisPointer: {
                type: 'line',
                animation: false,
                label: {
                    backgroundColor: '#6a7985'
                }
            },
            formatter (params, ticket, callback) {
                let ret

                if (params[0].value[1] === '-') {
                    ret = '<div>No Data</div>'
                } else {
                    ret = `
                        <div>${moment(params[0].value[0]).format('YYYY-MM-DD HH:mm:ss')}</div>
                        <div>CPU Usage：${params[0].value[1]}%</div>
                    `
                }

                return ret
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
                    color: '#868b97'
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
                // min: 0,
                // max: 100,
                // interval: 25,
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
                        return `${value}%`
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
            {
                type: 'line',
                name: 'CPU Usage',
                showSymbol: false,
                smooth: true,
                hoverAnimation: false,
                areaStyle: {
                    normal: {
                        color: {
                            type: 'linear',
                            x: 0,
                            y: 0,
                            x2: 0,
                            y2: 1,
                            colorStops: [
                                {
                                    offset: 0, color: '#30d878' // 0% 处的颜色
                                },
                                {
                                    offset: 1, color: '#c0f3d6' // 100% 处的颜色
                                }
                            ],
                            globalCoord: false
                        }
                    }
                },
                itemStyle: {
                    normal: {
                        color: '#30d878'
                    }
                }
            }
        ]
    },
    // Memory Usage
    memory: {
        tooltip: {
            trigger: 'axis',
            axisPointer: {
                type: 'line',
                animation: false,
                label: {
                    backgroundColor: '#6a7985'
                }
            },
            formatter (params, ticket, callback) {
                let ret

                if (params[0].value[1] === '-' && params[1].value[1] === '-') {
                    ret = '<div>No Data</div>'
                } else {
                    ret = `
                        <div>${moment(params[0].value[0]).format('YYYY-MM-DD HH:mm:ss')}</div>
                        <div>${window.i18n.t('总大小')}：${(params[0].value[1] / 1024 / 1024 / 1024).toFixed(2)}GB</div>
                        <div>${window.i18n.t('已使用')}：${(params[1].value[1] / 1024 / 1024 / 1024).toFixed(2)}GB</div>
                    `
                }

                return ret
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
                    color: '#868b97'
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
                        return `${(value / 1024 / 1024 / 1024).toFixed(0)}GB`
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
            {
                type: 'line',
                name: 'total',
                smooth: true,
                showSymbol: false,
                hoverAnimation: false,
                areaStyle: {
                    normal: {
                        // color: {
                        //     type: 'linear',
                        //     x: 0,
                        //     y: 0,
                        //     x2: 0,
                        //     y2: 1,
                        //     colorStops: [
                        //         {
                        //             offset: 0, color: '#52a2ff' // 0% 处的颜色
                        //         },
                        //         {
                        //             offset: 1, color: '#a9d1ff' // 100% 处的颜色
                        //         }
                        //     ],
                        //     globalCoord: false
                        // }
                    }
                }
                // itemStyle: {
                //     normal: {
                //         color: '#52a2ff'
                //     }
                // }
            },
            {
                type: 'line',
                name: 'used',
                smooth: true,
                showSymbol: false,
                hoverAnimation: false,
                areaStyle: {
                    normal: {
                        // color: {
                        //     type: 'linear',
                        //     x: 0,
                        //     y: 0,
                        //     x2: 0,
                        //     y2: 1,
                        //     colorStops: [
                        //         {
                        //             offset: 0, color: '#52a2ff' // 0% 处的颜色
                        //         },
                        //         {
                        //             offset: 1, color: '#a9d1ff' // 100% 处的颜色
                        //         }
                        //     ],
                        //     globalCoord: false
                        // }
                    }
                }
                // itemStyle: {
                //     normal: {
                //         color: 'red'
                //     }
                // }
            }
        ]
    },
    // 网络使用率
    network: {
        tooltip: {
            trigger: 'axis',
            axisPointer: {
                type: 'line',
                animation: false,
                label: {
                    backgroundColor: '#6a7985'
                }
            },
            formatter (params, ticket, callback) {
                let ret

                if (params[0].seriesName === 'noData') {
                    ret = '<div>No Data</div>'
                } else {
                    ret = `<div>${moment(params[0].value[0]).format('YYYY-MM-DD HH:mm:ss')}</div>`
                    params.forEach(item => {
                        ret += `<div>${item.seriesName}${item.value[2] === 'sent' ? window.i18n.t('发送') : window.i18n.t('接收')}：${(item.value[1] || 0).toFixed(2)}KB/s</div>`
                    })
                }

                return ret
            }
        },
        // legend: {
        //     data: ['sent', 'recv']
        // },
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
                    color: '#868b97'
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
                        return `${value}KB/s`
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
            {
                type: 'line',
                // showSymbol: true,
                smooth: true,
                showSymbol: false,
                hoverAnimation: false,
                areaStyle: {
                    normal: {
                        // color: {
                        //     type: 'linear',
                        //     x: 0,
                        //     y: 0,
                        //     x2: 0,
                        //     y2: 1,
                        //     colorStops: [
                        //         {
                        //             offset: 0, color: '#ffbe21' // 0% 处的颜色
                        //         },
                        //         {
                        //             offset: 1, color: '#a9d1ff' // 100% 处的颜色
                        //         }
                        //     ],
                        //     globalCoord: false
                        // }
                    }
                },
                itemStyle: {
                    normal: {
                        color: '#ffbe21'
                    }
                }
            },
            {
                type: 'line',
                // showSymbol: true,
                smooth: true,
                showSymbol: false,
                hoverAnimation: false,
                areaStyle: {
                    normal: {
                    //     color: {
                    //         type: 'linear',
                    //         x: 0,
                    //         y: 0,
                    //         x2: 0,
                    //         y2: 1,
                    //         colorStops: [
                    //             {
                    //                 offset: 0, color: '#ffbe21' // 0% 处的颜色
                    //             },
                    //             {
                    //                 offset: 1, color: '#a9d1ff' // 100% 处的颜色
                    //             }
                    //         ],
                    //         globalCoord: false
                    //     }
                    }
                },
                itemStyle: {
                    normal: {
                        color: 'red'
                    }
                }
            }
        ]
    },
    // Disk Usage
    storage: {
        tooltip: {
            trigger: 'axis',
            axisPointer: {
                type: 'line',
                animation: false,
                label: {
                    backgroundColor: '#6a7985'
                }
            },
            formatter (params, ticket, callback) {
                let ret
                if (params[0].seriesName === 'noData') {
                    ret = '<div>No Data</div>'
                } else {
                    ret = `<div>${moment(params[0].value[0]).format('YYYY-MM-DD HH:mm:ss')}</div>`
                    params.forEach(item => {
                        ret += `<div>${item.seriesName}${item.value[2] === 'read' ? window.i18n.t('读') : window.i18n.t('写')}：${(item.value[1] || 0).toFixed(2)}KB/s</div>`
                    })
                }

                return ret
            }
        },
        // legend: {
        //     data: ['sent', 'recv']
        // },
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
                    color: '#868b97'
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
                        return `${value}KB/s`
                    }
                },
                splitLine: {
                    lineStyle: {
                        color: ['#ebf0f5'],
                        type: 'dashed'
                    }
                }
            }
        ],
        series: [
            {
                type: 'line',
                smooth: true,
                showSymbol: false,
                hoverAnimation: false,
                areaStyle: {
                    normal: {
                    }
                },
                itemStyle: {
                    normal: {
                        color: '#ffbe21'
                    }
                }
            },
            {
                type: 'line',
                smooth: true,
                showSymbol: false,
                hoverAnimation: false,
                areaStyle: {
                    normal: {
                    }
                },
                itemStyle: {
                    normal: {
                        color: 'red'
                    }
                }
            }
        ]
    }
}

/**
 * containerDetail 页面 和 containerDetailForNode 页面
 *
 * @type {Object}
 */
export const containerDetailChart = {
    cpu: {
        tooltip: {
            trigger: 'axis',
            confine: true,
            axisPointer: {
                type: 'line',
                animation: false,
                label: {
                    backgroundColor: '#6a7985'
                }
            },
            formatter (params, ticket, callback) {
                let ret = ''
                const containerName = params[0].seriesName
                if (containerName === 'noData') {
                    ret = '<div>No Data</div>'
                } else {
                    ret += '<div style="width: 450px;">'
                    ret += `<div>${moment(params[0].value[0]).format('YYYY-MM-DD HH:mm:ss')}`
                        + '<div style="text-align: left; white-space: normal;word-break: break-all;">'
                        + `${containerName}</div>`
                        + `</div>`
                    params.forEach(item => {
                        ret += `<div>CPU Usage：`
                            + `<span style="font-weight: 700; color: #30d873;">${(item.value[1]).toFixed(2)}%</span>`
                            + `</div>`
                    })
                    ret += '</div>'
                }
                return ret
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
                        return moment(value).format('HH:mm')
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
                        return `${value}%`
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
            {
                type: 'line',
                smooth: true,
                showSymbol: false,
                hoverAnimation: false,
                areaStyle: {
                    normal: {
                        color: '#30d878',
                        opacity: 0.1
                    }
                },
                lineStyle: {
                    normal: {
                        color: '#30d878'
                    }
                },
                // 折线拐点标志的样式
                itemStyle: {
                    normal: {
                        // opacity: 0,
                        color: '#868b97'
                    }
                }
            }
        ]
    },
    mem: {
        tooltip: {
            trigger: 'axis',
            confine: true,
            axisPointer: {
                type: 'line',
                animation: false,
                label: {
                    backgroundColor: '#6a7985'
                }
            },
            formatter (params, ticket, callback) {
                let ret = ''
                const containerName = params[0].seriesName
                if (containerName === 'noData') {
                    ret = '<div>No Data</div>'
                } else {
                    ret += '<div style="width: 450px;">'
                    ret += `<div>${moment(params[0].value[0]).format('YYYY-MM-DD HH:mm:ss')}`
                        + '<div style="text-align: left; white-space: normal;word-break: break-all;">'
                        + `${containerName}</div>`
                        + `</div>`
                    params.forEach(item => {
                        ret += `<div>${window.i18n.t('内存已使用')}：`
                            + `<span style="font-weight: 700; color: #3a84ff;">${(item.value[1]).toFixed(2)}MB</span>`
                            + `</div>`
                    })
                    ret += '</div>'
                }
                return ret
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
                        return moment(value).format('HH:mm')
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
                        return `${value}MB`
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
            {
                type: 'line',
                smooth: true,
                showSymbol: false,
                hoverAnimation: false,
                areaStyle: {
                    normal: {
                        color: '#3a84ff',
                        opacity: 0.1
                    }
                },
                lineStyle: {
                    normal: {
                        color: '#3a84ff'
                    }
                },
                // 折线拐点标志的样式
                itemStyle: {
                    normal: {
                        // opacity: 0,
                        color: '#868b97'
                    }
                }
            }
        ]
    },
    memInternal: {
        tooltip: {
            trigger: 'axis',
            confine: true,
            axisPointer: {
                type: 'line',
                animation: false,
                label: {
                    backgroundColor: '#6a7985'
                }
            },
            formatter (params, ticket, callback) {
                let ret = ''
                const containerName = params[0].seriesName
                if (containerName === 'noData') {
                    ret = '<div>No Data</div>'
                } else {
                    ret += '<div style="width: 450px;">'
                    ret += `<div>${moment(params[0].value[0]).format('YYYY-MM-DD HH:mm:ss')}`
                        + '<div style="text-align: left; white-space: normal;word-break: break-all;">'
                        + `${containerName}</div>`
                        + `</div>`
                    params.forEach(item => {
                        ret += `<div>Memory Usage：`
                            + `<span style="font-weight: 700; color: #3a84ff;">${(item.value[1]).toFixed(2)}%</span>`
                            + `</div>`
                    })
                    ret += '</div>'
                }
                return ret
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
                        return moment(value).format('HH:mm')
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
                        return `${value}%`
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
            {
                type: 'line',
                smooth: true,
                showSymbol: false,
                hoverAnimation: false,
                areaStyle: {
                    normal: {
                        color: '#3a84ff',
                        opacity: 0.1
                    }
                },
                lineStyle: {
                    normal: {
                        color: '#3a84ff'
                    }
                },
                // 折线拐点标志的样式
                itemStyle: {
                    normal: {
                        // opacity: 0,
                        color: '#868b97'
                    }
                }
            }
        ]
    },
    net: {
        tooltip: {
            trigger: 'axis',
            confine: true,
            axisPointer: {
                type: 'line',
                animation: false,
                label: {
                    backgroundColor: '#6a7985'
                }
            },
            formatter (params, ticket, callback) {
                let ret = ''
                const containerName = params[0].value[3]
                if (containerName === 'noData') {
                    ret = '<div>No Data</div>'
                } else {
                    ret += '<div style="width: 450px;">'
                    ret += `<div>${moment(params[0].value[0]).format('YYYY-MM-DD HH:mm:ss')}`
                        + '<div style="text-align: left; white-space: normal;word-break: break-all;">'
                        + `${containerName}</div>`
                        + `</div>`
                    params.forEach(item => {
                        ret += `<div>${item.value[2] === 'tx' ? window.i18n.t('发送') : window.i18n.t('接收')}：`
                            + `<span style="font-weight: 700; color: #30d873;">${(item.value[1]).toFixed(2)}Bytes/s</span>`
                            + `</div>`
                    })
                    ret += '</div>'
                }
                return ret
            }
        },
        legend: {
            show: false,
            data: ['发送', '接收'],
            selected: {
                '发送': true,
                '接收': true
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
                    color: '#868b97'
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
                        return `${value}Bytes/s`
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
            {
                type: 'line',
                // showSymbol: true,
                smooth: true,
                showSymbol: false,
                hoverAnimation: false,
                areaStyle: {
                    normal: {
                    }
                },
                itemStyle: {
                    normal: {
                        color: '#ffbe21'
                    }
                }
            },
            {
                type: 'line',
                // showSymbol: true,
                smooth: true,
                showSymbol: false,
                hoverAnimation: false,
                areaStyle: {
                    normal: {
                    }
                },
                itemStyle: {
                    normal: {
                        color: 'red'
                    }
                }
            }
        ]
    },
    disk: {
        tooltip: {
            trigger: 'axis',
            confine: true,
            axisPointer: {
                type: 'line',
                animation: false,
                label: {
                    backgroundColor: '#6a7985'
                }
            },
            formatter (params, ticket, callback) {
                let ret = ''
                const containerName = params[0].value[3]
                if (containerName === 'noData') {
                    ret = '<div>No Data</div>'
                } else {
                    ret += '<div style="width: 450px;">'
                    ret += `<div>${moment(params[0].value[0]).format('YYYY-MM-DD HH:mm:ss')}`
                        + '<div style="text-align: left; white-space: normal;word-break: break-all;">'
                        + `${containerName}</div>`
                        + `</div>`
                    params.forEach(item => {
                        console.error(item)
                        ret += `<div>${window.i18n.t('读')}：`
                            + `<span style="font-weight: 700; color: #30d873;">${(item.value[1]).toFixed(2)}Bytes/s</span>`
                            + `</div>`
                        ret += `<div>${window.i18n.t('写')}：`
                            + `<span style="font-weight: 700; color: #30d873;">${(item.value[2]).toFixed(2)}Bytes/s</span>`
                            + `</div>`
                    })
                    ret += '</div>'
                }
                return ret
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
                    color: '#868b97'
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
                        return `${value}Bytes/s`
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
            {
                type: 'line',
                smooth: true,
                showSymbol: false,
                hoverAnimation: false,
                areaStyle: {
                    normal: {
                        opacity: 0.1
                    }
                },
                itemStyle: {
                    normal: {
                        color: '#ffbe21'
                    }
                }
            }
        ]
    },
    diskInternal: {
        tooltip: {
            trigger: 'axis',
            confine: true,
            axisPointer: {
                type: 'line',
                animation: false,
                label: {
                    backgroundColor: '#6a7985'
                }
            },
            formatter (params, ticket, callback) {
                let ret = ''
                const containerName = params[0].seriesName
                if (containerName === 'noData') {
                    ret = '<div>No Data</div>'
                } else {
                    ret += '<div style="width: 450px;">'
                    ret += `<div>${moment(params[0].value[0]).format('YYYY-MM-DD HH:mm:ss')}`
                    params.forEach(item => {
                        ret += `<div>${item.seriesName}Disk Usage：${(item.value[1]).toFixed(2)}%</div>`
                    })
                    ret += '</div>'
                }
                return ret
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
                    color: '#868b97'
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
                        return `${value}%`
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
            {
                type: 'line',
                smooth: true,
                showSymbol: false,
                hoverAnimation: false,
                areaStyle: {
                    normal: {
                        opacity: 0.1
                    }
                },
                itemStyle: {
                    normal: {
                        color: '#ffbe21'
                    }
                }
            }
        ]
    }
}

/**
 * instanceDetail 页面
 *
 * @type {Object}
 */
export const instanceDetailChart = {
    cpu: {
        tooltip: {
            trigger: 'axis',
            confine: true,
            axisPointer: {
                type: 'line',
                animation: false,
                label: {
                    backgroundColor: '#6a7985'
                }
            },
            formatter (params, ticket, callback) {
                let ret = '<div style="width: 300px;">'
                ret += `<div>${moment(params[0].value[0]).format('YYYY-MM-DD HH:mm:ss')}</div>`
                if (params[0].seriesName === 'noData') {
                    ret = '<div>No Data</div>'
                } else {
                    params.forEach(item => {
                        ret += `<div style="${STYLE_STR} color: ${item.value[2]}">`
                            + `${item.seriesName}：<span style="font-weight: 700; color: #30d873;">`
                            + `${(item.value[1]).toFixed(2)}%</span></div>`
                    })
                    ret += '</div>'
                }
                return ret
            }
        },
        grid: {
            // top: '4%',
            // left: '0',
            // right: '5%',
            // bottom: '3%',
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
                        return moment(value).format('HH:mm')
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
                        return `${value}%`
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
        ]
    },
    memInternal: {
        tooltip: {
            trigger: 'axis',
            confine: true,
            axisPointer: {
                type: 'line',
                animation: false,
                label: {
                    backgroundColor: '#6a7985'
                }
            },
            formatter (params, ticket, callback) {
                let ret = '<div style="width: 300px;">'
                ret += `<div>${moment(params[0].value[0]).format('YYYY-MM-DD HH:mm:ss')}</div>`
                if (params[0].seriesName === 'noData') {
                    ret = '<div>No Data</div>'
                } else {
                    params.forEach(item => {
                        ret += `<div style="${STYLE_STR} color: ${item.value[2]}">`
                            + `${item.seriesName}：<span style="font-weight: 700; color: #3a84ff;">`
                            + `${(item.value[1]).toFixed(2)}%</span></div>`
                    })
                    ret += '</div>'
                }
                return ret
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
                        return moment(value).format('HH:mm')
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
                        return `${value}%`
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
        ]
    },
    mem: {
        tooltip: {
            trigger: 'axis',
            confine: true,
            axisPointer: {
                type: 'line',
                animation: false,
                label: {
                    backgroundColor: '#6a7985'
                }
            },
            formatter (params, ticket, callback) {
                let ret = '<div style="width: 300px;">'
                ret += `<div>${moment(params[0].value[0]).format('YYYY-MM-DD HH:mm:ss')}</div>`
                if (params[0].seriesName === 'noData') {
                    ret = '<div>No Data</div>'
                } else {
                    params.forEach(item => {
                        ret += `<div style="${STYLE_STR} color: ${item.value[2]}">`
                            + `${item.seriesName}：<span style="font-weight: 700; color: #3a84ff;">`
                            + `${(item.value[1]).toFixed(2)}MB</span></div>`
                    })
                    ret += '</div>'
                }
                return ret
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
                        return moment(value).format('HH:mm')
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
                        return `${value}MB`
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
        ]
    }
}
