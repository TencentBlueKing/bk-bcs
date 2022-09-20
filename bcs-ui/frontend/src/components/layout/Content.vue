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
    <div class="biz-content-wrapper content">
      <slot></slot>
    </div>
  </div>
</template>
<script lang="ts">
import BcsContentHeader from '@/views/content-header.vue';
import { defineComponent } from '@vue/composition-api';
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
    return {};
  },
});
</script>
<style lang="postcss" scoped>
.content {
  padding: 20px;
}
</style>
