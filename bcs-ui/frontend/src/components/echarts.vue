<template>
  <div class="echarts" />
</template>

<script>
import * as echarts from 'echarts';

export default {
  props: {
    options: Object,
    autoResize: {
      type: Boolean,
      default: true,
    },
  },
  watch: {
    options(newValue) {
      this.chart?.setOption(newValue);
    },
  },
  mounted() {
    this.init();
  },
  activated() {
    if (this.autoResize) {
      this.chart?.resize();
    }
  },
  beforeDestroy() {
    this.destroy();
  },
  methods: {
    showLoading(type, options) {
      this.chart?.showLoading(type, options);
    },
    hideLoading() {
      this.chart?.hideLoading();
    },
    init() {
      if (this.chart) {
        return;
      }

      const chart = echarts.init(this.$el);

      chart.setOption(this.options || {}, true);

      if (this.autoResize) {
        this.resizeObserver = new ResizeObserver((entries) => {
          window.requestAnimationFrame(() => {
            for (const entry of entries) {
              if (entry.target === this.$el) {
                this.chart?.resize();
              }
            }
          });
        });
        this.resizeObserver.observe(this.$el);
      }
      this.chart = chart;
    },
    resize() {
      this.chart?.resize();
    },
    destroy() {
      this.resizeObserver?.disconnect();
      this.resizeObserver?.unobserve(this.$el);
      this.resizeObserver = null;
      this.chart?.dispose();
      this.chart = null;
    },
    refresh() {
      if (this.chart) {
        this.destroy();
        this.init();
      }
    },
    setOption(option) {
      this.chart?.setOption(option);
    },
  },
};
</script>

<style scoped>
.echarts {
  width: 600px;
  height: 400px;
}
</style>

