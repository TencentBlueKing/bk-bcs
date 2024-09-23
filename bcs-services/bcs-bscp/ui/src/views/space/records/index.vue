<template>
  <section class="record-management-page">
    <div class="operate-area">
      <service-selector />
      <date-picker class="date-picker" @change-time="updateParams" />
      <search-option @send-search-data="updateParams" />
    </div>
    <record-table ref="recordTableRef" :space-id="spaceId" :search-params="searchParams" />
  </section>
</template>
<script setup lang="ts">
  import { ref, onMounted } from 'vue';
  import { useRoute } from 'vue-router';
  import serviceSelector from './components/service-selector.vue';
  import datePicker from './components/date-picker.vue';
  import searchOption from './components/search-option.vue';
  import recordTable from './components/record-table.vue';
  import { IRecordQuery } from '../../../../types/record';

  const route = useRoute();

  const recordTableRef = ref();
  const spaceId = ref(String(route.params.spaceId));
  const searchParams = ref<IRecordQuery>({}); // 外部搜索数据参数汇总
  const dateTimeParams = ref<{ start_time?: string; end_time?: string }>({}); // 日期组件参数
  const optionParams = ref<IRecordQuery>(); // 搜索组件参数
  const init = ref(true);

  onMounted(() => {
    mergeData();
    // 操作记录id
    const id = Number(route.query.id);
    if (id && id > 0) {
      searchParams.value.id = id;
    }
    init.value = false;
  });

  const updateParams = (data: string[] | IRecordQuery) => {
    if (Array.isArray(data)) {
      dateTimeParams.value.start_time = data[0];
      dateTimeParams.value.end_time = data[1];
    } else {
      optionParams.value = data;
    }
    if (!init.value) {
      mergeData();
    }
  };

  const mergeData = () => {
    searchParams.value = {
      ...dateTimeParams.value,
      ...optionParams.value,
      app_id: Number(route.params.appId),
      all: Number(route.params.appId) <= -1,
    };
    if (Number(route.query.id) <= -1) {
      delete searchParams.value.id;
    }
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
