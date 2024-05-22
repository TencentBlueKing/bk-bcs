<template>
  <div class="g2-tooltip" ref="tooltipRef">
    <div v-if="needDownIcon" class="down-wrap">
      <div class="icon-wrap">
        <span class="action-icon bk-bscp-icon icon-download" />
      </div>
      <span v-if="down" class="content">{{ `${$t('下钻')}: ${down}` }}</span>
    </div>
    <!-- 标题容器，会自己填充 -->
    <div class="g2-tooltip-title"></div>
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

  defineProps<{
    needDownIcon?: boolean;
    down?: string;
  }>();
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
    .down-wrap {
      display: flex;
      position: absolute;
      top: -24px;
      left: 0;
      height: 20px;
      opacity: 0.96;
      background-image: linear-gradient(180deg, #ffffffe6 0%, #fcfcfce6 100%);
      border: 1px solid #ffffff;
      box-shadow: 0 0 6px 1px #00000029;
      border-radius: 2px;
      .icon-wrap {
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 16px;
        color: #7594ef;
        width: 20px;
      }
      .content {
        display: flex;
        align-items: center;
        justify-content: center;
        margin-right: 8px;
        font-size: 12px;
        color: #63656e;
      }
    }
  }
</style>
