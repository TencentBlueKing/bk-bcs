<template>
  <!-- 加Loading会导致跟index的loading错位的效果 -->
  <div id="app">
    <Navigation>
      <RouterView />
      <template #sideMenu>
        <RouterView name="sideMenu" />
      </template>
    </Navigation>
    <!-- 权限弹窗 -->
    <PermDialog ref="applyPermRef" />
    <!-- 登录弹窗 -->
    <BkPaaSLogin ref="loginRef" :width="$INTERNAL ? 700 : 400" :height="$INTERNAL ? 510 : 400" />
  </div>
</template>
<script lang="ts">
import { defineComponent, onBeforeUnmount, onMounted, ref } from 'vue';

import { bus } from '@/common/bus';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import { useAppData } from '@/composables/use-app';
import $i18n from '@/i18n/i18n-setup';
import PermDialog from '@/views/app/apply-perm.vue';
import BkPaaSLogin from '@/views/app/login.vue';
import Navigation from '@/views/app/navigation.vue';

export default defineComponent({
  name: 'App',
  components: { Navigation, BkPaaSLogin, PermDialog },
  setup() {
    const { getUserInfo } = useAppData();
    const isLoading = ref(false);
    const applyPermRef = ref<any>(null);
    const loginRef = ref<any>(null);

    // 权限弹窗
    bus.$on('show-apply-perm-modal', (data) => {
      if (!data) return;
      applyPermRef.value?.show(data);
    });
    // 关闭登录弹窗
    window.addEventListener('message', (event) => {
      if (event.data === 'closeLoginModal') {
        window.location.reload();
      }
    });

    // 校验域名是否正确
    const validateAllowDomains = () => {
      const allowDomains = (window.PREFERRED_DOMAINS || '').split(',');
      const item = allowDomains.find(item => item.trim() === location.hostname);
      if (!item && allowDomains[0]) {
        window.location.href = `//${allowDomains[0]}${location.pathname}`;
      }
    };

    // 校验资源是否更新
    const resourceHash = ref('');
    const flag = ref(false);
    const validateResourceVersion = () => {
      setTimeout(async () => {
        try {
          const res = await fetch(`${window.BK_STATIC_URL}/static/static_version.txt`, { cache: 'no-store' });
          const hash = await res.text();
          if (resourceHash.value && (resourceHash.value !== hash)) {
            $bkInfo({
              type: 'warning',
              clsName: 'custom-info-confirm',
              title: $i18n.t('bcs.newVersion'),
              defaultInfo: true,
              okText: $i18n.t('generic.button.reload'),
              confirmFn: () => {
                window.location.reload();
              },
            });
            flag.value = true;
          }
          resourceHash.value = hash;
          !flag.value && validateResourceVersion();
        } catch (err) {
          console.log(err);
        }
      }, 15000);
    };

    onBeforeUnmount(() => {
      bus.$off('show-apply-perm-modal');
    });

    onMounted(async () => {
      validateResourceVersion();
      validateAllowDomains();

      window.$loginModal = loginRef.value;

      isLoading.value = true;
      await getUserInfo();
      isLoading.value = false;
      document.title = $i18n.t('bcs.title');
    });

    return {
      isLoading,
      applyPermRef,
      loginRef,
    };
  },
});
</script>
<style lang="postcss">
@import '@/css/reset.css';
@import '@/css/app.css';
@import '@/fonts/font-icon/style.css';
@import '@/css/main.css';
</style>
