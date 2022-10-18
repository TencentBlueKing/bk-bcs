<!-- eslint-disable vue/multi-word-component-names -->
<!-- eslint-disable vue/no-mutating-props -->
<template>
  <div class="biz-actions">
    <a
      :href="PROJECT_CONFIG.doc.contact"
      class="bk-text-button"
      v-bk-tooltips.top="$t('蓝鲸容器助手')"
      v-if="$INTERNAL">{{$t('联系我们')}}</a>
    <a :href="PROJECT_CONFIG.doc.help" target="_blank" class="bk-text-button">{{$t('帮助')}}</a>
    <slot></slot>
  </div>
</template>

<script>

export default {};
</script>

<style>
    @import './index.css';
</style>
