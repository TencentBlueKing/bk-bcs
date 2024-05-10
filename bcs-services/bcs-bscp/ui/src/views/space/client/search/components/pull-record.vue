<template>
  <bk-sideslider :is-show="show" width="1200" quick-close @closed="emits('close')">
    <template #header>
      <div class="header">
        <span class="title">{{ $t('配置拉取记录') }}</span>
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
          :placeholder="$t('当前配置版本/目标配置版本/最近一次拉取配置状态')"
          @search="loadTableData" />
      </div>
      <bk-loading :loading="loading">
        <bk-table :data="tableData" :border="['outer', 'row']" :pagination="pagination">
          <bk-table-column :label="$t('开始时间')" width="154">
            <template #default="{ row }">
              <span v-if="row.spec">
                {{ datetimeFormat(row.spec.start_time) }}
              </span>
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('结束时间')" width="154">
            <template #default="{ row }">
              <span v-if="row.spec">
                {{ datetimeFormat(row.spec.end_time) }}
              </span>
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('源版本')" width="133">
            <template #default="{ row }">
              <div
                v-if="row.spec && row.spec.original_release_id"
                class="config-version"
                @click="linkToApp(row.spec.original_release_id)">
                <Share class="icon" />
                <span class="text">{{ row.original_release_name }}</span>
              </div>
              <span v-else>--</span>
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('目标配置版本')" width="134">
            <template #default="{ row }">
              <div
                v-if="row.spec && row.spec.target_release_id"
                class="config-version"
                @click="linkToApp(row.spec.target_release_id)">
                <Share class="icon" />
                <span class="text">{{ row.target_release_name }}</span>
              </div>
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('配置拉取方式')" prop="attachment.client_mode" width="110"></bk-table-column>
          <bk-table-column :label="$t('配置拉取耗时')" width="128">
            <template #default="{ row }">
              <span v-if="row.spec">
                {{
                  row.spec.total_seconds > 1
                    ? `${Math.round(row.spec.total_seconds)}s`
                    : `${Math.round(row.spec.total_seconds * 1000)}ms`
                }}
              </span>
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('配置拉取文件数')" width="120">
            <template #default="{ row }">
              <div v-if="row.spec">{{ `${row.spec.download_file_num}/${row.spec.total_file_num}` }}</div>
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('配置拉取文件大小')" width="130">
            <template #default="{ row }">
              <div v-if="row.spec">{{ byteUnitConverse(row.spec.download_file_size) }}</div>
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('配置拉取状态')" width="110">
            <template #default="{ row }">
              <div v-if="row.spec" class="release_change_status">
                <Spinner v-if="row.spec.release_change_status === 'Skip'" class="spinner-icon" fill="#3A84FF" />
                <div v-else :class="['dot', row.spec.release_change_status]"></div>
                <span>{{ CLIENT_STATUS_MAP[row.spec.release_change_status as keyof typeof CLIENT_STATUS_MAP] }}</span>
                <InfoLine
                  v-if="row.spec.release_change_status === 'Failed'"
                  class="info-icon"
                  fill="#979BA5"
                  v-bk-tooltips="{ content: getErrorDetails(row.spec) }" />
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
  import {
    CLIENT_STATUS_MAP,
    CLIENT_ERROR_SUBCLASSES_MAP,
    CLIENT_ERROR_CATEGORY_MAP,
  } from '../../../../../constants/client';
  import dayjs from 'dayjs';
  import TableEmpty from '../../../../../components/table/table-empty.vue';
  import { useI18n } from 'vue-i18n';
  const { t } = useI18n();

  const router = useRouter();

  const props = defineProps<{
    bkBizId: string;
    appId: number;
    show: boolean;
    id: number;
    uid: string;
  }>();
  const emits = defineEmits(['close']);

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
        loadTableData();
      }
    },
  );

  const loadTableData = async () => {
    try {
      loading.value = true;
      isSearchEmpty.value = searchStr.value !== '';
      let search_value;
      if (searchStr.value === '成功') {
        search_value = 'success';
      } else if (searchStr.value === '失败') {
        search_value = 'failed';
      } else {
        search_value = searchStr.value;
      }
      const params = {
        start: pagination.value.limit * (pagination.value.current - 1),
        limit: pagination.value.limit,
        start_time: initDateTime.value[0],
        end_time: initDateTime.value[1],
        search_value,
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
    emits('close');
    router.push({ name: 'service-config', params: { spaceId: props.bkBizId, appId: props.appId, versionId } });
  };

  const handleClearSearchStr = () => {
    searchStr.value = '';
    loadTableData();
  };

  const getErrorDetails = (item: any) => {
    const { release_change_failed_reason, specific_failed_reason, failed_detail_reason } = item;
    const category = CLIENT_ERROR_CATEGORY_MAP.find((item) => item.value === release_change_failed_reason)?.name;
    const subclasses = CLIENT_ERROR_SUBCLASSES_MAP.find((item) => item.value === specific_failed_reason)?.name || '--';
    return `${t('错误类别')}: ${category}
    ${t('错误子类别')}: ${subclasses}
    ${t('错误详情')}: ${failed_detail_reason}`;
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
    .icon {
      color: #979ba5;
    }
    &:hover {
      color: #3a84ff;
      .icon {
        color: #3a84ff;
      }
    }
  }
</style>
