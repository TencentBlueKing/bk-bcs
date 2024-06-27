<template>
  <bk-select
    v-model="localApp.id"
    ref="selectorRef"
    class="service-selector"
    :popover-options="{ theme: 'light bk-select-popover' }"
    :popover-min-width="360"
    :filterable="true"
    :input-search="false"
    :clearable="false"
    :loading="loading"
    :search-placeholder="$t('请输入关键字')"
    @change="handleAppChange">
    <template #trigger>
      <div class="selector-trigger">
        <bk-overflow-title v-if="localApp?.name" class="app-name" type="tips">
          {{ localApp.name }}
        </bk-overflow-title>
        <span v-else class="no-app">{{ $t('暂无服务') }}</span>
        <AngleUpFill class="arrow-icon arrow-fill" />
      </div>
    </template>
    <bk-option v-for="item in serviceList" :key="item.id" :value="item.id" :label="item.spec.name">
      <div
        v-cursor="{
          active: !item.permissions.view,
        }"
        :class="['service-option-item', { 'no-perm': !item.permissions.view }]">
        <div class="name-text">{{ item.spec.name }}</div>
        <div class="type-tag" :class="{ 'type-tag--en': locale === 'en' }">
          {{ item.spec.config_type === 'file' ? $t('文件型') : $t('键值型') }}
        </div>
      </div>
    </bk-option>
  </bk-select>
</template>

<script lang="ts" setup>
  import { ref, onMounted } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import { IAppItem } from '../../../../../../types/app';
  import { getAppList } from '../../../../../api';
  import { AngleUpFill } from 'bkui-vue/lib/icon';
  import { useI18n } from 'vue-i18n';

  const emits = defineEmits(['change-service']);

  const { locale } = useI18n();
  const route = useRoute();
  const router = useRouter();

  const loading = ref(false);
  const localApp = ref({
    name: '',
    id: Number(route.params.appId),
    serviceType: '',
  });
  const bizId = ref(String(route.params.spaceId));
  const serviceList = ref<IAppItem[]>([]);

  onMounted(async () => {
    await loadServiceList();
    const service = serviceList.value.find((service) => service.id === Number(route.params.appId));
    if (service) {
      localApp.value = {
        name: service.spec.name,
        id: service.id!,
        serviceType: service.spec.config_type!,
      };
      emits('change-service', localApp.value.serviceType, localApp.value.name);
    } else if (serviceList.value.length) {
      handleAppChange(serviceList.value[0].id!);
    }
  });

  // 载入服务列表
  const loadServiceList = async () => {
    loading.value = true;
    try {
      const query = {
        start: 0,
        all: true,
      };
      const resp = await getAppList(bizId.value, query);
      serviceList.value = resp.details;
    } catch (e) {
      console.error(e);
    } finally {
      loading.value = false;
    }
  };
  // 下拉列表操作
  const handleAppChange = async (appId: number) => {
    const service = serviceList.value.find((service) => service.id === appId);
    if (service) {
      localApp.value = {
        name: service.spec.name,
        id: service.id!,
        serviceType: service.spec.config_type!,
      };
    }
    setLastAccessedServiceDetail(appId);
    await router.push({ name: route.name!, params: { spaceId: bizId.value, appId } });
    emits('change-service', localApp.value.serviceType, localApp.value.name);
  };
  const setLastAccessedServiceDetail = (appId: number) => {
    localStorage.setItem('lastAccessedServiceDetail', JSON.stringify({ spaceId: bizId.value, appId }));
  };
</script>

<style scoped lang="scss">
  .service-selector {
    &.popover-show {
      .selector-trigger .arrow-icon {
        transform: rotate(-180deg);
      }
    }
    &.is-focus {
      .selector-trigger {
        outline: 0;
      }
    }
    .selector-trigger {
      padding: 0 10px 0;
      height: 32px;
      cursor: pointer;
      display: flex;
      align-items: center;
      justify-content: space-between;
      border-radius: 2px;
      transition: all 0.3s;
      background: #f0f1f5;
      font-size: 14px;
      .app-name {
        max-width: 220px;
        color: #313238;
      }
      .no-app {
        font-size: 16px;
        color: #c4c6cc;
      }
      .arrow-icon {
        margin-left: 13.5px;
        color: #979ba5;
        transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
      }
    }
  }
  .service-option-item {
    display: flex;
    justify-content: flex-start;
    align-items: center;
    width: 100%;
    .name-text {
      margin-right: 5px;
      flex: 1;
      white-space: nowrap;
      text-overflow: ellipsis;
      overflow: hidden;
    }
    .type-tag {
      flex-shrink: 0;
      width: 52px;
      height: 22px;
      line-height: 22px;
      color: #63656e;
      font-size: 12px;
      text-align: center;
      background: #f0f1f5;
      border-radius: 2px;
      &--en {
        width: 96px;
      }
    }
  }
</style>
