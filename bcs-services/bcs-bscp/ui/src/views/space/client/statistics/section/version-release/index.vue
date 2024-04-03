<template>
  <div>
    <SectionTitle :title="'配置版本发布'" />
    <Card title="客户端配置版本" :height="344">
      <template #head-suffix>
        <TriggerBtn v-model:currentType="currentType" style="margin-left: 16px" />
      </template>
      <bk-loading class="loading-wrap" :loading="loading">
        <component v-if="data?.length" :is="currentComponent" :data="data" />
        <bk-exception
          v-else
          class="exception-wrap-item exception-part"
          type="empty"
          scene="part"
          description="没有数据" />
      </bk-loading>
    </Card>
  </div>
</template>

<script lang="ts" setup>
  import { ref, onMounted, computed, watch } from 'vue';
  import SectionTitle from '../../components/section-title.vue';
  import Card from '../../components/card.vue';
  import TriggerBtn from '../../components/trigger-btn.vue';
  import Pie from './pie.vue';
  import Column from './column.vue';
  import Table from './table.vue';
  import { IClientConfigVersionItem } from '../../../../../../../types/client';
  import { getConfigVersionData } from '../../../../../../api/client';

  const props = defineProps<{
    bkBizId: string;
    appId: number;
  }>();

  const currentType = ref('pie');
  const componentMap = {
    pie: Pie,
    column: Column,
    table: Table,
  };
  const data = ref<IClientConfigVersionItem[]>();
  const currentComponent = computed(() => componentMap[currentType.value as keyof typeof componentMap]);
  const loading = ref(false);

  onMounted(() => {
    loadChartData();
  });

  watch(
    () => props.appId,
    () => {
      loadChartData();
    },
  );

  const loadChartData = async () => {
    try {
      loading.value = true;
      const res = await getConfigVersionData(props.bkBizId, props.appId, {});
      data.value = res.client_config_version;
    } catch (error) {
      console.error(error);
    } finally {
      loading.value = false;
    }
  };
</script>

<style scoped lang="scss">
  .loading-wrap {
    height: 100%;
  }
</style>
