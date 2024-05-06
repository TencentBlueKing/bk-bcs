<template>
  <div class="biz-content">
    <ContentHeader
      :title="title"
      :desc="desc"
      :hide-back="hideBack"
      :tabs="tabs"
      :active="activeTab"
      :cluster-id="clusterId"
      :namespace="namespace"
      @tab-change="handleTabChange">
      <template #right>
        <slot name="header-right"></slot>
      </template>
    </ContentHeader>
    <div class="biz-content-wrapper content" ref="contentRef">
      <slot></slot>
    </div>
  </div>
</template>
<script lang="ts">
import { defineComponent, PropType, ref } from 'vue';

import ContentHeader from '@/components/layout/Header.vue';
export default defineComponent({
  name: 'LayoutContent',
  components: {
    ContentHeader,
  },
  props: {
    title: {
      type: String,
      default: '',
    },
    desc: {
      type: String,
      default: '',
    },
    hideBack: {
      type: Boolean,
      default: false,
    },
    tabs: {
      type: Array as PropType<{name: string;displayName: string}[]>,
      default: () => ([]),
    },
    activeTab: {
      type: String,
      default: '',
    },
    clusterId: {
      type: String,
      default: '',
    },
    namespace: {
      type: String,
      default: '',
    },
  },
  emits: ['tab-change'],
  setup(props, ctx) {
    const contentRef = ref<any>(null);
    const handleScollTop = () => {
      contentRef.value.scrollTop = 0;
    };
    const handleTabChange = (item) => {
      ctx.emit('tab-change', item);
    };
    return {
      contentRef,
      handleScollTop,
      handleTabChange,
    };
  },
});
</script>
<style lang="postcss" scoped>
.content {
  padding: 20px 24px 0px 24px;
}
</style>
