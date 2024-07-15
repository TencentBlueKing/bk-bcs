<template>
  <div class="relative flex flex-col flex-1 h-full">
    <slot name="header">
      <ContentHeader
        :title="title"
        :desc="desc"
        :hide-back="hideBack"
        :tabs="tabs"
        :active="activeTab"
        :cluster-id="clusterId"
        :namespace="namespace"
        v-if="title"
        @tab-change="handleTabChange">
        <template #default>
          <slot name="title"></slot>
        </template>
        <template #right>
          <slot name="header-right"></slot>
        </template>
      </ContentHeader>
    </slot>
    <div class="w-full flex-1 overflow-auto" :style="{ padding: `${padding}px` }" ref="contentRef">
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
    padding: {
      type: Number,
      default: 20,
    },
  },
  emits: ['tab-change'],
  setup(props, ctx) {
    const contentRef = ref<any>(null);

    const handleScrollTop = () => {
      contentRef.value.scrollTop = 0;
    };
    const handleTabChange = (item) => {
      ctx.emit('tab-change', item);
    };

    return {
      contentRef,
      handleScrollTop,
      handleTabChange,
    };
  },
});
</script>
