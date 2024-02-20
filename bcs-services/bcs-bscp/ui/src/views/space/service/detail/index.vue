<template>
  <div class="service-detail-page">
    <div :class="['page-detail-content', { 'version-detail-view': versionDetailView }]">
      <div class="version-list-area">
        <VersionListAside :version-detail-view="versionDetailView" :bk-biz-id="bkBizId" :app-id="appId" />
        <div :class="['view-change-trigger', { extend: versionDetailView }]" @click="handleToggleView">
          <AngleDoubleRight class="arrow-icon" />
          <span :class="['text', { 'en-text': locale === 'en-US' }]">{{ t('版本详情') }} </span>
        </div>
      </div>
      <div class="config-setting-area">
        <detail-header
          :class="{ 'version-detail-header': versionDetailView }"
          :bk-biz-id="bkBizId"
          :app-id="appId"
          :version-detail-view="versionDetailView" />
        <div class="setting-content-container">
          <router-view v-if="!appDataLoading" :bk-biz-id="bkBizId" :app-id="appId"></router-view>
        </div>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
  import { onBeforeUnmount, onMounted, ref, watch } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import { useI18n } from 'vue-i18n';
  import { storeToRefs } from 'pinia';
  import { AngleDoubleRight } from 'bkui-vue/lib/icon';
  import BkMessage from 'bkui-vue/lib/message';
  import useServiceStore from '../../../../store/service';
  import useConfigStore from '../../../../store/config';
  import { GET_UNNAMED_VERSION_DATA } from '../../../../constants/config';
  import { permissionCheck, getAppDetail } from '../../../../api';
  import DetailHeader from './components/detail-header.vue';
  import VersionListAside from './config/version-list-aside/index.vue';
  import { AxiosError } from 'axios';

  const route = useRoute();
  const router = useRouter();
  const { t, locale } = useI18n();
  const serviceStore = useServiceStore();
  const configStore = useConfigStore();

  const { permCheckLoading, hasEditServicePerm } = storeToRefs(serviceStore);
  const { versionData, versionDetailView } = storeToRefs(configStore);

  const bkBizId = ref(String(route.params.spaceId));
  const appId = ref(Number(route.params.appId));
  const appDataLoading = ref(true);

  watch(
    () => route.params.appId,
    (val) => {
      if (val) {
        appId.value = Number(val);
        bkBizId.value = String(route.params.spaceId);
        versionData.value = GET_UNNAMED_VERSION_DATA();
        getPermData();
        getAppData();
        setLastAccessedServiceDetail();
      }
    },
  );

  onMounted(() => {
    getPermData();
    getAppData();
    setLastAccessedServiceDetail();
  });

  onBeforeUnmount(() => {
    serviceStore.$reset();
    configStore.$reset();
  });

  const getPermData = async () => {
    permCheckLoading.value = true;
    const res = await permissionCheck({
      resources: [
        {
          biz_id: bkBizId.value,
          basic: {
            type: 'app',
            action: 'update',
            resource_id: appId.value,
          },
        },
      ],
    });
    hasEditServicePerm.value = res.is_allowed;
    permCheckLoading.value = false;
  };

  // 加载服务详情数据
  const getAppData = async () => {
    appDataLoading.value = true;
    try {
      const res = await getAppDetail(bkBizId.value, appId.value);
      serviceStore.$patch((state) => {
        state.appData = res;
      });
      appDataLoading.value = false;
    } catch (e) {
      if ((e as AxiosError).response?.status === 404) {
        router.push({ name: 'service-all' });
        BkMessage({
          theme: 'error',
          message: '访问的服务不存在，请选择其它服务',
        });
      }
    }
  };

  const setLastAccessedServiceDetail = () => {
    localStorage.setItem('lastAccessedServiceDetail', JSON.stringify({ spaceId: bkBizId.value, appId: appId.value }));
  };

  // 切换视图
  const handleToggleView = () => {
    // 非配置管理tab切换为版本详情视图
    if (!versionDetailView.value && route.name !== 'service-config') {
      router.push({ name: 'service-config', params: { spaceId: bkBizId.value, appId: appId.value } });
    }

    // 版本详情视图下，选中版本为废弃版本时，切换到配置详情视图，需要默认选中未命名版本
    if (versionDetailView.value && versionData.value.spec.deprecated) {
      router.push({ name: 'service-config', params: { spaceId: bkBizId.value, appId: appId.value, versionId: 0 } });
    }
    versionDetailView.value = !versionDetailView.value;
  };
</script>
<style lang="scss" scoped>
  .service-detail-page {
    height: 100%;
  }
  .page-detail-content {
    display: flex;
    align-items: top;
    height: 100%;
    &.version-detail-view {
      .version-list-area {
        width: calc(100% - 366px);
      }
      .config-setting-area {
        width: 366px;
      }
    }
  }
  .version-list-area {
    position: relative;
    width: 280px;
    height: 100%;
    z-index: 10;
    border-right: 1px solid #dcdee5;
    // transition: width 0.1s ease-in-out;
  }
  .config-setting-area {
    height: 100%;
    width: calc(100% - 280px);
    .service-detail-header {
      &.version-detail-header {
        background: #fafbfd;
        border: none;
      }
    }
    .setting-content-container {
      height: calc(100% - 41px);
    }
  }
  .view-change-trigger {
    position: absolute;
    top: 37%;
    right: -16px;
    padding-top: 8px;
    width: 16px;
    color: #ffffff;
    background: #c4c6cc;
    border-radius: 0 4px 4px 0;
    text-align: center;
    height: auto;
    cursor: pointer;
    &:hover {
      background: #a3c5fd;
    }
    &.extend {
      .arrow-icon {
        transform: rotate(180deg);
      }
    }
    .text {
      display: inline-block;
      margin-top: -8px;
      font-size: 12px;
      transform: scale(0.833);
    }
    .en-text {
      writing-mode: vertical-lr;
    }
    .arrow-icon {
      font-size: 14px;
    }
  }
</style>
