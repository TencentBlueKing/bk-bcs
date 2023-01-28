<template>
  <div class="biz-content">
    <BcsContentHeader
      :title="title"
      :desc="desc"
      :hide-back="hideBack">
      <template #right>
        <slot name="header-right"></slot>
      </template>
    </BcsContentHeader>
    <div class="biz-content-wrapper content" ref="contentRef">
      <slot></slot>
    </div>
  </div>
</template>
<script lang="ts">
import BcsContentHeader from '@/components/layout/Header.vue';
import { defineComponent, ref } from '@vue/composition-api';
export default defineComponent({
  name: 'LayoutContent',
  components: {
    BcsContentHeader,
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
  },
  setup() {
    const contentRef = ref<any>(null);
    const handleScollTop = () => {
      contentRef.value.scrollTop = 0;
    };
    return {
      contentRef,
      handleScollTop,
    };
  },
});
</script>
<style lang="postcss" scoped>
.content {
  padding: 20px;
}
</style>
