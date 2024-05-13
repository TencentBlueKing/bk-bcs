<template>
  <div class="g2-tooltip" ref="tooltipRef">
    <!-- 标题容器，会自己填充 -->
    <div class="g2-tooltip-title">
    </div>
    <slot name="title" />

    <!-- 列表容器，会自己填充 -->
    <ul class="g2-tooltip-list"></ul>
    <!-- 自定义尾部-->
    <li class="g2-tooltip-bottom" @click="emits('jump')">
      <span class="bk-bscp-icon icon-help-fill"></span>
      <span class="g2-tooltip-name">{{ $t('查看数据详情') }}</span>
    </li>
  </div>
</template>

<script lang="ts" setup>
  import { ref } from 'vue';

  const emits = defineEmits(['jump']);

  const tooltipRef = ref();

  const getDom = () => {
    return tooltipRef.value;
  };
  defineExpose({
    getDom,
  });
</script>

<style lang="scss">
  .g2-tooltip {
    position: absolute;
    min-width: 160px;
    .g2-tooltip-bottom {
      display: flex;
      align-items: center;
      justify-content: center;
      height: 40px;
      border-top: 1px solid #dcdee5;
      color: #3a84ff;
      cursor: pointer;
    }
    .g2-tooltip-list-item {
      .g2-tooltip-value {
        margin-left: 0 !important;
      }
    }
  }
</style>
