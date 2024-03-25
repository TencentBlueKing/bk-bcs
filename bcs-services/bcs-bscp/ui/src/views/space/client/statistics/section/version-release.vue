<template>
  <div>
    <SectionTitle :title="'配置版本发布'" />
    <Card title="客户端配置版本" :height="344">
      <template #head-suffix>
        <TriggerBtn v-model:currentType="currentType" style="margin-left: 16px" />
      </template>
      <div v-if="currentType === 'pie'" ref="pieCanvasRef" class="canvas-wrap"></div>
      <div v-else-if="currentType === 'column'" ref="columnCanvasRef" class="canvas-wrap"></div>
    </Card>
  </div>
</template>

<script lang="ts" setup>
  import { ref, onMounted, watch, nextTick } from 'vue';
  import { Pie, Column } from '@antv/g2plot';
  import SectionTitle from '../components/section-title.vue';
  import Card from '../components/card.vue';
  import TriggerBtn from '../components/trigger-btn.vue';

  const pieCanvasRef = ref<HTMLElement>();
  const columnCanvasRef = ref<HTMLElement>();
  const currentType = ref('pie');
  let piePlot: Pie;
  let columnPlot: Column;
  // 准备数据
  const data = [
    { type: '分类一', value: 27 },
    { type: '分类二', value: 25 },
    { type: '分类三', value: 18 },
    { type: '分类四', value: 15 },
    { type: '分类五', value: 10 },
    { type: '其他', value: 5 },
  ];

  watch(
    () => currentType.value,
    (val) => {
      if (val === 'pie') {
        nextTick(() => {
          piePlot = new Pie(pieCanvasRef.value!, {
            appendPadding: 10,
            data,
            angleField: 'value',
            colorField: 'type',
            radius: 0.9,
            label: {
              type: 'inner',
              offset: '-30%',
              content: ({ percent }) => `${(percent * 100).toFixed(0)}%`,
              style: {
                fontSize: 14,
                textAlign: 'center',
              },
            },
            interactions: [{ type: 'element-active' }],
          });
          piePlot.render();
        });
      } else if (val === 'column') {
        nextTick(() => {
          columnPlot = new Column(columnCanvasRef.value!, {
            data,
            xField: 'type',
            yField: 'value',
            label: {
              // 可手动配置 label 数据标签位置
              position: 'middle', // 'top', 'bottom', 'middle',
              // 配置样式
              style: {
                fill: '#FFFFFF',
                opacity: 0.6,
              },
            },
            xAxis: {
              label: {
                autoHide: true,
                autoRotate: false,
              },
            },
            meta: {
              type: {
                alias: '类别',
              },
              sales: {
                alias: '销售额',
              },
            },
          });
          columnPlot.render();
        });
      }
    },
    { immediate: true },
  );

  onMounted(() => {});
</script>

<style scoped lang="scss">
  .canvas-wrap {
    height: 100%;
  }
</style>
