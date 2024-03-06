<template>
  <div class="head">
    <div class="head-left">
      <span class="title">客户端统计</span>
      <bk-select
        v-model="localApp"
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
            <span class="app-name">{{ appData?.spec.name }}</span>
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
      <bk-select v-model="heartbeatTime" class="heartbeat-selector" :clearable="false">
        <bk-option v-for="item in heartbeatTimeList" :id="item.value" :key="item.value" :name="item.label" />
      </bk-select>
      <bk-input
        v-model="searchStr"
        class="search-client-input"
        :placeholder="'UID/IP/标签/当前配置版本/目标配置版本/最近一次拉取配置状态/附加信息/在线状态/客户端组件版本'"
        :clearable="true">
        <template #suffix>
          <Search class="search-input-icon" />
        </template>
      </bk-input>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { ref, onMounted } from 'vue';
  import { useRoute } from 'vue-router';
  import { storeToRefs } from 'pinia';
  import { AngleUpFill, Search } from 'bkui-vue/lib/icon';
  import { getAppList } from '../../../../api';
  import { IAppItem } from '../../../../../types/app';
  import useClientStore from '../../../../store/client';

  const clientStore = useClientStore();
  const { appData } = storeToRefs(clientStore);
  const route = useRoute();

  const loading = ref(false);
  const localApp = ref(appData.value.id);
  const serviceList = ref<IAppItem[]>([]);
  const heartbeatTime = ref('1');
  const heartbeatTimeList = ref([
    {
      value: '1',
      label: '近1分钟',
    },
    {
      value: '2',
      label: '近2分钟',
    },
    {
      value: '3',
      label: '近3分钟',
    },
  ]);
  const searchStr = ref('');
  const selectorRef = ref();

  const bizId = route.params.spaceId as string;

  onMounted(() => {
    loadServiceList();
  });

  const loadServiceList = async () => {
    loading.value = true;
    try {
      const query = {
        start: 0,
        all: true,
      };
      const resp = await getAppList(bizId, query);
      serviceList.value = resp.details;
    } catch (e) {
      console.error(e);
    } finally {
      loading.value = false;
    }
  };

  const handleAppChange = (id: number) => {
    const service = serviceList.value.find((service) => service.id === id);
    appData.value.id = service!.id as number;
    appData.value.spec = service!.spec;
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
    margin-bottom: 24px;
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
      .search-client-input {
        width: 600px;
      }
      .search-input-icon {
        padding-right: 10px;
        color: #979ba5;
        background: #ffffff;
      }
    }
  }
</style>
