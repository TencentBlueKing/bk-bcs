<template>
  <div>
    <SectionTitle :title="t('客户端组件信息统计')" />
    <div class="content-wrap">
      <div class="left">
        <Teleport :disabled="!isOpenFullScreen" to="body">
          <div
            ref="containerRef"
            :class="{ fullscreen: isOpenFullScreen }"
            @mouseenter="isShowOperationBtn = true"
            @mouseleave="isShowOperationBtn = false">
            <Card :title="t('组件类型 / 版本分布')" :height="416">
              <template #operation>
                <OperationBtn
                  v-show="isShowOperationBtn"
                  :is-open-full-screen="isOpenFullScreen"
                  @refresh="loadChartData"
                  @toggle-full-screen="isOpenFullScreen = !isOpenFullScreen" />
              </template>
              <template #head-suffix>
                <div class="head-suffix">
                  <div v-if="currentType === 'column'" class="icon-wrap">
                    <span
                      class="action-icon bk-bscp-icon icon-download"
                      v-bk-tooltips="{ content: $t('可下钻图表') }" />
                  </div>
                  <TriggerBtn v-model:currentType="currentType" />
                </div>
              </template>
              <bk-loading class="loading-wrap" :loading="loading">
                <component v-if="data?.length" :is="currentComponent" :data="needDataType" />
                <bk-exception
                  v-else
                  class="exception-wrap-item exception-part"
                  type="empty"
                  scene="part"
                  :description="t('暂无数据')">
                  <template #type>
                    <span class="bk-bscp-icon icon-bar-chart exception-icon" />
                  </template>
                </bk-exception>
              </bk-loading>
            </Card>
          </div>
        </Teleport>
      </div>
      <div class="right">
        <Card v-for="item in resourceData" :key="item.name" :title="item.name" :width="207" :height="128">
          <div class="resource-info">
            <span v-if="item.value">
              <span class="time">{{ item.key.includes('cpu') ? item.value : Math.round(item.value) }}</span>
              <span class="unit">{{ item.unit }}</span>
            </span>
            <span v-else class="empty">{{ t('暂无数据') }}</span>
          </div>
        </Card>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { ref, onMounted, computed, watch } from 'vue';
  import Card from '../../components/card.vue';
  import TriggerBtn from '../../components/trigger-btn.vue';
  import SectionTitle from '../../components/section-title.vue';
  import OperationBtn from '../../components/operation-btn.vue';
  import Pie from './pie.vue';
  import Column from './column/index.vue';
  import Table from './table.vue';
  import { getClientComponentInfoData } from '../../../../../../api/client';
  import {
    IVersionDistributionItem,
    IVersionDistributionPie,
    IVersionDistributionPieItem,
    IInfoCard,
    IClinetCommonQuery,
  } from '../../../../../../../types/client';
  import useClientStore from '../../../../../../store/client';
  import { storeToRefs } from 'pinia';
  import { useI18n } from 'vue-i18n';

  const { t } = useI18n();

  const clientStore = useClientStore();
  const { searchQuery } = storeToRefs(clientStore);

  const props = defineProps<{
    bkBizId: string;
    appId: number;
  }>();

  const resourceData = ref<IInfoCard[]>([
    {
      value: 0,
      name: t('平均 CPU 资源占用'),
      key: 'cpu_avg_usage',
    },
    {
      value: 0,
      name: t('平均内存资源占用'),
      key: 'memory_avg_usage',
    },
    {
      value: 0,
      name: t('最大 CPU 资源占用'),
      key: 'cpu_max_usage',
    },
    {
      value: 0,
      name: t('最大内存资源占用'),
      key: 'memory_max_usage',
    },
    {
      value: 0,
      name: t('最小 CPU 资源占用'),
      key: 'cpu_min_usage',
    },
    {
      value: 0,
      name: t('最小内存资源占用'),
      key: 'memory_min_usage',
    },
  ]);
  const currentType = ref('column');
  const componentMap = {
    pie: Pie,
    column: Column,
    table: Table,
  };
  const data = ref<IVersionDistributionItem[]>([]);
  const sunburstData = ref<IVersionDistributionPie>({
    name: t('组件类型分布'),
    children: [],
  });
  const loading = ref(false);
  const isOpenFullScreen = ref(false);
  const containerRef = ref();
  const initialWidth = ref(0);
  const isShowOperationBtn = ref(false);

  const currentComponent = computed(() => componentMap[currentType.value as keyof typeof componentMap]);
  const needDataType = computed(() => (currentType.value === 'table' ? data.value : sunburstData.value));

  watch(
    () => props.appId,
    () => loadChartData(),
  );

  watch(
    () => searchQuery.value,
    () => {
      loadChartData();
    },
    { deep: true },
  );

  watch(
    () => isOpenFullScreen.value,
    (val) => {
      containerRef.value!.style.width = val ? '100%' : `${initialWidth.value}px`;
    },
  );

  onMounted(() => {
    loadChartData();
    initialWidth.value = containerRef.value!.offsetWidth;
  });

  const loadChartData = async () => {
    const params: IClinetCommonQuery = {
      last_heartbeat_time: searchQuery.value.last_heartbeat_time,
      search: searchQuery.value.search,
    };
    try {
      loading.value = true;
      const res = await getClientComponentInfoData(props.bkBizId, props.appId, params);
      data.value = res.version_distribution.map((item: IVersionDistributionItem) => {
        const { client_type } = item;
        let name = '';
        switch (client_type) {
          case 'sidecar':
            name = `SideCar ${t('客户端')}`;
            break;
          case 'sdk':
            name = `SDK ${t('客户端')}`;
            break;
          case 'agent':
            name = t('主机插件客户端');
            break;
          case 'command':
            name = `CLI ${t('客户端')}`;
            break;
        }
        return {
          ...item,
          name,
        };
      });
      sunburstData.value.children = convertToTree(data.value);
      Object.entries(res.resource_usage).forEach(([key, value]) => {
        const item = resourceData.value.find((item) => item.key === key) as IInfoCard;
        item!.value = value as number;
        if (!item.key.includes('cpu')) {
          item.unit = 'MB';
        } else {
          item.unit = t('核');
        }
      });
    } catch (error) {
      console.error(error);
    } finally {
      loading.value = false;
    }
  };

  const convertToTree = (data: IVersionDistributionItem[]) => {
    const tree: IVersionDistributionPieItem[] = [];
    data.forEach((item) => {
      const { client_type, client_version, value, percent, name } = item;
      let typeNode = tree.find((node) => node.name === name);
      if (!typeNode) {
        typeNode = { name: name!, children: [], percent: 0, value: 0, client_type };
        tree.push(typeNode);
      }
      const versionNode: IVersionDistributionPieItem = {
        name: client_version,
        percent,
        value,
        client_type,
      };
      typeNode.children?.push(versionNode);
      typeNode.percent += percent;
      typeNode!.value += value;
    });
    return tree;
  };
</script>

<style scoped lang="scss">
  .fullscreen {
    position: fixed;
    top: 0;
    left: 0;
    width: 100vw;
    height: 100vh;
    z-index: 5000;
    .card {
      width: 100%;
      height: 100vh !important;
      .loading-wrap {
        height: 100%;
      }
      :deep(.operation-btn) {
        top: 0 !important;
      }
    }
  }
  .content-wrap {
    display: flex;
    .left {
      flex: 1;
      margin-right: 16px;
      .loading-wrap {
        height: 100%;
      }
    }
    .right {
      display: flex;
      justify-content: space-between;
      align-content: space-between;
      flex-wrap: wrap;
      width: 430px;
      .resource-info {
        margin-left: 8px;
        color: #63656e;
        .time {
          font-size: 32px;
          color: #63656e;
          font-weight: 700;
        }
        .unit {
          font-size: 16px;
          margin-left: 2px;
        }
        .empty {
          font-size: 12px;
          color: #979ba5;
        }
      }
    }
  }
  :deep(.bk-exception-part) {
    height: 100%;
    justify-content: center;
    transform: translateY(-20px);
  }

  .head-suffix {
    margin-left: 16px;
    display: flex;
    align-items: center;
    gap: 8px;
    .icon-wrap {
      font-size: 12px;
      width: 18px;
      height: 18px;
      background: #f0f3ff;
      border-radius: 2px;
      text-align: center;
      line-height: 18px;
      color: #7594ef;
    }
  }
</style>
