<template>
  <div class="bcs-footer">
    <div v-if="$INTERNAL">
      <div class="mb5 link">
        <a href="wxwork://message?uin=8444252571319680">{{ $t('blueking.onCall') }}</a> |
        <a :href="PAAS_HOST" target="_blank">{{ $t('blueking.desktop') }}</a>
      </div>
      <p>
        Copyright © 2012 Tencent BlueKing. All Rights Reserved. {{ version }}
      </p>
    </div>
    <div v-else>
      <div class="mb5 link">
        <a href="https://wpa1.qq.com/KziXGWJs?_type=wpa&qidian=true" target="_blank">{{ $t('blueking.support') }}</a> |
        <a href="https://bk.tencent.com/s-mart/community/" target="_blank">{{ $t('blueking.community') }}</a> |
        <a href="https://bk.tencent.com/index/" target="_blank">{{ $t('blueking.website') }}</a>
      </div>
      <p>Copyright © 2012 Tencent BlueKing. All Rights Reserved. {{ version }}</p>
    </div>
  </div>
</template>
<script lang="ts">
import { defineComponent } from 'vue';

export default defineComponent({
  name: 'BcsFooter',
  setup() {
    return {
      PAAS_HOST: window.PAAS_HOST,
      version: localStorage.getItem('__bcs_latest_version__'),
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
</style>
