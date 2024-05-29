<template>
  <div ref="canvasRef" class="canvas-wrap">
    <Tooltip ref="tooltipRef" @jump="emits('jump', labelValue)" />
  </div>
</template>

<script lang="ts" setup>
  import { ref, onMounted, watch } from 'vue';
  import { Pie, Sunburst } from '@antv/g2plot';
  import Tooltip from '../../../components/tooltip.vue';
  import { IClientLabelItem } from '../../../../../../../../types/client';
  import { useI18n } from 'vue-i18n';

  interface ISunburstDataType {
    name: string;
    children: ISunburstChildrenType[];
  }

  interface ISunburstChildrenType {
    name: string;
    value: number;
    percent: number;
    children?: ISunburstChildrenType[];
  }

  const { t } = useI18n();

  const props = defineProps<{
    data: IClientLabelItem[];
    bkBizId: string;
    appId: number;
    isShowSunburst: boolean;
  }>();

  const emits = defineEmits(['jump', 'drillDown']);

  let piePlot: Pie | null;
  let sunburstPlot: Sunburst | null;
  const canvasRef = ref<HTMLElement>();
  const tooltipRef = ref();
  const labelValue = ref('');
  const sunburstData = ref<ISunburstDataType>();

  watch(
    () => props.data,
    () => {
      convertToTree(props.data);
      if (props.isShowSunburst) {
        if (sunburstPlot) {
          sunburstPlot.changeData(sunburstData.value);
        } else {
          piePlot!.destroy();
          piePlot = null;
          initSunburstChart();
        }
      } else {
        if (piePlot) {
          piePlot.changeData(props.data);
        } else {
          sunburstPlot!.destroy();
          sunburstPlot = null;
          initPieChart();
        }
      }
    },
  );

  onMounted(() => {
    convertToTree(props.data);
    props.isShowSunburst ? initSunburstChart() : initPieChart();
  });

  const initPieChart = () => {
    piePlot = new Pie(canvasRef.value!, {
      data: props.data,
      angleField: 'count',
      colorField: 'primary_val',
      radius: 1,
      padding: [40, 300, 40, 20],
      label: {
        type: 'inner',
        offset: '-30%',
        content: ({ percent }) => `${(percent * 100).toFixed(1)}%`,
        style: {
          fontSize: 14,
          textAlign: 'center',
        },
        autoRotate: false,
      },
      tooltip: {
        fields: ['count', 'percent'],
        showTitle: true,
        title: 'primary_val',
        container: tooltipRef.value?.getDom(),
        enterable: true,
        customItems: (originalItems: any[]) => {
          labelValue.value = originalItems[0].title;
          originalItems[0].name = t('客户端数量');
          originalItems[1].name = t('占比');
          originalItems[1].value = `${(originalItems[1].value * 100).toFixed(1)}%`;
          return originalItems;
        },
      },
      interactions: [{ type: 'element-highlight' }],
      state: {
        active: {
          style: {
            stroke: '#ffffff',
          },
        },
      },
      legend: {
        position: 'right',
        offsetX: -200,
        layout: 'horizontal',
        flipPage: false,
        maxWidth: 300,
        reversed: true,
      },
    });
    piePlot.on('plot:click', (e: any) => {
      if (!e.data) return;
      emits('drillDown', e.data.data as IClientLabelItem);
    });
    piePlot.render();
  };

  const initSunburstChart = () => {
    sunburstPlot = new Sunburst(canvasRef.value!, {
      data: sunburstData.value,
      color: ['#2C2599', '#FFA66B', '#85CCA8', '#3E96C2'],
      interactions: [{ type: 'element-highlight' }],
      state: {
        active: {
          style: {
            stroke: '#ffffff',
          },
        },
      },
      label: {
        content: ({ data }) => `${(data.percent * 100).toFixed(1)}%`,
        style: {
          fontSize: 14,
          textAlign: 'center',
        },
        autoRotate: false,
      },
      legend: {
        position: 'right',
        layout: 'vertical',
      },
      tooltip: {
        fields: ['value', 'name'],
        showTitle: true,
        title: 'name',
        container: tooltipRef.value?.getDom(),
        enterable: true,
        customItems: (originalItems: any[]) => {
          originalItems[0].marker = false;
          originalItems[0].name = t('客户端数量');
          originalItems[1].name = t('占比');
          originalItems[1].value = `${(originalItems[1].data.data.percent * 100).toFixed(1)}%`;
          return originalItems;
        },
      },
      hierarchyConfig: {
        padding: 0.003,
      },
    });
    sunburstPlot.render();
  };

  const convertToTree = (data: IClientLabelItem[]) => {
    const tree: ISunburstChildrenType[] = [];
    let sunburstTitle = '';
    data.forEach((item) => {
      const { primary_val, primary_key, foreign_val, percent, count } = item;
      let typeNode = tree.find((node) => node.name === primary_val);
      if (!typeNode) {
        typeNode = { name: primary_val!, children: [], percent: 0, value: 0 };
        tree.push(typeNode);
      }
      const versionNode: ISunburstChildrenType = {
        name: foreign_val,
        percent,
        value: count,
      };
      typeNode.children?.push(versionNode);
      typeNode.value += count;
      typeNode.percent += percent;
      sunburstTitle = primary_key;
    });
    sunburstData.value = {
      name: sunburstTitle,
      children: tree,
    };
  };
</script>

<style lang="scss" scoped>
  :deep(.g2-tooltip) {
    visibility: hidden;
    .g2-tooltip-title {
      padding-left: 16px;
      font-size: 14px;
    }
    .g2-tooltip-list-item:nth-child(2) {
      .g2-tooltip-marker {
        display: none !important;
      }
      .g2-tooltip-name {
        margin-left: 16px;
      }
    }
    .g2-tooltip-list-item:nth-child(1) {
      .g2-tooltip-marker {
        position: absolute;
        top: 15px;
      }
      .g2-tooltip-name {
        margin-left: 16px;
      }
    }
  }
</style>

<style lang="scss">
  .canvas-wrap {
    position: relative;
    display: flex;
    align-items: center;
    height: 100%;
  }
</style>
