<template>
  <div class="app-footer">
    <div class="links">
      <a href="https://wpa1.qq.com/KziXGWJs?_type=wpa&qidian=true" target="_blank">{{ support }}</a> |
      <a href="https://bk.tencent.com/s-mart/community/" target="_blank">{{ forum }}</a> |
      <a href="https://bk.tencent.com/index/" target="_blank">{{ website }}</a>
    </div>
    <p>Copyright © 2012 Tencent BlueKing. All Rights Reserved. {{ version }}</p>
  </div>
</template>
<script lang="ts">
import { useI18n } from 'vue-i18n';
import { defineComponent } from 'vue';
import { storeToRefs } from 'pinia';
import useGlobalStore from '../store/global';
const { bscpVersion } = storeToRefs(useGlobalStore());

export default defineComponent({
  name: 'AppFooter',
  setup() {
    const { t } = useI18n();
    return {
      version: bscpVersion,
      support: t('技术支持'),
      forum: t('社区论坛'),
      website: t('产品官网'),
    };
  },
});
</script>
<style lang="scss" scoped>
.app-footer {
  font-size: 12px;
  color: #b7c0ca;
  width: 100%;
  text-align: center;
  line-height: 20px;
  padding-top: 25px;
  .links a {
    color: #3a84ff;
  }
}
</style>
