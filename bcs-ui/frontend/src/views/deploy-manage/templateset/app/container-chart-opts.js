/*
* Tencent is pleased to support the open source community by making
* 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition) available.
*
* Copyright (C) 2021 THL A29 Limited, a Tencent company.  All rights reserved.
*
* 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition) is licensed under the MIT License.
*
* License for 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition):
*
* ---------------------------------------------------
* Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated
* documentation files (the "Software"), to deal in the Software without restriction, including without limitation
* the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and
* to permit persons to whom the Software is furnished to do so, subject to the following conditions:
*
* The above copyright notice and this permission notice shall be included in all copies or substantial portions of
* the Software.
*
* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO
* THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF
* CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
* IN THE SOFTWARE.
*/

export function createChartOption() {
  return {
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'line',
        animation: false,
        label: {
          backgroundColor: '#6a7985',
        },
      },
    },
    grid: {
      show: false,
      top: '4%',
      left: '4%',
      right: '5%',
      bottom: '3%',
      containLabel: true,
    },
    xAxis: [
      {
        type: 'time',
        boundaryGap: false,
        axisLine: {
          show: true,
          lineStyle: {
            color: '#dde4eb',
          },
        },
        axisTick: {
          alignWithLabel: true,
          length: 5,
          lineStyle: {
            color: '#ebf0f5',
          },
        },
        axisLabel: {
          color: '#868b97',
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
            type: 'dashed',
          },
        },
      },
    ],
    yAxis: [
      {
        boundaryGap: [0, '2%'],
        type: 'value',
        axisLine: {
          show: true,
          lineStyle: {
            color: '#dde4eb',
          },
        },
        axisTick: {
          alignWithLabel: true,
          length: 0,
          lineStyle: {
            color: 'red',
          },
        },
        axisLabel: {
          color: '#868b97',
          formatter(value) {
            return `${value.toFixed(1)}%`;
          },
        },
        splitLine: {
          show: true,
          lineStyle: {
            color: ['#ebf0f5'],
            type: 'dashed',
          },
        },
      },
    ],
    series: [],
  };
}
