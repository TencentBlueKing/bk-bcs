<template>
  <div v-if="isDrillDown" class="nav">
    <span class="group-dimension" @click="handleCancelDrillDown">{{ $t('组件类型分布') }}</span> /
    <span class="drill-down-data">{{ drillDownData }}</span>
  </div>
  <columnChart v-if="!isDrillDown" :data="props.data" @jump="jumpToSearch" @drill-down="handleDrillDown" />
  <pieChart v-else :data="pieData" @jump="jumpToSearch" />
</template>

<script lang="ts" setup>
  import { ref, watch } from 'vue';
  import { useRouter, useRoute } from 'vue-router';
  import pieChart from './pie.vue';
  import columnChart from './column.vue';

  const props = defineProps<{
    data: any;
  }>();

  const columnData = ref(props.data.children || []);
  const pieData = ref([]);
  const router = useRouter();
  const route = useRoute();
  const isDrillDown = ref(false);
  const drillDownData = ref('');

  const bizId = ref(String(route.params.spaceId));
  const appId = ref(Number(route.params.appId));

  watch(
    () => props.data,
    () => {
      columnData.value = props.data.children;
      isDrillDown.value = false;
      drillDownData.value = '';
    },
    { deep: true },
  );

  const handleCancelDrillDown = () => {
    isDrillDown.value = false;
    drillDownData.value = '';
  };

  const jumpToSearch = (query: { [key: string]: string }) => {
    const routeData = router.resolve({
      name: 'client-search',
      params: { appId: appId.value, bizId: bizId.value },
      query,
    });
    window.open(routeData.href, '_blank');
  };

  const handleDrillDown = (data: any) => {
    console.log(data);
    isDrillDown.value = true;
    drillDownData.value = data.name;
    pieData.value = data.children;
  };
</script>

<style scoped lang="scss">
  :deep(.g2-tooltip) {
    visibility: hidden;
    .g2-tooltip-list-item {
      .g2-tooltip-marker {
        border-radius: initial !important;
      }
    }
  }
  .nav {
    font-size: 12px;
    color: #313238;
    .group-dimension {
      margin-right: 8px;
      cursor: pointer;
      &:hover {
        color: #3a84ff;
      }
    }
    .drill-down-data {
      color: #979ba5;
      margin-left: 8px;
    }
  }
</style>
