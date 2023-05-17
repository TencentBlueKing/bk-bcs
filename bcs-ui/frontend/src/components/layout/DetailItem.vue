<template>
  <div :class="['bcs-detail-item', type]">
    <div class="bcs-detail-item-label">
      {{label}}
      <span v-if="type === 'horizontal'">:</span>
      <span v-if="desc" v-bk-tooltips="desc">
        <i class="bcs-icon bcs-icon-question-circle"></i>
      </span>
    </div>
    <div class="bcs-detail-item-content">
      <slot></slot>
    </div>
  </div>
</template>
<script lang="ts">
import { defineComponent } from 'vue';

export default defineComponent({
  name: 'DetailItem',
  props: {
    label: {
      type: String,
      default: '',
    },
    type: {
      type: String,
      default: 'horizontal',
      validator(v: string) {
        return ['horizontal', 'vertical'].includes(v);
      },
    },
    desc: {
      type: String,
      default: '',
    },
  },
});
</script>
<style lang="postcss" scoped>
.bcs-detail-item {
  font-size: 14px;
  &.horizontal {
    display: flex;
    align-items: flex-start;
    justify-content: flex-end;
    line-height: 32px;
    .bcs-detail-item-label {
      display: flex;
      align-items: center;
      justify-content: flex-end;
      flex: 0 0 auto;
      color: #b2b5bd;
      margin-right: 14px;
      text-align: right;
    }
  }
  &.vertical {
    .bcs-detail-item-label {
      display: flex;
      text-align: left;
      margin-bottom: 14px;
    }
  }

  &-label {
    padding-right: 2px;
    white-space: pre;
    word-break: keep-all;
  }
  &-content {
    color: #63656e;
    flex: 1;
    word-break: break-all;
  }
}
</style>
