<template>
  <div id="app">
    <Navigation>
      <RouterView v-if="!!projectList.length" />
      <ProjectGuide v-else-if="!isLoading" />
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
import { defineComponent, onBeforeUnmount, onMounted, ref } from '@vue/composition-api';
import Navigation from '@/views/app/navigation.vue';
import BkPaaSLogin from '@/views/app/login.vue';
import PermDialog from '@/views/app/apply-perm.vue';
import ProjectGuide from '@/views/app/empty-project-guide.vue';
import $store from '@/store';
import $i18n from '@/i18n/i18n-setup';
import { bus } from '@/common/bus';
import useProject from '@/views/app/use-project';

export default defineComponent({
  name: 'App',
  components: { Navigation, BkPaaSLogin, PermDialog, ProjectGuide },
  setup() {
    const { projectList, getProjectList } = useProject();
    const isLoading = ref(true);
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

    onBeforeUnmount(() => {
      bus.$off('show-apply-perm-modal');
    });

    onMounted(async () => {
      validateAllowDomains();

      window.$loginModal = loginRef.value.login;

      isLoading.value = true;
      await Promise.all([
        $store.dispatch('userInfo'),
        getProjectList(),
      ]);
      isLoading.value = false;
      document.title = $i18n.t('容器管理平台 | 腾讯蓝鲸智云');
    });

    return {
      isLoading,
      applyPermRef,
      loginRef,
      projectList,
    };
  },
});
</script>
<style lang="postcss">
@import '@/css/reset.css';
@import '@/css/app.css';
@import '@/fonts/style.css';
@import '@/css/main.css';
</style>
