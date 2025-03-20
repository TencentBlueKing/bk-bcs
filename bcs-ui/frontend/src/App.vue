<template>
  <!-- 加Loading会导致跟index的loading错位的效果 -->
  <div id="app">
    <NoticeComponent :api-url="apiUrl" ref="noticeRef" id="bcs-notice-com" @show-alert-change="init" />
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
    <!-- 资源更新提示 -->
    <bcs-dialog
      v-model="showResourceUpdateDialog"
      theme="warning"
      :mask-close="false"
      header-position="left"
      footer-position="center"
      :title="$t('bcs.newVersion.title')"
      width="480"
      @confirm="reloadPage">
      <p class="text-[14px] leading-[21px] !text-left">{{ $t('bcs.newVersion.p1') }}</p>
      <p class="text-[14px] leading-[21px] !text-left mt-[10px]">{{ $t('bcs.newVersion.p2') }}</p>
    </bcs-dialog>
    <!-- AI小鲸 -->
    <AiAssistant ref="AiAssistantRef" />
  </div>
</template>
<script lang="ts" setup>
import { onBeforeUnmount, onMounted, provide, ref } from 'vue';

import NoticeComponent from '@blueking/notice-component-vue2';

import '@blueking/notice-component-vue2/dist/style.css';
import { bus } from '@/common/bus';
import { BCS_UI_PREFIX } from '@/common/constant';
import AiAssistant from '@/components/assistant/ai-assistant.vue';
import { Preset } from '@/components/assistant/use-assistant-store';
import { AiSendMsgFnInjectKey, useAppData } from '@/composables/use-app';
import useCalcHeight from '@/composables/use-calc-height';
import usePlatform from '@/composables/use-platform';
import PermDialog from '@/views/app/apply-perm.vue';
import BkPaaSLogin from '@/views/app/login.vue';
import Navigation from '@/views/app/navigation.vue';


const { getUserInfo } = useAppData();
const { config, getPlatformInfo, setDocumentTitle, setShortcutIcon } = usePlatform();
const isLoading = ref(false);
const applyPermRef = ref<any>(null);
const loginRef = ref<any>(null);

// 通知
const apiUrl = ref(`${BCS_UI_PREFIX}/announcements`);
// 设置内容高度
const noticeRef = ref();
const { init } = useCalcHeight([
  {
    prop: 'height',
    el: '.bk-navigation',
    calc: noticeRef,
  },
  {
    el: '.container-content',
    calc: [noticeRef, '.bk-navigation-header'],
  },
  {
    prop: 'height',
    el: '.nav-slider-list',
    calc: [noticeRef, '.bk-navigation-header', '.nav-slider-footer'],
  },
  {
    prop: 'height',
    el: '.v-m-menu-box',
    calc: [noticeRef, '.bk-navigation-header', '.v-m-view-selector'],
  },
]);

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
const showResourceUpdateDialog = ref(false);
const validateResourceVersion = () => {
  setTimeout(async () => {
    try {
      const res = await fetch(`${window.BK_STATIC_URL}/static/static_version.txt`, { cache: 'no-store' });
      const hash = await res?.text();
      if (resourceHash.value && (resourceHash.value !== hash)) {
        showResourceUpdateDialog.value = true;
        flag.value = true;
      }
      resourceHash.value = hash;
      !flag.value && validateResourceVersion();
    } catch (err) {
      console.log(err);
    }
  }, 15000);
};
const reloadPage = () => {
  window.location.reload();
};

// ai 小鲸
const AiAssistantRef = ref<InstanceType<typeof AiAssistant>>();
function handleAI(message: string, pre?: Preset) {
  AiAssistantRef.value?.handleSendMsg(message, pre);
}
provide(AiSendMsgFnInjectKey, handleAI);

// observer noticeRef
let observer: MutationObserver;
const observerNoticeEl = () => {
  if (!noticeRef.value?.$el) return;
  observer = new MutationObserver(init);

  observer.observe(noticeRef.value?.$el, {
    childList: true,
    attributes: true,
  });
};

onMounted(async () => {
  validateResourceVersion();
  validateAllowDomains();

  await getPlatformInfo();
  window.$loginModal = loginRef.value;
  setTimeout(() => {
    observerNoticeEl();
  });
  isLoading.value = true;
  await getUserInfo();
  isLoading.value = false;
  setDocumentTitle(config.i18n);
  setShortcutIcon(config.favicon);
});

onBeforeUnmount(() => {
  bus.$off('show-apply-perm-modal');
  observer?.takeRecords();
  observer?.disconnect();
});
</script>
<style lang="postcss">
@import '@/css/reset.css';
@import '@/css/app.css';
@import '@/fonts/font-icon/style.css';
@import '@/css/main.css';
</style>
