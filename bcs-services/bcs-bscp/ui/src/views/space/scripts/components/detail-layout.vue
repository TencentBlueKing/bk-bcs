<template>
  <section :class="['script-detail-layout', { 'show-notice': showNotice }]">
    <header>
      <div class="nav-title">
        <ArrowsLeft class="arrow-icon" @click="emits('close')" />
        <span class="title">{{ props.name }}</span>
      </div>
    </header>
    <div :class="['layout-content', { 'without-footer': !props.showFooter }]">
      <slot name="content"></slot>
    </div>
    <footer v-if="props.showFooter" class="layout-footer">
      <slot name="footer"></slot>
    </footer>
  </section>
</template>
<script setup lang="ts">
  import { ArrowsLeft } from 'bkui-vue/lib/icon';
  import { storeToRefs } from 'pinia';
  import useGlobalStore from '../../../../store/global';

  const { showNotice } = storeToRefs(useGlobalStore());

  const props = withDefaults(
    defineProps<{
      name: string;
      showFooter?: boolean;
    }>(),
    {
      showFooter: true,
    },
  );

  const emits = defineEmits(['close']);
</script>
<style lang="scss" scoped>
  .script-detail-layout {
    position: fixed;
    top: 52px;
    left: 0;
    width: 100%;
    height: calc(100vh - 52px);
    background: #ffffff;
    z-index: 2000;
    &.show-notice {
      top: 92px;
      height: calc(100vh - 92px);
    }
    .nav-title {
      display: flex;
      align-items: center;
      position: relative;
      padding: 0 24px;
      background: #ffffff;
      box-shadow: 0 3px 4px 0 #0000000a;
      z-index: 1;
    }
    .arrow-icon {
      font-size: 24px;
      color: #3a84ff;
      cursor: pointer;
    }
    .title {
      padding: 14px 0;
      font-size: 16px;
      line-height: 24px;
      color: #313238;
    }
    .layout-content {
      height: calc(100% - 100px);
      &.without-footer {
        height: calc(100% - 52px);
      }
    }
    .layout-footer {
      padding: 8px 48px;
      background: #fafbfd;
      box-shadow: 0 -1px 0 0 #dcdee5;
    }
  }
</style>
