<template>
  <Card :title="`按 ${label} 统计`" :height="368">
    <template #head-suffix>
      <bk-tag theme="info" type="stroke" style="margin-left: 8px"> 标签 </bk-tag>
      <TriggerBtn v-model:currentType="currentType" style="margin-left: 8px" />
    </template>
    <bk-loading class="loading-wrap" :loading="loading">
      <component v-if="selectedLabelData?.length" :is="currentComponent" :data="selectedLabelData" :label="label" />
      <bk-exception
        v-else
        class="exception-wrap-item exception-part"
        type="empty"
        scene="part"
        description="没有数据" />
    </bk-loading>
  </Card>
</template>

<script lang="ts" setup>
  import { ref, onMounted, computed, watch } from 'vue';
  import Card from '../../../components/card.vue';
  import TriggerBtn from '../../../components/trigger-btn.vue';
  import Pie from './pie.vue';
  import Column from './column.vue';
  import Table from './table.vue';
  import { IClientLabelItem, IClinetCommonQuery } from '../../../../../../../../types/client';
  import { getClientLabelData } from '../../../../../../../api/client';
  import useClientStore from '../../../../../../../store/client';
  import { storeToRefs } from 'pinia';

  const clientStore = useClientStore();

  const { searchQuery } = storeToRefs(clientStore);

  const props = defineProps<{
    bkBizId: string;
    appId: number;
    selectedLabel: string;
  }>();

  const currentType = ref('column');
  const componentMap = {
    pie: Pie,
    column: Column,
    table: Table,
  };
  const allLabeldata = ref<{ [key: string]: IClientLabelItem[] }>();
  const loading = ref(false);

  const currentComponent = computed(() => componentMap[currentType.value as keyof typeof componentMap]);
  const selectedLabelData = computed(() => allLabeldata.value?.[props.selectedLabel]);
  const label = computed(() => props.selectedLabel.charAt(0).toUpperCase() + props.selectedLabel.slice(1));

  onMounted(() => {
    loadChartData();
  });

  watch(
    () => props.appId,
    () => {
      loadChartData();
    },
  );

  watch(
    () => searchQuery.value,
    () => {
      loadChartData();
    },
    { deep: true },
  );

  const loadChartData = async () => {
    const params: IClinetCommonQuery = {
      last_heartbeat_time: searchQuery.value.last_heartbeat_time,
      search: searchQuery.value.search,
    };
    try {
      loading.value = true;
      const res = await getClientLabelData(props.bkBizId, props.appId, params);
      allLabeldata.value = res;
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
