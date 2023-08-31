import { formatBytes, formatDate } from '@/common/util';
import { Decimal } from 'decimal.js';

export default function (unit) {
  const axisLabel = (value) => {
    if (!value) return '';

    let label = value;
    switch (unit) {
      // 字节类型纵坐标
      case 'byte':
        label = `${formatBytes(value, 2)}`;
        break;
        // 百分比类型纵坐标
      case 'percent':
        // eslint-disable-next-line no-case-declarations
        const valueLen = String(value).length > 3 ? 3 : String(value).length;
        label = `${new Decimal(value).toPrecision(valueLen)}%`;
        break;
      case 'percent-number':
        label = `${Number(value).toFixed(2)}%`;
        break;
      case 'number':
        label = Number(value).toFixed(2);
        break;
    }
    return label;
  };
  return {
    tooltip: {
      trigger: 'axis',
      confine: true,
      axisPointer: {
        type: 'line',
        animation: false,
        label: {
          backgroundColor: '#6a7985',
        },
      },
      extraCssText: 'white-space: break-spaces;',
      formatter: (params) => {
        const date = formatDate(params?.[0]?.axisValue * 1000, 'YYYY-MM-DD hh:mm:ss');
        let ret = `<div>${date}</div>`;
        params.forEach((p) => {
          ret += `<div>${p.seriesName}：${axisLabel(p.value?.[1])}</div>`;
        });

        return ret;
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
          formatter: (value) => {
            const time = formatDate(value * 1000, 'hh:mm');
            const date = formatDate(value * 1000, 'MM-DD');
            return `${time}\n${date}`;
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
            return axisLabel(value);
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
  };
}
