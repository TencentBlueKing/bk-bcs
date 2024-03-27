<template>
  <bk-sideslider :is-show="show" width="1200" quick-close @closed="emits('close')">
    <template #header>
      <div class="header">
        <span class="title">配置拉取记录</span>
        <span class="uid">UID : {{ uid }}</span>
      </div>
    </template>
    <div class="content-wrap">
      <div class="operate-area">
        <bk-date-picker
          class="date-picker"
          :model-value="initDateTime"
          type="daterange"
          format="yyyy-MM-dd"
          disable-date
          @change="handleDateChange" />
        <SearchInput
          v-model="searchStr"
          :width="600"
          :placeholder="'当前配置版本/目标配置版本/最近一次拉取配置状态'"
          @search="loadTableData" />
      </div>
      <bk-loading :loading="loading">
        <bk-table :data="tableData" :border="['outer', 'row']" :pagination="pagination">
          <bk-table-column label="开始时间" width="154">
            <template #default="{ row }">
              <span v-if="row.spec">
                {{ datetimeFormat(row.spec.start_time) }}
              </span>
            </template>
          </bk-table-column>
          <bk-table-column label="结束时间" width="154">
            <template #default="{ row }">
              <span v-if="row.spec">
                {{ datetimeFormat(row.spec.end_time) }}
              </span>
            </template>
          </bk-table-column>
          <bk-table-column label="当前配置版本" width="133">
            <template #default="{ row }">
              <div
                v-if="row.spec && row.spec.original_release_id"
                class="config-version"
                @click="linkToApp(row.spec.original_release_id)">
                <Share fill="#979BA5" />
                <span class="text">{{ row.original_release_name }}</span>
              </div>
            </template>
          </bk-table-column>
          <bk-table-column label="目标配置版本" width="134">
            <template #default="{ row }">
              <div
                v-if="row.spec && row.spec.target_release_id"
                class="config-version"
                @click="linkToApp(row.spec.target_release_id)">
                <Share fill="#979BA5" />
                <span class="text">{{ row.target_release_name }}</span>
              </div>
            </template>
          </bk-table-column>
          <bk-table-column label="配置拉取方式" prop="attachment.client_mode" width="104"></bk-table-column>
          <bk-table-column label="配置拉取耗时(秒)" width="128">
            <template #default="{ row }">
              <span v-if="row.spec">
                {{ parseInt(row.spec.total_seconds) }}
              </span>
            </template>
          </bk-table-column>
          <bk-table-column label="配置拉取文件数" width="116">
            <template #default="{ row }">
              <div v-if="row.spec">{{ `${row.spec.download_file_num}/${row.spec.total_file_num}` }}</div>
            </template>
          </bk-table-column>
          <bk-table-column label="配置拉取文件大小" width="128">
            <template #default="{ row }">
              <div v-if="row.spec">{{ byteUnitConverse(row.spec.download_file_size) }}</div>
            </template>
          </bk-table-column>
          <bk-table-column label="配置拉取状态" width="104">
            <template #default="{ row }">
              <div v-if="row.spec" class="release_change_status">
                <Spinner v-if="row.spec.release_change_status === 'Skip'" class="spinner-icon" fill="#3A84FF" />
                <div v-else :class="['dot', row.spec.release_change_status]"></div>
                <span>{{ CLIENT_STATUS_MAP[row.spec.release_change_status as keyof typeof CLIENT_STATUS_MAP] }}</span>
                <InfoLine
                  v-if="row.spec.release_change_status === 'Failed'"
                  class="info-icon"
                  fill="#979BA5"
                  v-bk-tooltips="{ content: row.spec.failed_detail_reason }" />
              </div>
            </template>
          </bk-table-column>
          <template #empty>
            <TableEmpty :is-search-empty="isSearchEmpty" @clear="handleClearSearchStr" />
          </template>
        </bk-table>
      </bk-loading>
    </div>
  </bk-sideslider>
</template>

<script lang="ts" setup>
  import { ref, watch } from 'vue';
  import { useRouter } from 'vue-router';
  import { Share, Spinner, InfoLine } from 'bkui-vue/lib/icon';
  import SearchInput from '../../../../../components/search-input.vue';
  import { getClientPullRecord } from '../../../../../api/client';
  import { datetimeFormat, byteUnitConverse } from '../../../../../utils';
  import { CLIENT_STATUS_MAP } from '../../../../../constants/client';
  import dayjs from 'dayjs';
  import TableEmpty from '../../../../../components/table/table-empty.vue';

  const router = useRouter();

  const props = defineProps<{
    bkBizId: string;
    appId: number;
    show: boolean;
    id: number;
    uid: string;
  }>();
  const emits = defineEmits(['close']);

  const isShowSlider = ref(false);
  const initDateTime = ref([dayjs(new Date()).format('YYYY-MM-DD'), dayjs(new Date()).format('YYYY-MM-DD')]);
  const searchStr = ref('');
  const tableData = ref();
  const loading = ref(false);
  const isSearchEmpty = ref(false);

  const pagination = ref({
    count: 0,
    current: 1,
    limit: 10,
  });

  watch(
    () => props.show,
    (val) => {
      if (val) {
        isShowSlider.value = true;
        loadTableData();
      }
    },
  );

  const loadTableData = async () => {
    try {
      loading.value = true;
      isSearchEmpty.value = searchStr.value === '';
      const params = {
        start: pagination.value.limit * (pagination.value.current - 1),
        limit: pagination.value.limit,
        start_time: initDateTime.value[0],
        end_time: initDateTime.value[1],
        search_value: searchStr.value,
      };
      const resp = await getClientPullRecord(props.bkBizId, props.appId, props.id, params);
      pagination.value.count = resp.data.count;
      tableData.value = resp.data.details;
    } catch (error) {
      console.error(error);
    } finally {
      loading.value = false;
    }
  };

  const handleDateChange = (val: string[]) => {
    initDateTime.value = val;
    loadTableData();
  };

  const linkToApp = (versionId: number) => {
    router.push({ name: 'service-config', params: { spaceId: props.bkBizId, appId: props.appId, versionId } });
  };

  const handleClearSearchStr = () => {
    searchStr.value = '';
    loadTableData();
  };
</script>

<style scoped lang="scss">
  .header {
    .title {
      position: relative;
      &::after {
        position: absolute;
        right: -9px;
        content: '';
        width: 1px;
        height: 24px;
        background: #dcdee5;
      }
    }
    .uid {
      font-size: 14px;
      color: #63656e;
      margin-left: 17px;
    }
  }
  .content-wrap {
    padding: 20px 24px;
    .operate-area {
      display: flex;
      justify-content: flex-end;
      margin-bottom: 16px;
      .date-picker {
        width: 300px;
        margin-right: 8px;
      }
    }
  }
  .release_change_status {
    display: flex;
    align-items: center;
    .spinner-icon {
      font-size: 12px;
      margin-right: 10px;
    }
    .dot {
      margin-right: 10px;
      width: 8px;
      height: 8px;
      background: #f0f1f5;
      border: 1px solid #c4c6cc;
      border-radius: 50%;
      &.Success {
        background: #e5f6ea;
        border: 1px solid #3fc06d;
      }
      &.Failed {
        background: #ffe6e6;
        border: 1px solid #ea3636;
      }
    }
    .info-icon {
      font-size: 14px;
      margin-left: 9px;
    }
  }
  .config-version {
    display: flex;
    align-items: center;
    cursor: pointer;
    .text {
      margin-left: 4px;
    }
  }
</style>
