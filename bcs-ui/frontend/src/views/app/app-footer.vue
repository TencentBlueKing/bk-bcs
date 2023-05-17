<template>
  <div class="bcs-footer">
    <div v-if="$INTERNAL">
      <div class="mb5 link">
        <a href="wxwork://message?uin=8444252571319680">{{ $t('联系BK助手') }}</a> |
        <a :href="PAAS_HOST" target="_blank">{{ $t('蓝鲸桌面') }}</a>
      </div>
      <p>
        Copyright © 2012-{{(new Date()).getFullYear()}} Tencent BlueKing. All Rights Reserved
      </p>
    </div>
    <div v-else>
      <div class="mb5 link">
        <a href="https://wpa1.qq.com/KziXGWJs?_type=wpa&qidian=true" target="_blank">{{ $t('技术支持') }}</a> |
        <a href="https://bk.tencent.com/s-mart/community/" target="_blank">{{ $t('社区论坛') }}</a> |
        <a href="https://bk.tencent.com/index/" target="_blank">{{ $t('产品官网') }}</a>
      </div>
      <p>Copyright © 2012-{{(new Date()).getFullYear()}} Tencent BlueKing. All Rights Reserved. Todo</p>
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
