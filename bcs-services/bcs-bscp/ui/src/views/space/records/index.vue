<template>
  <section class="record-management-page">
    <div class="operate-area">
      <service-selector @select-service="selectService" />
      <date-picker class="date-picker" @change-time="splitTime" />
      <search-option />
    </div>
    <record-table
      ref="recordTableRef"
      :space-id="spaceId"
      :search-params="searchParams"
      @update-search-params="searchParams = $event" />
  </section>
</template>
<script setup lang="ts">
  import { ref } from 'vue';
  import { useRoute } from 'vue-router';
  import serviceSelector from './components/service-selector.vue';
  import datePicker from './components/date-picker.vue';
  import searchOption from './components/search-option.vue';
  import recordTable from './components/record-table.vue';
  import { IRecordQuery } from '../../../../types/record';

  const route = useRoute();

  const recordTableRef = ref(null);
  const spaceId = ref(String(route.params.spaceId));
  const searchParams = ref<IRecordQuery>({});

  const selectService = (serviceId: number) => {
    searchParams.value.all = !(serviceId > -1); // 不正确的服务id，则全部搜索
    if (serviceId > -1) {
      searchParams.value.app_id = serviceId;
    }
  };
  const splitTime = (time: string[]) => {
    searchParams.value.start_time = time[0];
    searchParams.value.end_time = time[1];
  };
</script>
<style lang="scss" scoped>
  .record-management-page {
    height: calc(100% - 33px);
    padding: 24px;
    background: #f5f7fa;
    overflow: hidden;
    .date-picker {
      margin-left: 8px;
    }
  }
  .operate-area {
    display: flex;
    align-items: center;
    justify-content: flex-start;
    margin-bottom: 16px;
  }
</style>
