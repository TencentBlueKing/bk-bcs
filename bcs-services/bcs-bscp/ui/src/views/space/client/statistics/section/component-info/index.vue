<template>
  <div>
    <SectionTitle :title="'客户端组件信息统计'" />
    <div class="content-wrap">
      <div class="left">
        <Card title="组件版本分布" :height="416">
          <template #head-suffix>
            <TriggerBtn v-model:currentType="currentType" style="margin-left: 16px" />
          </template>
          <bk-loading class="loading-wrap" :loading="loading">
            <component v-if="data?.length" :is="currentComponent" :data="needData" />
            <bk-exception
              v-else
              class="exception-wrap-item exception-part"
              type="empty"
              scene="part"
              description="没有数据" />
          </bk-loading>
        </Card>
      </div>
      <div class="right">
        <Card v-for="item in resourceData" :key="item.name" :title="item.name" :width="207" :height="128">
          <div class="resource-info">
            <span v-if="item.value">
              <span class="time">{{ item.key.includes('cpu') ? item.value : Math.round(item.value) }}</span>
              <span class="unit">{{ item.key.includes('cpu') ? '核' : 'MB' }}</span>
            </span>
            <span v-else class="empty">暂无数据</span>
          </div>
        </Card>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { ref, onMounted, computed } from 'vue';

  import Card from '../../components/card.vue';
  import TriggerBtn from '../../components/trigger-btn.vue';
  import SectionTitle from '../../components/section-title.vue';
  import Pie from './pie.vue';
  import Column from './column.vue';
  import Table from './table.vue';
  import { getClientComponentInfoData } from '../../../../../../api/client';
  import {
    IVersionDistributionItem,
    IVersionDistributionPie,
    IVersionDistributionPieItem,
    IInfoCard,
  } from '../../../../../../../types/client';

  const props = defineProps<{
    bkBizId: string;
    appId: number;
  }>();

  const resourceData = ref<IInfoCard[]>([
    {
      value: 0,
      name: '平均 CPU 资源占用',
      key: 'cpu_avg_usage',
    },
    {
      value: 0,
      name: '平均内存资源占用',
      key: 'memory_avg_usage',
    },
    {
      value: 0,
      name: '最大 CPU 资源占用',
      key: 'cpu_max_usage',
    },
    {
      value: 0,
      name: '最大内存资源占用',
      key: 'memory_max_usage',
    },
    {
      value: 0,
      name: '最小 CPU 资源占用',
      key: 'cpu_min_usage',
    },
    {
      value: 0,
      name: '最小内存资源占用',
      key: 'memory_min_usage',
    },
  ]);
  const currentType = ref('pie');
  const componentMap = {
    pie: Pie,
    column: Column,
    table: Table,
  };
  const data = ref<IVersionDistributionItem[]>([]);
  const sunburstData = ref<IVersionDistributionPie>({
    name: '配置版本',
    children: [],
  });
  const loading = ref(false);
  const currentComponent = computed(() => componentMap[currentType.value as keyof typeof componentMap]);
  const needData = computed(() => (currentType.value === 'pie' ? sunburstData.value : data.value));

  onMounted(() => {
    loadChartData();
  });

  const loadChartData = async () => {
    try {
      loading.value = true;
      const res = await getClientComponentInfoData(props.bkBizId, props.appId, {});
      data.value = res.version_distribution;
      sunburstData.value.children = convertToTree(res.version_distribution);
      Object.entries(res.resource_usage).map(
        ([key, value]) => (resourceData.value.find((item) => item.key === key)!.value = value as number),
      );
    } catch (error) {
      console.error(error);
    } finally {
      loading.value = false;
    }
  };

  const convertToTree = (data: IVersionDistributionItem[]) => {
    const tree: IVersionDistributionPieItem[] = [];
    data.forEach((item) => {
      const { client_type, client_version, value, percent } = item;
      let typeNode = tree.find((node) => node.name === client_type);
      if (!typeNode) {
        typeNode = { name: client_type, children: [], percent: 0, value: 0 };
        tree.push(typeNode);
      }
      const versionNode: IVersionDistributionPieItem = {
        name: client_version,
        percent,
        value,
      };
      typeNode.children?.push(versionNode);
      typeNode.percent += percent;
      typeNode!.value += value;
    });
    return tree;
  };
</script>

<style scoped lang="scss">
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
</style>
