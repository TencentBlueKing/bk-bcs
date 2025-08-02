<template>
  <div class="bcs-footer">
    <div v-if="$INTERNAL">
      <div class="mb5 link">
        <a :href="config.helperLink">{{ config.i18n.helperText }}</a> |
        <a :href="PAAS_HOST" target="_blank">{{ $t('blueking.desktop') }}</a>
      </div>
      <p>
        {{ config.footerCopyright }}
      </p>
    </div>
    <div v-else>
      <div class="mb5 link">
        <span v-html="footerHTML"></span>
      </div>
      <p>{{ config.footerCopyrightContent }}</p>
    </div>
  </div>
</template>
<script lang="ts">
import { computed, defineComponent } from 'vue';
import xss from 'xss';

import usePlatform from '@/composables/use-platform';

export default defineComponent({
  name: 'BcsFooter',
  setup() {
    const { config } = usePlatform();
    const footerHTML = computed(() => xss(config.i18n.footerInfoHTML));
    return {
      config,
      PAAS_HOST: window.PAAS_HOST,
      version: localStorage.getItem('__bcs_latest_version__'),
      footerHTML,
    };
  },
});
</script>
<style lang="postcss" scoped>
.bcs-footer {
  font-size: 12px;
  color: #b7c0ca;
  width: 100%;
  text-align: center;
  line-height: 20px;
  padding-top: 25px;
  .link a {
    color: #3a84ff;
  }
}

>>> .link-item {
  color: #3a84ff;
}
</style>
