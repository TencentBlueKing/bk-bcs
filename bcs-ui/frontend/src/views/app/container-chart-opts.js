/**
 * container 图表配置
 *
 * @return {Object} node-overview 图表配置
 */

export function createChartOption (ctx) {
    return {
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
                    color: '#868b97'
                    // formatter (value, index) {
                    //     // return moment(parseInt(value + '000', 10)).format('HH:mm')
                    //     if (String(parseInt(value, 10)).length === 10) {
                    //         value = parseInt(value, 10) + '000'
                    //     }
                    //     return moment(parseInt(value, 10)).format('HH:mm')
                    // }
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
    }
}
