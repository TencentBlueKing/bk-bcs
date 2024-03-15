<template>
  <div class="head">
    <div class="head-left">
      <span class="title">{{ title }}</span>
      <bk-select
        v-model="localApp.id"
        ref="selectorRef"
        class="service-selector"
        :popover-options="{ theme: 'light bk-select-popover service-selector-popover' }"
        :popover-min-width="360"
        :filterable="true"
        :input-search="false"
        :clearable="false"
        :loading="loading"
        :search-placeholder="$t('请输入关键字')"
        @change="handleAppChange">
        <template #trigger>
          <div class="selector-trigger">
            <span class="app-name">{{ localApp?.name }}</span>
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
          </div>
        </bk-option>
      </bk-select>
    </div>
    <div class="head-right">
      <div class="selector-tips">最后心跳时间</div>
      <bk-select
        v-model="heartbeatTime"
        class="heartbeat-selector"
        :clearable="false"
        @change="handleHeartbeatTimeChange">
        <bk-option v-for="item in heartbeatTimeList" :id="item.value" :key="item.value" :name="item.label" />
      </bk-select>
      <SearchSelector :bk-biz-id="bizId" :app-id="localApp.id" />
      <bk-button theme="primary" style="margin-left: 8px" @click="emits('search')">
        <Search class="search-icon" />查询
      </bk-button>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { ref, onMounted } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import { storeToRefs } from 'pinia';
  import { AngleUpFill, Search } from 'bkui-vue/lib/icon';
  import { getAppList } from '../../../../api';
  import { CLIENT_HEARTBEAT_LIST } from '../../../../constants/client';
  import { IAppItem } from '../../../../../types/app';
  import useClientStore from '../../../../store/client';
  import SearchSelector from './search-selector.vue';

  defineProps<{
    title: string;
  }>();

  const emits = defineEmits(['search']);

  const clientStore = useClientStore();
  const { searchQuery } = storeToRefs(clientStore);
  const route = useRoute();
  const router = useRouter();

  const loading = ref(false);
  const localApp = ref({
    name: '',
    id: 0,
  });
  const serviceList = ref<IAppItem[]>([]);
  const heartbeatTime = ref(searchQuery.value.last_heartbeat_time);
  const heartbeatTimeList = ref(CLIENT_HEARTBEAT_LIST);
  const selectorRef = ref();

  const bizId = ref(String(route.params.spaceId));

  onMounted(async () => {
    await loadServiceList();
    const service = serviceList.value.find((service) => service.id === Number(route.params.appId));
    if (service) {
      localApp.value = {
        name: service.spec.name,
        id: service.id!,
      };
      emits('search');
    } else {
      handleAppChange(serviceList.value[0].id!);
    }
  });

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

  const handleAppChange = (appId: number) => {
    const service = serviceList.value.find((service) => service.id === appId);
    if (service) {
      localApp.value = {
        name: service.spec.name,
        id: service.id!,
      };
    }
    setLastSelectedClientService(appId);
    router.push({ name: route.name!, params: { spaceId: bizId.value, appId } });
    emits('search');
  };

  const handleHeartbeatTimeChange = (value: number) => {
    clientStore.$patch((state) => {
      state.searchQuery.last_heartbeat_time = value;
    });
    emits('search');
  };

  const setLastSelectedClientService = (appId: number) => {
    localStorage.setItem('lastSelectedClientService', JSON.stringify({ spaceId: bizId.value, appId }));
  };
</script>

<style scoped lang="scss">
  .head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    font-size: 20px;
    line-height: 28px;
    height: 32px;
    .head-left {
      display: flex;
      align-items: center;
      .title {
        position: relative;
        color: #313238;
        font-weight: 700;
        &::after {
          position: absolute;
          right: -16px;
          content: '';
          width: 1px;
          height: 24px;
          background: #dcdee5;
        }
      }
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
          margin-left: 33px;
          cursor: pointer;
          .app-name {
            color: #63656e;
          }
          .arrow-icon {
            margin-left: 13.5px;
            font-size: 14px;
            color: #979ba5;
            transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
          }
        }
      }
    }
    .head-right {
      display: flex;
      align-items: center;
      font-size: 12px;
      .selector-tips {
        width: 88px;
        height: 32px;
        background: #fafbfd;
        border: 1px solid #c4c6cc;
        border-radius: 2px 0 0 2px;
        line-height: 32px;
        text-align: center;
        border-right: none;
        color: #63656e;
      }
      .heartbeat-selector {
        width: 112px;
        margin-right: 8px;
      }
      .search-icon {
        margin-right: 8px;
      }
    }
  }
</style>
