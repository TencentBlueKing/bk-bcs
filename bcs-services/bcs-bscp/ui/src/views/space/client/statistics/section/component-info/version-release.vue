<template>
  <Card title="组件版本分布" :height="416">
    <div ref="canvasRef" id="canvas"></div>
  </Card>
</template>

<script lang="ts" setup>
  import { Sunburst } from '@antv/g2plot';
  import Card from '../../components/card.vue';
  import { ref, onMounted } from 'vue';
  const canvasRef = ref<HTMLElement>();
  onMounted(() => {
    fetch('https://gw.alipayobjects.com/os/antfincdn/ryp44nvUYZ/coffee.json')
      .then((data) => data.json())
      .then((data) => {
        const plot = new Sunburst(canvasRef.value!, {
          data,
        });
        plot.render();
      });
  });
</script>

<style scoped lang="scss">
  .canvas-wrap {
    height: 100%;
  }
</style>
