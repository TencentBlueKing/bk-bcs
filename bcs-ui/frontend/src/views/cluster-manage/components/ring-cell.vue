<template>
  <Ring
    :percent="percent"
    :size="size"
    :ext-cls="extCls"
    :text="text"
    :text-style="textStyle"
    :stroke-width="strokeWidth"
    :fill-width="fillWidth"
    :fill-color="fillColor">
  </Ring>
</template>

<script>
import Ring from '@/components/ring';

export default {
  components: {
    Ring,
  },
  props: {
    percent: {
      type: [Number, String],
      default: 0,
    },
    size: {
      type: Number,
      default: 45,
    },
    extCls: {
      type: String,
      default: 'node-ring',
    },
    text: {
      type: String,
      default: 'always',
    },
    textStyle: {
      type: Object,
      default: () => ({
        fontSize: '9px',
        color: '#737987',
        fontWeight: '700',
      }),
    },
    strokeWidth: {
      type: Number,
      default: 4,
    },
    fillWidth: {
      type: Number,
      default: 4,
    },
    fillColor: {
      type: String,
      default: '#3a84ff',
    },
  },
};
</script>
