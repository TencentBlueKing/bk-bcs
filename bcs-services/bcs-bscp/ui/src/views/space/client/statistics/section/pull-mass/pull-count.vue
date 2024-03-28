<template>
  <Card title="拉取数量趋势" :height="360">
    <template #head-suffix>
      <bk-select v-model="selectTime" class="time-selector" :clearable="false">
        <bk-option v-for="item in selectorTimeList" :id="item.value" :key="item.value" :name="item.label" />
      </bk-select>
    </template>
    <div ref="canvasRef" class="canvas-wrap"></div>
  </Card>
</template>

<script lang="ts" setup>
  import { ref, onMounted } from 'vue';
  import Card from '../../components/card.vue';
  import { DualAxes } from '@antv/g2plot';

  const canvasRef = ref<HTMLElement>();

  const selectTime = ref(7);

  const selectorTimeList = [
    {
      value: 7,
      label: '近7天',
    },
    {
      value: 15,
      label: '近15天',
    },
    {
      value: 30,
      label: '近30天',
    },
  ];

  const data = [
    // 客户端拉取趋势
    {
      date: '2024/03/14',
      type: 'sidecar',
      count: 4,
    },
    {
      date: '2024/03/14',
      type: 'sdk',
      count: 3,
    },
    {
      date: '2024/03/15',
      type: 'sidecar',
      count: 8,
    },
    {
      date: '2024/03/15',
      type: 'sdk',
      count: 3,
    },
    {
      date: '2024/03/16',
      type: 'sidecar',
      count: 1,
    },
    {
      date: '2024/03/16',
      type: 'sdk',
      count: 2,
    },
  ];
  const transformData = [
    { date: '2024/03/14', total: 10 },
    { date: '2024/03/15', total: 7 },
    { date: '2024/03/16', total: 14 },
  ];

  onMounted(() => {
    const dualAxes = new DualAxes(canvasRef.value!, {
      data: [data, transformData],
      xField: 'date',
      yField: ['count', 'total'],
      geometryOptions: [
        {
          geometry: 'column',
          isGroup: true,
          seriesField: 'type',
        },
        {
          geometry: 'line',
          lineStyle: {
            lineWidth: 2,
          },
        },
      ],
    });

    dualAxes.render();
  });
</script>

<style scoped lang="scss">
  .canvas-wrap {
    height: 100%;
  }
  .time-selector {
    margin-left: 16px;
    width: 88px;
    :deep(.bk-input) {
      height: 26px;
    }
  }
</style>
